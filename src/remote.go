// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drive

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	"github.com/mxk/go-flowrate/flowrate"
	"github.com/odeke-em/drive/config"
	"github.com/odeke-em/statos"

	expb "github.com/odeke-em/exponential-backoff"
	drive "google.golang.org/api/drive/v2"
	"google.golang.org/api/googleapi"
)

const (
	// OAuth 2.0 OOB redirect URL for authorization.
	RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	// OAuth 2.0 full Drive scope used for authorization.
	DriveScope = "https://www.googleapis.com/auth/drive"

	// OAuth 2.0 access type for offline/refresh access.
	AccessType = "offline"

	// Google Drive webpage host
	DriveResourceHostURL = "https://googledrive.com/host/"

	// Google Drive entry point
	DriveResourceEntryURL = "https://drive.google.com"

	DriveRemoteSep = "/"
)

const (
	OptNone = 1 << iota
	OptConvert
	OptOCR
	OptUpdateViewedDate
	OptContentAsIndexableText
	OptPinned
	OptNewRevision
)

var (
	ErrPathNotExists   = nonExistantRemoteErr(fmt.Errorf("remote path doesn't exist"))
	ErrNetLookup       = netLookupFailedErr(fmt.Errorf("net lookup failed"))
	ErrClashesDetected = clashesDetectedErr(fmt.Errorf("clashes detected. Use `%s` to override this behavior or `%s` to try fixing this",
		CLIOptionIgnoreNameClashes, CLIOptionFixClashesKey))
	ErrClashFixingAborted             = clashFixingAbortedErr(fmt.Errorf("clash fixing aborted"))
	ErrGoogleAPIInvalidQueryHardCoded = invalidGoogleAPIQueryErr(fmt.Errorf("GoogleAPI: Error 400: Invalid query, invalid"))

	errNilParent = nonExistantRemoteErr(fmt.Errorf("remote parent doesn't exist"))
)

var (
	UnescapedPathSep = fmt.Sprintf("%c", os.PathSeparator)
	EscapedPathSep   = url.QueryEscape(UnescapedPathSep)
)

func errCannotMkdirAll(p string) error {
	return mkdirFailedErr(fmt.Errorf("cannot mkdirAll: `%s`", p))
}

type Remote struct {
	client       *http.Client
	service      *drive.Service
	encrypter    func(io.Reader) (io.Reader, error)
	decrypter    func(io.Reader) (io.ReadCloser, error)
	progressChan chan int
}

// NewRemoteContextFromServiceAccount returns a remote initialized
// with credentials from a Google Service Account.
// For more information about these accounts, see:
// https://developers.google.com/identity/protocols/OAuth2ServiceAccount
// https://developers.google.com/accounts/docs/application-default-credentials
//
// You'll also need to configure access to Google Drive.
func NewRemoteContextFromServiceAccount(jwtConfig *jwt.Config) (*Remote, error) {
	client := jwtConfig.Client(context.Background())
	return remoteFromClient(client)
}

func NewRemoteContext(context *config.Context) (*Remote, error) {
	client := newOAuthClient(context)
	return remoteFromClient(client)
}

func remoteFromClient(client *http.Client) (*Remote, error) {
	service, err := drive.New(client)
	if err != nil {
		return nil, err
	}

	progressChan := make(chan int)
	rem := &Remote{
		progressChan: progressChan,
		service:      service,
		client:       client,
	}
	return rem, nil
}

func hasExportLinks(f *File) bool {
	if f == nil || f.IsDir {
		return false
	}
	return len(f.ExportLinks) >= 1
}

func (r *Remote) changes(startChangeId int64) (chan *drive.Change, error) {
	req := r.service.Changes.List()
	if startChangeId >= 0 {
		req = req.StartChangeId(startChangeId)
	}

	changeChan := make(chan *drive.Change)
	go func() {
		pageToken := ""
		for {
			if pageToken != "" {
				req = req.PageToken(pageToken)
			}
			res, err := req.Do()
			if err != nil {
				break
			}
			for _, chItem := range res.Items {
				changeChan <- chItem
			}
			pageToken = res.NextPageToken
			if pageToken == "" {
				break
			}
		}
		close(changeChan)
	}()

	return changeChan, nil
}

