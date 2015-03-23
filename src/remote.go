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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"github.com/odeke-em/drive/config"
	drive "github.com/odeke-em/google-api-go-client/drive/v2"
)

const (
	// Google OAuth 2.0 service URLs
	GoogleOAuth2AuthURL  = "https://accounts.google.com/o/oauth2/auth"
	GoogleOAuth2TokenURL = "https://accounts.google.com/o/oauth2/token"

	// OAuth 2.0 OOB redirect URL for authorization.
	RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	// OAuth 2.0 full Drive scope used for authorization.
	DriveScope = "https://www.googleapis.com/auth/drive"

	// OAuth 2.0 access type for offline/refresh access.
	AccessType = "offline"

	// Google Drive webpage host
	DriveResourceHostURL = "https://googledrive.com/host/"
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
	ErrPathNotExists = errors.New("remote path doesn't exist")
)

var (
	UnescapedPathSep = fmt.Sprintf("%c", os.PathSeparator)
	EscapedPathSep   = url.QueryEscape(UnescapedPathSep)
)

var regExtStrMap = map[string]string{
	"csv":   "text/csv",
	"html?": "text/html",
	"te?xt": "text/plain",

	"gif":   "image/gif",
	"png":   "image/png",
	"svg":   "image/svg+xml",
	"jpe?g": "image/jpeg",

	"odt": "application/vnd.oasis.opendocument.text",
	"rtf": "application/rtf",
	"pdf": "application/pdf",

	"docx?": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"pptx?": "application/vnd.openxmlformats-officedocument.wordprocessingml.presentation",
	"xlsx?": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

var regExtMap = func() map[*regexp.Regexp]string {
	regMap := make(map[*regexp.Regexp]string)
	for regStr, mimeType := range regExtStrMap {
		regExComp, err := regexp.Compile(regStr)
		if err == nil {
			regMap[regExComp] = mimeType
		}
	}
	return regMap
}()

func mimeTypeFromExt(ext string) string {
	bExt := []byte(ext)
	for regEx, mimeType := range regExtMap {
		if regEx != nil && regEx.Match(bExt) {
			return mimeType
		}
	}
	return ""
}

type Remote struct {
	transport *oauth.Transport
	service   *drive.Service
}

func NewRemoteContext(context *config.Context) *Remote {
	transport := newTransport(context)
	service, _ := drive.New(transport.Client())
	return &Remote{service: service, transport: transport}
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

	if inTrash || (typeMask&InTrash) != 0 {
		exprBuilder = append(exprBuilder, "trashed=true")
	} else {
		exprBuilder = append(exprBuilder, fmt.Sprintf("'%s' in parents", parentId), "trashed=false")
	}

	// Folder and NonFolder are mutually exclusive.
	if (typeMask & Folder) != 0 {
		exprBuilder = append(exprBuilder, fmt.Sprintf("mimeType = '%s'", DriveFolderMimeType))
	}
	return strings.Join(exprBuilder, " and ")
}

func (r *Remote) change(changeId string) (*drive.Change, error) {
	return r.service.Changes.Get(changeId).Do()
}

func RetrieveRefreshToken(context *config.Context) (string, error) {
	transport := newTransport(context)
	url := transport.Config.AuthCodeURL("")
	fmt.Println("Visit this URL to get an authorization code")
	fmt.Println(url)
	fmt.Print("Paste the authorization code: ")
	var code string
	fmt.Scanln(&code)
	token, err := transport.Exchange(code)
	if err != nil {
		return "", err
	}
	return token.RefreshToken, nil
}

func (r *Remote) FindById(id string) (file *File, err error) {
	req := r.service.Files.Get(id)
	var f *drive.File
	if f, err = req.Do(); err != nil {
		return
	}
	return NewRemoteFile(f), nil
}

func rootLike(p string) bool {
	return p == "/" || p == "" || p == "root"
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

func (r *Remote) FindByPath(p string) (file *File, err error) {
	return r.findByPath(p, false)
}

func (r *Remote) FindByPathTrashed(p string) (file *File, err error) {
	return r.findByPath(p, true)
}

func reqDoPage(req *drive.FilesListCall, hidden bool, promptOnPagination bool) chan *File {
	fileChan := make(chan *File)
	go func() {
		pageToken := ""
		for {
			if pageToken != "" {
				req = req.PageToken(pageToken)
			}
			results, err := req.Do()
			if err != nil {
				fmt.Println(err)
				break
			}
			for _, f := range results.Items {
				if isHidden(f.Title, hidden) { // ignore hidden files
					continue
				}
				fileChan <- NewRemoteFile(f)
			}
			pageToken = results.NextPageToken
			if pageToken == "" {
				break
			}

			if promptOnPagination && !nextPage() {
				fileChan <- nil
				break
			}
		}
		close(fileChan)
	}()
	return fileChan
}

func (r *Remote) findByParentIdRaw(parentId string, trashed, hidden bool) (fileChan chan *File) {
	req := r.service.Files.List()
	req.Q(fmt.Sprintf("%s in parents and trashed=%v", strconv.Quote(parentId), trashed))
	return reqDoPage(req, hidden, false)
}

func (r *Remote) FindByParentId(parentId string, hidden bool) chan *File {
	return r.findByParentIdRaw(parentId, false, hidden)
}

func (r *Remote) FindByParentIdTrashed(parentId string, hidden bool) chan *File {
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
		Role: permInfo.role.String(),
		Type: permInfo.accountType.String(),
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
	if len(exportURL) < 1 {
		url = DriveResourceHostURL + id
	} else {
		url = exportURL
	}
	resp, err := r.transport.Client().Get(url)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.Body, err
	}
	return resp.Body, nil
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

func toUTCString(t time.Time) string {
	utc := t.UTC().Round(time.Second)
	// Ugly but straight forward formatting as time.Parse is such a prima donna
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%0d.000Z",
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
	parentId       string
	fsAbsPath      string
	src            *File
	dest           *File
	mask           int
	ignoreChecksum bool
	nonStatable    bool
}

func (r *Remote) upsertByComparison(body io.Reader, args *upsertOpt) (f *File, err error) {
	uploaded := &drive.File{
		// Must ensure that the path is prepared for a URL upload
		Title:   urlToPath(args.src.Name, false),
		Parents: []*drive.ParentReference{&drive.ParentReference{Id: args.parentId}},
	}
	if args.src.IsDir {
		uploaded.MimeType = DriveFolderMimeType
	}

	// Ensure that the ModifiedDate is retrieved from local
	uploaded.ModifiedDate = toUTCString(args.src.ModTime)

	if args.src.Id == "" {
		req := r.service.Files.Insert(uploaded)
		if !args.src.IsDir && body != nil {
			req = req.Media(body)
		}
		if uploaded, err = req.Do(); err != nil {
			return
		}
		return NewRemoteFile(uploaded), nil
	}

	// update the existing
	req := r.service.Files.Update(args.src.Id, uploaded)

	// We always want it to match up with the local time
	req.SetModifiedDate(true)

	// Next set all the desired attributes
	// TODO: if ocr toggled respect the quota limits if ocr is enabled.
	if ocr(args.mask) {
		req.Ocr(true)
	}
	if convert(args.mask) {
		req.Convert(true)
	}
	if pin(args.mask) {
		req.Pinned(true)
	}
	if indexContent(args.mask) {
		req.UseContentAsIndexableText(true)
	}

	if !args.src.IsDir {
		if args.dest == nil || args.nonStatable {
			req = req.Media(body)
		} else if mask := fileDifferences(args.src, args.dest, args.ignoreChecksum); checksumDiffers(mask) {
			req = req.Media(body)
		}
	}
	if uploaded, err = req.Do(); err != nil {
		return
	}
	return NewRemoteFile(uploaded), nil
}

func (r *Remote) rename(fileId, newTitle string) (*File, error) {
	f := &drive.File{
		Title: newTitle,
	}

	req := r.service.Files.Update(fileId, f)
	uploaded, err := req.Do()
	if err != nil {
		return nil, err
	}

	return NewRemoteFile(uploaded), nil
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
	var body io.Reader
	body, err = os.Open(args.fsAbsPath)
	if args.src == nil {
		err = fmt.Errorf("bug on: src cannot be nil")
		return
	}
	if err != nil && !args.src.IsDir {
		return
	}
	return r.upsertByComparison(body, args)
}

func (r *Remote) findShared(p []string) (chan *File, error) {
	req := r.service.Files.List()
	expr := "sharedWithMe=true"
	if len(p) >= 1 {
		expr = fmt.Sprintf("title = '%s' and %s", p[0], expr)
	}
	req = req.Q(expr)

	return reqDoPage(req, false, false), nil
}

func (r *Remote) FindByPathShared(p string) (chan *File, error) {
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

func (r *Remote) FindMatches(dirPath string, keywords []string, inTrash bool) (chan *File, error) {
	parent, err := r.FindByPath(dirPath)
	filesChan := make(chan *File)
	if err != nil || parent == nil {
		close(filesChan)
		return filesChan, err
	}

	req := r.service.Files.List()
	keySearches := make([]string, len(keywords))

	for i, key := range keywords {
		quoted := strconv.Quote(key)
		keySearches[i] = fmt.Sprintf("(title contains %s and trashed=%v)", quoted, inTrash)
	}

	expr := strings.Join(keySearches, " or ")
	// And always make sure that we are searching from this parent
	expr = fmt.Sprintf("%s in parents and (%s)", strconv.Quote(parent.Id), expr)
	req.Q(expr)
	return reqDoPage(req, true, false), nil
}

func (r *Remote) findChildren(parentId string, trashed bool) chan *File {
	req := r.service.Files.List()
	req.Q(fmt.Sprintf("%s in parents and trashed=%v", strconv.Quote(parentId), trashed))
	return reqDoPage(req, true, false)
}

func (r *Remote) About() (about *drive.About, err error) {
	return r.service.About.Get().Do()
}

func (r *Remote) findByPathRecvRaw(parentId string, p []string, trashed bool) (file *File, err error) {
	// find the file or directory under parentId and titled with p[0]
	req := r.service.Files.List()
	// TODO: use field selectors
	var expr string
	head := urlToPath(p[0], false)
	quote := strconv.Quote
	if trashed {
		expr = fmt.Sprintf("title = %s and trashed=true", quote(head))
	} else {
		expr = fmt.Sprintf("%s in parents and title = %s and trashed=false",
			quote(parentId), quote(head))
	}
	req.Q(expr)

	// We only need the head file since we expect only one File to be created
	req.MaxResults(1)
	files, err := req.Do()
	if err != nil || len(files.Items) < 1 {
		// TODO: make sure only 404s are handled here
		return nil, ErrPathNotExists
	}

	first := files.Items[0]
	if len(p) == 1 {
		return NewRemoteFile(first), nil
	}
	return r.findByPathRecvRaw(first.Id, p[1:], trashed)
}

func (r *Remote) findByPathRecv(parentId string, p []string) (file *File, err error) {
	return r.findByPathRecvRaw(parentId, p, false)
}

func (r *Remote) findByPathTrashed(parentId string, p []string) (file *File, err error) {
	return r.findByPathRecvRaw(parentId, p, true)
}

func newAuthConfig(context *config.Context) *oauth.Config {
	return &oauth.Config{
		ClientId:     context.ClientId,
		ClientSecret: context.ClientSecret,
		AuthURL:      GoogleOAuth2AuthURL,
		TokenURL:     GoogleOAuth2TokenURL,
		RedirectURL:  RedirectURL,
		AccessType:   AccessType,
		Scope:        DriveScope,
	}
}

func newTransport(context *config.Context) *oauth.Transport {
	return &oauth.Transport{
		Config:    newAuthConfig(context),
		Transport: http.DefaultTransport,
		Token: &oauth.Token{
			RefreshToken: context.RefreshToken,
			Expiry:       time.Now(),
		},
	}
}