func buildExpression(parentId string, typeMask int, inTrash bool) string {
	var exprBuilder []string

	exprBuilder = append(exprBuilder, fmt.Sprintf("'%s' in parents and trashed=%t", parentId, inTrash))

	// Folder and NonFolder are mutually exclusive.
	if (typeMask & Folder) != 0 {
		exprBuilder = append(exprBuilder, fmt.Sprintf("mimeType = '%s'", DriveFolderMimeType))
	}
	return strings.Join(exprBuilder, " and ")
}

func (r *Remote) change(changeId string) (*drive.Change, error) {
	return r.service.Changes.Get(changeId).Do()
}

func RetrieveRefreshToken(ctx context.Context, context *config.Context) (string, error) {
	config := newAuthConfig(context)

	randState := fmt.Sprintf("%s%v", time.Now(), rand.Uint32())
	url := config.AuthCodeURL(randState, oauth2.AccessTypeOffline)

	fmt.Printf("Visit this URL to get an authorization code\n%s\n", url)
	code := prompt(os.Stdin, os.Stdout, "Paste the authorization code: ")

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return token.RefreshToken, nil
}

func (r *Remote) FindBackPaths(id string) (backPaths []string, err error) {
	f, fErr := r.FindById(id)
	if fErr != nil {
		err = fErr
		return
	}

	relPath := sepJoin(DriveRemoteSep, f.Name)
	if rootLike(relPath) {
		relPath = DriveRemoteSep
	}

	if len(f.Parents) < 1 {
		backPaths = append(backPaths, relPath)
		return
	}

	for _, p := range f.Parents {
		if p == nil {
			continue
		}

		if p.IsRoot {
			backPaths = append(backPaths, sepJoin(DriveRemoteSep, relPath))
			continue
		}

		fullSubPaths, pErr := r.FindBackPaths(p.Id)
		if pErr != nil {
			continue
		}

		for _, subPath := range fullSubPaths {
			backPaths = append(backPaths, sepJoin(DriveRemoteSep, subPath, relPath))
		}
	}

	return
}

func wrapInPaginationPair(f *File, err error) *paginationPair {
	fChan := make(chan *File)
	errsChan := make(chan error)

	go func() {
		defer close(fChan)
		fChan <- f
	}()

	go func() {
		defer close(errsChan)
		errsChan <- err
	}()

	return &paginationPair{errsChan: errsChan, filesChan: fChan}
}

func (r *Remote) FindByIdM(id string) *paginationPair {
	f, err := r.FindById(id)
	return wrapInPaginationPair(f, err)
}

func (r *Remote) FindById(id string) (*File, error) {
	req := r.service.Files.Get(id)
	f, err := req.Do()
	if err != nil {
		return nil, err
	}
	return NewRemoteFile(f), nil
}

func retryableChangeOp(fn func() (interface{}, error), debug bool, retryCount int) *expb.ExponentialBacker {
	if retryCount < 0 {
		retryCount = MaxFailedRetryCount
	}

	return &expb.ExponentialBacker{
		Do:          fn,
		Debug:       debug,
		RetryCount:  uint32(retryCount),
		StatusCheck: retryableErrorCheck,
	}
}

func (r *Remote) findByPathM(p string, trashed bool) *paginationPair {
	if rootLike(p) {
		return r.FindByIdM("root")
	}

	parts := strings.Split(p, RemoteSeparator)
	finder := r.findByPathRecvM
	if trashed {
		finder = r.findByPathTrashedM
	}

	return finder("root", parts[1:])
}

func (r *Remote) findByPath(p string, trashed bool) (*File, error) {
	if rootLike(p) {
		return r.FindById("root")
	}
	parts := strings.Split(p, "/")
	finder := r.findByPathRecv
	if trashed {
		finder = r.findByPathTrashed
	}
	return finder("root", parts[1:])
}

func (r *Remote) FindByPath(p string) (*File, error) {
	return r.findByPath(p, false)
}

func (r *Remote) FindByPathM(p string) *paginationPair {
	return r.findByPathM(p, false)
}

func (r *Remote) FindByPathTrashed(p string) (*File, error) {
	return r.findByPath(p, true)
}

func (r *Remote) FindByPathTrashedM(p string) *paginationPair {
	return r.findByPathM(p, true)
}

func reqDoPage(req *drive.FilesListCall, hidden bool, promptOnPagination bool) *paginationPair {
	return _reqDoPage(req, hidden, promptOnPagination, false)
}

type paginationPair struct {
	errsChan  chan error
	filesChan chan *File
}

func _reqDoPage(req *drive.FilesListCall, hidden bool, promptOnPagination, nilOnNoMatch bool) *paginationPair {
	filesChan := make(chan *File)
	errsChan := make(chan error)

	throttle := time.Tick(1e8)

	go func() {
		defer func() {
			close(errsChan)
			close(filesChan)
		}()

		pageToken := ""
		for pageIterCount := uint64(0); ; pageIterCount++ {
			if pageToken != "" {
				req = req.PageToken(pageToken)
			}
			results, err := req.Do()
			if err != nil {
				errsChan <- err
				break
			}

			iterCount := uint64(0)
			for _, f := range results.Items {
				if isHidden(f.Title, hidden) { // ignore hidden files
					continue
				}
				iterCount += 1
				filesChan <- NewRemoteFile(f)
			}

			pageToken = results.NextPageToken
			if pageToken == "" {
				if nilOnNoMatch && len(results.Items) < 1 && pageIterCount < 1 {
					// Item absolutely doesn't exist
					filesChan <- nil
				}
				break
			}

			<-throttle

			if iterCount < 1 {
				continue
			}

			if promptOnPagination && !nextPage() {
				filesChan <- nil
				break
			}
		}
	}()

	return &paginationPair{filesChan: filesChan, errsChan: errsChan}
}

func (r *Remote) findByParentIdRaw(parentId string, trashed, hidden bool) *paginationPair {
	req := r.service.Files.List()
	req.Q(fmt.Sprintf("%s in parents and trashed=%v", customQuote(parentId), trashed))
	return reqDoPage(req, hidden, false)
}

func (r *Remote) FindByParentId(parentId string, hidden bool) *paginationPair {
	return r.findByParentIdRaw(parentId, false, hidden)
}

func (r *Remote) FindByParentIdTrashed(parentId string, hidden bool) *paginationPair {
	return r.findByParentIdRaw(parentId, true, hidden)
}

func (r *Remote) EmptyTrash() error {
	return r.service.Files.EmptyTrash().Do()
}

func (r *Remote) Trash(id string) error {
	_, err := r.service.Files.Trash(id).Do()
	return err
}

func (r *Remote) Untrash(id string) error {
	_, err := r.service.Files.Untrash(id).Do()
	return err
}

func (r *Remote) Delete(id string) error {
	return r.service.Files.Delete(id).Do()
}

func (r *Remote) idForEmail(email string) (string, error) {
	perm, err := r.service.Permissions.GetIdForEmail(email).Do()
	if err != nil {
		return "", err
	}
	return perm.Id, nil
}

func (r *Remote) listPermissions(id string) ([]*drive.Permission, error) {
	res, err := r.service.Permissions.List(id).Do()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (r *Remote) insertPermissions(permInfo *permission) (*drive.Permission, error) {
	perm := &drive.Permission{
		Role:     permInfo.role.String(),
		Type:     permInfo.accountType.String(),
		WithLink: permInfo.withLink,
	}

	if permInfo.value != "" {
		perm.Value = permInfo.value
	}

	req := r.service.Permissions.Insert(permInfo.fileId, perm)

	if permInfo.message != "" {
		req = req.EmailMessage(permInfo.message)
	}
	req = req.SendNotificationEmails(permInfo.notify)
	return req.Do()
}

func (r *Remote) revokePermissions(p *permission) (err error) {
	foundPermissionsChan, fErr := r.findPermissions(p)
	if fErr != nil {
		return fErr
	}

	successes := 0
	for perm := range foundPermissionsChan {
		if perm == nil {
			continue
		}

		req := r.service.Permissions.Delete(p.fileId, perm.Id)
		if delErr := req.Do(); delErr != nil {
			err = reComposeError(err, fmt.Sprintf("err: %v fileId: %s permissionId %s", delErr, p.fileId, perm.Id))
		} else {
			successes += 1
		}
	}

	if err != nil {
		return err
	}

	if successes < 1 {
		err = noMatchesFoundErr(fmt.Errorf("no matches found!"))
	}

	return err
}

func stringifyPermissionForMatch(p *permission) string {
	// As of "Fri Nov 20 19:06:18 MST 2015", Google Drive v2 API doesn't support a PermissionsList.Q(query)
	// call hence the hack around will be to compare all the permissions returned to those being queried for
	queries := []string{}
	if role := p.role.String(); !unknownRole(role) {
		queries = append(queries, role)
	}

	if accountType := p.accountType.String(); !unknownAccountType(accountType) {
		queries = append(queries, accountType)
	}

	if p.value != "" {
		queries = append(queries, p.value)
	}

	return sepJoin("", preprocessBeforePermissionMatch(queries)...)
}

func preprocessBeforePermissionMatch(args []string) (preprocessed []string) {
	for _, arg := range args {
		preprocessed = append(preprocessed, strings.ToLower(strings.TrimSpace(arg)))
	}

	return preprocessed
}

func stringifyDrivePermissionForMatch(p *drive.Permission) string {
	// As of "Fri Nov 20 19:06:18 MST 2015", Google Drive v2 API doesn't support a PermissionsList.Q(query)
	// call hence the hack around will be  to compare all the permissions returned to those being queried for
	params := []string{p.Role, p.Type, p.EmailAddress}

	repr := []string{}
	for _, param := range params {
		repr = append(repr, strings.ToLower(param))
	}

	return sepJoin("", preprocessBeforePermissionMatch(repr)...)
}

func (r *Remote) findPermissions(pquery *permission) (permChan chan *drive.Permission, err error) {
	permChan = make(chan *drive.Permission)

	go func() {
		defer close(permChan)

		req := r.service.Permissions.List(pquery.fileId)

		results, err := req.Do()
		if err != nil {
			fmt.Println(err)
			return
		}
		if results == nil {
			return
		}

		requiredSignature := stringifyPermissionForMatch(pquery)
		// fmt.Println("requiredSignature", requiredSignature)
		for _, perm := range results.Items {
			if perm == nil {
				continue
			}

			if stringifyDrivePermissionForMatch(perm) == requiredSignature {
				// fmt.Println("perm", perm)
				permChan <- perm
			}
		}
	}()

	return
}

func (r *Remote) deletePermissions(id string, accountType AccountType) error {
	return r.service.Permissions.Delete(id, accountType.String()).Do()
}

func (r *Remote) Unpublish(id string) error {
	return r.deletePermissions(id, Anyone)
}

func (r *Remote) Publish(id string) (string, error) {
	_, err := r.insertPermissions(&permission{
		fileId:      id,
		value:       "",
		role:        Reader,
		accountType: Anyone,
	})
	if err != nil {
		return "", err
	}
	return DriveResourceHostURL + id, nil
}

func urlToPath(p string, fsBound bool) string {
	if fsBound {
		return strings.Replace(p, UnescapedPathSep, EscapedPathSep, -1)
	}
	return strings.Replace(p, EscapedPathSep, UnescapedPathSep, -1)
}

func (r *Remote) Download(id string, exportURL string) (io.ReadCloser, error) {
	var url string
	var body io.ReadCloser

	var resp *http.Response
	var err error

	if len(exportURL) < 1 {
		resp, err = r.service.Files.Get(id).Download()
	} else {
		resp, err = r.client.Get(exportURL)
	}

	if err == nil {
		if resp == nil {
			err = illogicalStateErr(fmt.Errorf("bug on: download for url \"%s\". resp and err are both nil", url))
		} else if httpOk(resp.StatusCode) { // TODO: Handle other statusCodes e.g redirects?
			body = resp.Body
		} else {
			err = downloadFailedErr(fmt.Errorf("download: failed for url \"%s\". StatusCode: %v", url, resp.StatusCode))
		}
	}

	if r.decrypter != nil && body != nil {
		decR, err := r.decrypter(body)
		_ = body.Close()
		if err != nil {
			return nil, err
		}
		body = decR
	}

	return body, err
}

func (r *Remote) Touch(id string) (*File, error) {
	f, err := r.service.Files.Touch(id).Do()
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, ErrPathNotExists
	}
	return NewRemoteFile(f), err
}

// SetModTime is an explicit command to just set the modification
// time of a remote file. It serves the purpose of Touch but with
// a custom time instead of the time on the remote server.
// See Issue https://github.com/odeke-em/drive/issues/726.
func (r *Remote) SetModTime(fileId string, modTime time.Time) (*File, error) {
	repr := &drive.File{}

	// Ensure that the ModifiedDate is retrieved from local
	repr.ModifiedDate = toUTCString(modTime)

	req := r.service.Files.Update(fileId, repr)

	// We always want it to match up with the local time
	req.SetModifiedDate(true)

	retrieved, err := req.Do()
	if err != nil {
		return nil, err
	}

	return NewRemoteFile(retrieved), nil
}

func toUTCString(t time.Time) string {
	utc := t.UTC().Round(time.Second)
	// Ugly but straight forward formatting as time.Parse is such a prima donna
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.000Z",
		utc.Year(), utc.Month(), utc.Day(),
		utc.Hour(), utc.Minute(), utc.Second())
}

func convert(mask int) bool {
	return (mask & OptConvert) != 0
}

func ocr(mask int) bool {
	return (mask & OptOCR) != 0
}

func pin(mask int) bool {
	return (mask & OptPinned) != 0
}

func indexContent(mask int) bool {
	return (mask & OptContentAsIndexableText) != 0
}

type upsertOpt struct {
	debug           bool
	parentId        string
	fsAbsPath       string
	relToRootPath   string
	src             *File
	dest            *File
	mask            int
	ignoreChecksum  bool
	mimeKey         string
	nonStatable     bool
	retryCount      int
	uploadChunkSize int
	uploadRateLimit int
}

func togglePropertiesInsertCall(req *drive.FilesInsertCall, mask int) *drive.FilesInsertCall {
	// TODO: if ocr toggled respect the quota limits if ocr is enabled.
	if ocr(mask) {
		req = req.Ocr(true)
	}
	if convert(mask) {
		req = req.Convert(true)
	}
	if pin(mask) {
		req = req.Pinned(true)
	}
	if indexContent(mask) {
		req = req.UseContentAsIndexableText(true)
	}
	return req
}

func togglePropertiesUpdateCall(req *drive.FilesUpdateCall, mask int) *drive.FilesUpdateCall {
	// TODO: if ocr toggled respect the quota limits if ocr is enabled.
	if ocr(mask) {
		req = req.Ocr(true)
	}
	if convert(mask) {
		req = req.Convert(true)
	}
	if pin(mask) {
		req = req.Pinned(true)
	}
	if indexContent(mask) {
		req = req.UseContentAsIndexableText(true)
	}
	return req
}

// shouldUploadBody tells whether a body of content should
// be uploaded. It rejects local directories.
// It returns true if the content is nonStatable e.g piped input
// or if the destination on the cloud is nil
// and also if there are checksum differences.
// For other changes such as modTime only varying, we can
// just change the modTime on the cloud as an operation of its own.
func (args *upsertOpt) shouldUploadBody() bool {
	if args.src.IsDir {
		return false
	}
	if args.dest == nil || args.nonStatable {
		return true
	}
	mask := fileDifferences(args.src, args.dest, args.ignoreChecksum)
	return checksumDiffers(mask)
}

func (r *Remote) upsertByComparison(body io.Reader, args *upsertOpt) (f *File, mediaInserted bool, err error) {
	uploaded := &drive.File{
		// Must ensure that the path is prepared for a URL upload
		Title:   urlToPath(args.src.Name, false),
		Parents: []*drive.ParentReference{&drive.ParentReference{Id: args.parentId}},
	}

	if args.src.IsDir {
		uploaded.MimeType = DriveFolderMimeType
	}

	if r.encrypter != nil && body != nil {
		encR, encErr := r.encrypter(body)
		if encErr != nil {
			err = encErr
			return
		}
		body = encR
	}

	// throttled reader: implement upload bandwidth limit
	// uploadRateLimit is in KiB/s
	reader := flowrate.NewReader(body, int64(args.uploadRateLimit*1024))

	if args.src.MimeType != "" {
		uploaded.MimeType = args.src.MimeType
	}

	if args.mimeKey != "" {
		uploaded.MimeType = guessMimeType(args.mimeKey)
	}

	// Ensure that the ModifiedDate is retrieved from local
	uploaded.ModifiedDate = toUTCString(args.src.ModTime)

	var mediaOptions []googleapi.MediaOption
	if args.uploadChunkSize > 0 {
		mediaOptions = append(mediaOptions, googleapi.ChunkSize(args.uploadChunkSize))
	}

	if args.src.Id == "" {
		req := r.service.Files.Insert(uploaded)

		if !args.src.IsDir && body != nil {
			req = req.Media(reader, mediaOptions...)
			mediaInserted = true
		}

		// Toggle the respective properties
		req = togglePropertiesInsertCall(req, args.mask)

		if uploaded, err = req.Do(); err != nil {
			return
		}

		f = NewRemoteFile(uploaded)
		return
	}

	// update the existing
	req := r.service.Files.Update(args.src.Id, uploaded)

	// We always want it to match up with the local time
	req.SetModifiedDate(true)

	if args.shouldUploadBody() {
		req = req.Media(reader, mediaOptions...)
		mediaInserted = true
	}

	// Next toggle the appropriate properties
	req = togglePropertiesUpdateCall(req, args.mask)

	if uploaded, err = req.Do(); err != nil {
		return
	}
	f = NewRemoteFile(uploaded)
	return
}

func (r *Remote) byFileIdUpdater(fileId string, f *drive.File) (*File, error) {
	req := r.service.Files.Update(fileId, f)
	uploaded, err := req.Do()
	if err != nil {
		return nil, err
	}

	return NewRemoteFile(uploaded), nil
}

func (r *Remote) rename(fileId, newTitle string) (*File, error) {
	f := &drive.File{
		Title: newTitle,
	}

	return r.byFileIdUpdater(fileId, f)
}

func (r *Remote) updateDescription(fileId, newDescription string) (*File, error) {
	f := &drive.File{
		Description: newDescription,
	}

	return r.byFileIdUpdater(fileId, f)
}

func (r *Remote) updateStarred(fileId string, star bool) (*File, error) {
	f := &drive.File{
		Labels: &drive.FileLabels{
			Starred: star,
			// Since "Starred" is a non-pointer value, we'll need to
			// unconditionally send it with API-requests using "ForceSendFields"
			ForceSendFields: []string{"Starred"},
		},
	}

	return r.byFileIdUpdater(fileId, f)
}

func (r *Remote) removeParent(fileId, parentId string) error {
	return r.service.Parents.Delete(fileId, parentId).Do()
}

func (r *Remote) insertParent(fileId, parentId string) error {
	parent := &drive.ParentReference{Id: parentId}
	_, err := r.service.Parents.Insert(fileId, parent).Do()
	return err
}

func (r *Remote) copy(newName, parentId string, srcFile *File) (*File, error) {
	f := &drive.File{
		Title:        urlToPath(newName, false),
		ModifiedDate: toUTCString(srcFile.ModTime),
	}
	if parentId != "" {
		f.Parents = []*drive.ParentReference{&drive.ParentReference{Id: parentId}}
	}
	copied, err := r.service.Files.Copy(srcFile.Id, f).Do()
	if err != nil {
		return nil, err
	}
	return NewRemoteFile(copied), nil
}

func (r *Remote) UpsertByComparison(args *upsertOpt) (f *File, err error) {
	/*
	   // TODO: (@odeke-em) decide:
	   //   + if to reject FIFO
	   //   + perform an assertion for fileStated.IsDir() == args.src.IsDir
	*/
	if args.src == nil {
		err = illogicalStateErr(fmt.Errorf("bug on: src cannot be nil"))
		return
	}

	var body io.Reader
	var cleanUp func() error

	if !args.src.IsDir {
		// In relation to issue #612, since we are not only resolving
		// relative to the current working directory, we should try reading
		// first from the source's original local fsAbsPath aka `BlobAt`
		// because the resolved path might be different from the original path.
		fsAbsPath := args.src.BlobAt
		if fsAbsPath == "" {
			fsAbsPath = args.fsAbsPath
		}

		if args.shouldUploadBody() {
			file, err := os.Open(fsAbsPath)
			if err != nil {
				return nil, err
			}

			// We need to make sure that we close all open handles.
			// See Issue https://github.com/odeke-em/drive/issues/711.
			cleanUp = file.Close
			body = file
		}
	}

	bd := statos.NewReader(body)

	go func() {
		commChan := bd.ProgressChan()
		for n := range commChan {
			r.progressChan <- n
		}
	}()

	resultLoad := make(chan *tuple)

	go func() {
		if cleanUp != nil {
			defer cleanUp()
		}

		emitter := func() (interface{}, error) {
			f, mediaInserted, err := r.upsertByComparison(bd, args)
			return &tuple{first: f, second: mediaInserted, last: err}, err
		}

		retrier := retryableChangeOp(emitter, args.debug, args.retryCount)

		res, err := expb.ExponentialBackOffSync(retrier)
		resultLoad <- &tuple{first: res, last: err}
	}()

	result := <-resultLoad
	if result == nil {
		return f, err
	}

	tup, tupOk := result.first.(*tuple)
	if tupOk {
		ff, fOk := tup.first.(*File)
		if fOk {
			f = ff
		}

		mediaInserted, mediaOk := tup.second.(bool)

		// Case in which for example just Chtime-ing
		if mediaOk && !mediaInserted && f != nil {
			chunks := chunkInt64(f.Size)
			for n := range chunks {
				r.progressChan <- n
			}
		}

		errV, errCastOk := tup.last.(error)
		if errCastOk {
			err = errV
		}
	}

	return f, err
}

func (r *Remote) findShared(p []string) *paginationPair {
	req := r.service.Files.List()
	expr := "sharedWithMe=true"
	if len(p) >= 1 {
		expr = fmt.Sprintf("title = '%s' and %s", p[0], expr)
	}
	req = req.Q(expr)

	return reqDoPage(req, false, false)
}

func (r *Remote) FindByPathShared(p string) *paginationPair {
	if p == "/" || p == "root" {
		return r.findShared([]string{})
	}
	parts := strings.Split(p, "/") // TODO: use path.Split instead
	nonEmpty := func(strList []string) []string {
		var nEmpty []string
		for _, p := range strList {
			if len(p) >= 1 {
				nEmpty = append(nEmpty, p)
			}
		}
		return nEmpty
	}(parts)
	return r.findShared(nonEmpty)
}

func (r *Remote) FindStarred(trashed, hidden bool) *paginationPair {
	req := r.service.Files.List()
	expr := fmt.Sprintf("(starred=true) and (trashed=%v)", trashed)
	req.Q(expr)
	return reqDoPage(req, hidden, false)
}

func (r *Remote) FindMatches(mq *matchQuery) *paginationPair {
	parent, err := r.FindByPath(mq.dirPath)
	if err != nil || parent == nil {
		if parent == nil && err == nil {
			err = errNilParent
		}
		return wrapInPaginationPair(parent, err)
	}

	req := r.service.Files.List()

	parQuery := fmt.Sprintf("(%s in parents)", customQuote(parent.Id))
	expr := sepJoinNonEmpty(" and ", parQuery, mq.Stringer())

	req.Q(expr)
	return reqDoPage(req, true, false)
}

func (r *Remote) findChildren(parentId string, trashed bool) *paginationPair {
	req := r.service.Files.List()
	req.Q(fmt.Sprintf("%s in parents and trashed=%v", customQuote(parentId), trashed))
	return reqDoPage(req, true, false)
}

func (r *Remote) About() (*drive.About, error) {
	return r.service.About.Get().Do()
}

func (r *Remote) findByPathRecvRawM(parentId string, p []string, trashed bool) *paginationPair {
	chanOChan := make(chan *paginationPair)

	resolvedFilesChan := make(chan *File)
	resolvedErrsChan := make(chan error)

	go func() {
		defer close(chanOChan)

		if len(p) < 1 {
			return
		}

		first, rest := p[0], p[1:]
		// find the file or directory under parentId and titled with p[0]
		req := r.service.Files.List()
		// TODO: use field selectors
		var expr string
		head := urlToPath(first, false)
		if trashed {
			expr = fmt.Sprintf("title = %s and trashed=true", customQuote(head))
		} else {
			expr = fmt.Sprintf("%s in parents and title = %s and trashed=false",
				customQuote(parentId), customQuote(head))
		}

		req.Q(expr)
		pager := _reqDoPage(req, true, false, true)

		if len(rest) < 1 {
			chanOChan <- pager
			return
		}

		resultsChan := pager.filesChan
		errsChan := pager.errsChan

		working := true
		for working {
			select {
			case err := <-errsChan:
				if err != nil {
					chanOChan <- wrapInPaginationPair(nil, err)
				}
			case f, stillHasContent := <-resultsChan:
				if !stillHasContent {
					working = false
					break
				}

				if f != nil {
					chanOChan <- r.findByPathRecvRawM(f.Id, rest, trashed)
				} else {
					// Ensure that we properly send
					// back nil even if files were not found.
					// See https://github.com/odeke-em/drive/issues/933.
					chanOChan <- wrapInPaginationPair(nil, nil)
				}
			}
		}
	}()

	go func() {
		defer func() {
			close(resolvedFilesChan)
			close(resolvedErrsChan)
		}()

		for curPagePair := range chanOChan {
			if curPagePair == nil {
				continue
			}

			errsChan := curPagePair.errsChan
			filesChan := curPagePair.filesChan

			working := true
			for working {
				select {
				case err := <-errsChan:
					if err != nil {
						resolvedErrsChan <- err
					}
				case f, stillHasContent := <-filesChan:
					if !stillHasContent {
						working = false
						break
					}
					resolvedFilesChan <- f
				}
			}
		}
	}()

	return &paginationPair{errsChan: resolvedErrsChan, filesChan: resolvedFilesChan}
}

func (r *Remote) findByPathRecvRaw(parentId string, p []string, trashed bool) (*File, error) {
	// find the file or directory under parentId and titled with p[0]
	req := r.service.Files.List()
	// TODO: use field selectors
	var expr string
	head := urlToPath(p[0], false)
	if trashed {
		expr = fmt.Sprintf("title = %s and trashed=true", customQuote(head))
	} else {
		expr = fmt.Sprintf("%s in parents and title = %s and trashed=false",
			customQuote(parentId), customQuote(head))
	}
	req.Q(expr)

	// We only need the head file since we expect only one File to be created
	req.MaxResults(1)

	files, err := req.Do()

	if err != nil {
		if err.Error() == ErrGoogleAPIInvalidQueryHardCoded.Error() { // Send the user back the query information
			err = invalidGoogleAPIQueryErr(fmt.Errorf("err: %v query: `%s`", err, expr))
		}
		return nil, err
	}

	if files == nil || len(files.Items) < 1 {
		return nil, ErrPathNotExists
	}

	first := files.Items[0]
	if len(p) == 1 {
		return NewRemoteFile(first), nil
	}
	return r.findByPathRecvRaw(first.Id, p[1:], trashed)
}

func (r *Remote) findByPathRecv(parentId string, p []string) (*File, error) {
	return r.findByPathRecvRaw(parentId, p, false)
}

func (r *Remote) findByPathRecvM(parentId string, p []string) *paginationPair {
	return r.findByPathRecvRawM(parentId, p, false)
}

func (r *Remote) findByPathTrashedM(parentId string, p []string) *paginationPair {
	return r.findByPathRecvRawM(parentId, p, true)
}

func (r *Remote) findByPathTrashed(parentId string, p []string) (*File, error) {
	return r.findByPathRecvRaw(parentId, p, true)
}

func newAuthConfig(context *config.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     context.ClientId,
		ClientSecret: context.ClientSecret,
		RedirectURL:  RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{DriveScope},
	}
}

func newOAuthClient(configContext *config.Context) *http.Client {
	config := newAuthConfig(configContext)

	token := oauth2.Token{
		RefreshToken: configContext.RefreshToken,
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	return config.Client(context.Background(), &token)
}

type WatchRequest struct {
	FileId  string
	Request *drive.Channel

	Cancel chan struct{}
}

type WatchResponse struct {
	Channel *drive.Channel

	Err    error
	Cancel chan struct{}
}

func (rem *Remote) watchForChanges(wreq *WatchRequest) (chan *WatchResponse, error) {
	watchChan := make(chan *WatchResponse)

	go func() {
		defer close(watchChan)

		if wreq.Request.Kind == "" {
			wreq.Request.Kind = "api#channel"
		}
		if wreq.Request.Type == "" {
			wreq.Request.Type = "web_hook"
		}

		watchCall := rem.service.Files.Watch(wreq.FileId, wreq.Request)
		cancel := wreq.Cancel

		working := true
		var pageToken string
		throttle := time.NewTicker(1e8)
		for working {
			select {
			case <-cancel:
				working = false
				break
			default:
				wchan, err := watchCall.Do()
				fmt.Printf("wchan: %#v err: %v\n", wchan, err)
				watchChan <- &WatchResponse{Err: err, Channel: wchan}

				log.Printf("token: %s\n", pageToken)
				if wchan != nil {
					pageToken = wchan.Token
				}

				if pageToken == "" {
					working = false
					break
				}
				<-throttle.C
			}
		}
	}()

	return watchChan, nil
}
