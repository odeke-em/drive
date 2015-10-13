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
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/odeke-em/drive/config"
	drive "google.golang.org/api/drive/v2"
)

type Operation int

const (
	OpNone Operation = 1 << iota
	OpAdd
	OpDelete
	OpIndexAddition
	OpMod
	OpModConflict
)

type CrudValue int

const (
	None   CrudValue = 0
	Create CrudValue = 1 << iota
	Read
	Update
	Delete
)

var (
	AllCrudOperations CrudValue = Create | Read | Update | Delete
)

const (
	DifferNone    = 0
	DifferDirType = 1 << iota
	DifferMd5Checksum
	DifferModTime
	DifferSize
)

const (
	DriveFolderMimeType = "application/vnd.google-apps.folder"
)

// Arbitrary value. TODO: Get better definition of BigFileSize.
var BigFileSize = int64(1024 * 1024 * 400)

var opPrecedence = map[Operation]int{
	OpNone:        0,
	OpDelete:      1,
	OpMod:         2,
	OpModConflict: 3,
	OpAdd:         4,
}

type File struct {
	// AlternateLink opens the file in a relevant Google editor or viewer
	AlternateLink string
	BlobAt        string
	// Copyable decides if the user has allowed for the file to be copied
	Copyable           bool
	ExportLinks        map[string]string
	Id                 string
	IsDir              bool
	Md5Checksum        string
	MimeType           string
	ModTime            time.Time
	LastViewedByMeTime time.Time
	Name               string
	Size               int64
	Etag               string
	Shared             bool
	// UserPermission contains the permissions for the authenticated user on this file
	UserPermission *drive.Permission
	// CacheChecksum when set avoids recomputation of checksums
	CacheChecksum bool
	// Monotonically increasing version number for the file
	Version int64
	// The onwers of this file.
	OwnerNames []string
	// Permissions contains the overall permissions for this file
	Permissions           []*drive.Permission
	LastModifyingUsername string
	OriginalFilename      string
	Labels                *drive.FileLabels
	Description           string
}

func NewRemoteFile(f *drive.File) *File {
	return &File{
		AlternateLink:      f.AlternateLink,
		BlobAt:             f.DownloadUrl,
		Copyable:           f.Copyable,
		Etag:               f.Etag,
		ExportLinks:        f.ExportLinks,
		Id:                 f.Id,
		IsDir:              f.MimeType == DriveFolderMimeType,
		Md5Checksum:        f.Md5Checksum,
		MimeType:           f.MimeType,
		ModTime:            parseTimeAndRound(f.ModifiedDate),
		LastViewedByMeTime: parseTimeAndRound(f.LastViewedByMeDate),
		// We must convert each title to match that on the FS.
		Name:                  urlToPath(f.Title, true),
		Size:                  f.FileSize,
		Shared:                f.Shared,
		UserPermission:        f.UserPermission,
		Version:               f.Version,
		OwnerNames:            f.OwnerNames,
		Permissions:           f.Permissions,
		LastModifyingUsername: f.LastModifyingUserName,
		OriginalFilename:      f.OriginalFilename,
		Labels:                f.Labels,
		Description:           f.Description,
	}
}

func DupFile(f *File) *File {
	if f == nil {
		return f
	}

	return &File{
		BlobAt:      f.BlobAt,
		Etag:        f.Etag,
		ExportLinks: f.ExportLinks,
		Id:          f.Id,
		IsDir:       f.IsDir,
		Md5Checksum: f.Md5Checksum,
		MimeType:    f.MimeType,
		ModTime:     f.ModTime,
		Copyable:    f.Copyable,
		// We must convert each title to match that on the FS.
		Name:               f.Name,
		Size:               f.Size,
		Shared:             f.Shared,
		UserPermission:     f.UserPermission,
		Version:            f.Version,
		OwnerNames:         f.OwnerNames,
		Permissions:        f.Permissions,
		LastViewedByMeTime: f.LastViewedByMeTime,
		Labels:             f.Labels,
		AlternateLink:      f.AlternateLink,
		OriginalFilename:   f.OriginalFilename,
		Description:        f.Description,
	}
}

func NewLocalFile(absPath string, f os.FileInfo) *File {
	return &File{
		Id:      "",
		Name:    f.Name(),
		ModTime: f.ModTime().Round(time.Second),
		IsDir:   f.IsDir(),
		Size:    f.Size(),
		BlobAt:  absPath,
		// TODO: Read the CacheChecksum toggle dynamically if set
		// by the requester ie if the file is rapidly changing.
		CacheChecksum: true,
	}
}

func fauxLocalFile(relToRootPath string) *File {
	return &File{
		Id:      "",
		IsDir:   false,
		ModTime: time.Now(),
		Name:    relToRootPath,
		Size:    0,
	}
}

func (f *File) Url() (url string) {
	if f == nil {
		return
	}

	if hasExportLinks(f) {
		return f.AlternateLink
	}

	if f.Id != "" {
		url = fmt.Sprintf("%s/open?id=%s", DriveResourceEntryURL, f.Id)
	}

	return
}

func (f *File) localAliases(prefix string) (aliases []string) {
	aliases = append(aliases, prefix)

	if f == nil {
		return
	}

	suffixes := []string{}

	if runtime.GOOS == OSLinuxKey && hasExportLinks(f) {
		suffixes = append(suffixes, DesktopExtension)
	}

	for _, suffix := range suffixes {
		join := sepJoin(".", prefix, suffix)
		aliases = append(aliases, join)
	}

	return
}

type Change struct {
	Dest           *File
	Parent         string
	Path           string
	Src            *File
	Force          bool
	NoClobber      bool
	IgnoreConflict bool
	IgnoreChecksum bool
	g              *Commands
}

type ByPrecedence []*Change

func (cl ByPrecedence) Less(i, j int) bool {
	if cl[i] == nil {
		return false
	}
	if cl[j] == nil {
		return true
	}

	c1, c2 := cl[i], cl[j]
	return opPrecedence[c1.Op()] < opPrecedence[c2.Op()]
}

func (cl ByPrecedence) Len() int {
	return len(cl)
}

func (cl ByPrecedence) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
}

func (self *File) sameDirType(other *File) bool {
	return other != nil && self.IsDir == other.IsDir
}

func (op *Operation) description() (symbol, info string) {
	switch *op {
	case OpAdd:
		return "\033[32m+\033[0m", "Addition"
	case OpDelete:
		return "\033[31m-\033[0m", "Deletion"
	case OpMod:
		return "\033[33mM\033[0m", "Modification"
	case OpIndexAddition:
		return "\033[34mI+\033[0m", "Index addition"
	case OpModConflict:
		return "\033[35mX\033[0m", "Clashing modification"
	default:
		return "", ""
	}
}

func (f *File) largeFile() bool {
	return f.Size > BigFileSize
}

func (c *Change) Symbol() string {
	op := c.Op()
	symbol, _ := op.description()
	return symbol
}

func md5Checksum(f *File) string {
	if f == nil || f.IsDir {
		return ""
	}
	if f.Md5Checksum != "" {
		return f.Md5Checksum
	}

	if f.largeFile() { // Just warn the user in case of impatience.
		// TODO: Only turn on warnings if verbosity is set.
		fmt.Printf("\033[91mmd5Checksum\033[00m: `%s` (%v)\nmight take time to checksum.\n",
			f.Name, prettyBytes(f.Size))
	}
	fh, err := os.Open(f.BlobAt)

	if err != nil {
		return ""
	}
	defer fh.Close()

	h := md5.New()
	_, err = io.Copy(h, fh)
	if err != nil {
		return ""
	}
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	if f.CacheChecksum {
		f.Md5Checksum = checksum
	}
	return checksum
}

func checksumDiffers(mask int) bool {
	return (mask & DifferMd5Checksum) != 0
}

func dirTypeDiffers(mask int) bool {
	return (mask & DifferDirType) != 0
}

func modTimeDiffers(mask int) bool {
	return (mask & DifferModTime) != 0
}

func sizeDiffers(mask int) bool {
	return (mask & DifferSize) != 0
}

func fileDifferences(src, dest *File, ignoreChecksum bool) int {
	if src == nil || dest == nil {
		return DifferMd5Checksum | DifferSize | DifferModTime | DifferDirType
	}

	difference := DifferNone
	if src.Size != dest.Size {
		difference |= DifferSize
	}
	if !src.ModTime.Equal(dest.ModTime) {
		difference |= DifferModTime
	}
	if src.IsDir != dest.IsDir {
		difference |= DifferDirType
	}

	if ignoreChecksum {
		if sizeDiffers(difference) {
			difference |= DifferMd5Checksum
		}
	} else {
		// Only compute the checksum if the size differs
		if sizeDiffers(difference) || md5Checksum(src) != md5Checksum(dest) {
			difference |= DifferMd5Checksum
		}
	}
	return difference
}

func (c *Change) crudValue() CrudValue {
	op := c.Op()
	if op == OpAdd {
		return Create
	}
	if op == OpMod || op == OpModConflict {
		return Update
	}
	if op == OpDelete {
		return Delete
	}
	return None
}

func (c *Change) op() Operation {
	if c == nil {
		return OpNone
	}

	if c.Src == nil && c.Dest == nil {
		return OpNone
	}

	indexingOnly := false
	if c.g != nil && c.g.opts != nil {
		indexingOnly = c.g.opts.indexingOnly
	}

	if c.Src != nil && c.Dest == nil {
		return indexExistanceOrDeferTo(c, OpAdd, indexingOnly)
	}
	if c.Src == nil && c.Dest != nil {
		return indexExistanceOrDeferTo(c, OpDelete, indexingOnly)
	}
	if c.Src.IsDir != c.Dest.IsDir {
		return indexExistanceOrDeferTo(c, OpMod, indexingOnly)
	}
	if c.Src.IsDir {
		return indexExistanceOrDeferTo(c, OpNone, indexingOnly)
	}

	if indexingOnly {
		return indexExistanceOrDeferTo(c, OpNone, indexingOnly)
	}

	mask := fileDifferences(c.Src, c.Dest, c.IgnoreChecksum)

	if sizeDiffers(mask) || checksumDiffers(mask) {
		if c.IgnoreConflict {
			return OpMod
		}
		return OpModConflict
	}
	if modTimeDiffers(mask) {
		return OpMod
	}
	return OpNone
}

func indexExistanceOrDeferTo(c *Change, deferTo Operation, indexingOnly bool) Operation {
	if !indexingOnly {
		return deferTo
	} else if deferTo != OpNone {
		return OpNone
	}
	ok, err := c.checkIndexExistance()
	if err != nil && err != config.ErrNoSuchDbKey {
		c.g.log.LogErrf("checkIndexExists: \"%s\" %v\n", c.Path, err)
	}

	if !ok {
		return OpIndexAddition
	}

	return deferTo
}

func (c *Change) checkIndexExistance() (bool, error) {
	var f *File = nil
	if c.Src != nil && c.Src.Id != "" {
		f = c.Src
	} else if c.Dest != nil && c.Dest.Id != "" {
		f = c.Dest
	}

	if f == nil || f.Id == "" {
		return false, nil
	}

	index, err := c.g.context.DeserializeIndex(f.Id)
	if err != nil {
		return false, err
	}

	exists := index != nil
	return exists, nil
}

func (c *Change) Op() Operation {
	if c == nil {
		return OpNone
	}

	op := c.op()
	if c.Force {
		if op == OpModConflict {
			return OpMod
		} else if op == OpNone {
			return OpAdd
		}

		return op
	}
	if op != OpAdd && c.NoClobber {
		return OpNone
	}
	return op
}

func (f *File) ToIndex() *config.Index {
	return &config.Index{
		FileId:      f.Id,
		Etag:        f.Etag,
		Md5Checksum: f.Md5Checksum,
		MimeType:    f.MimeType,
		ModTime:     f.ModTime.Unix(),
		Version:     f.Version,
	}
}

type fuzzyStringsValuePair struct {
	fuzzyLevel fuzziness
	inTrash    bool
	joiner     joiner
	values     []string
}

type matchQuery struct {
	dirPath           string
	inTrash           bool
	keywordSearches   []fuzzyStringsValuePair
	mimeQuerySearches []fuzzyStringsValuePair
	titleSearches     []fuzzyStringsValuePair
	ownerSearches     []fuzzyStringsValuePair
}

type fuzziness int

const (
	Not fuzziness = 1 << iota
	Like
	NotIn
	Is
)

func (fz *fuzziness) Stringer() string {
	switch *fz {
	case Not:
		return "!="
	case NotIn:
		return "not in"
	case Like:
		return "contains"
	case Is:
		return "="
	}

	return "="
}

type joiner int

const (
	Or joiner = 1 << iota
	And
)

func (jn *joiner) Stringer() string {
	switch *jn {
	case Or:
		return "or"
	case And:
		return "and"
	}
	return "or"
}

func mimeQueryStringify(fz *fuzzyStringsValuePair) string {
	fuzzyDesc := fz.fuzzyLevel.Stringer()

	keySearches := []string{}
	quote := strconv.Quote

	for _, query := range fz.values {
		resolvedMimeType := mimeTypeFromQuery(query)

		// If it cannot be resolved, use the value passed in
		if resolvedMimeType == "" {
			resolvedMimeType = query
		}

		keySearches = append(keySearches, fmt.Sprintf("(mimeType %s %s)", fuzzyDesc, quote(resolvedMimeType)))
	}

	return strings.Join(keySearches, fmt.Sprintf(" %s ", fz.joiner.Stringer()))
}

func ownerQueryStringify(fz *fuzzyStringsValuePair) string {
	keySearches := []string{}
	quote := strconv.Quote

	for _, owner := range fz.values {
		query := ""
		switch fz.fuzzyLevel {
		case NotIn:
			query = fmt.Sprintf("(not %s in owners)", quote(owner))
		default:
			fuzzyDesc := fz.fuzzyLevel.Stringer()
			query = fmt.Sprintf("(%s %s owners)", quote(owner), fuzzyDesc)
		}

		keySearches = append(keySearches, query)
	}

	return strings.Join(keySearches, fmt.Sprintf(" %s ", fz.joiner.Stringer()))
}

func titleQueryStringify(fz *fuzzyStringsValuePair) string {
	fuzzyDesc := fz.fuzzyLevel.Stringer()

	keySearches := []string{}
	quote := strconv.Quote

	for _, title := range fz.values {
		keySearches = append(keySearches, fmt.Sprintf("(title %s %s and trashed=%v)", fuzzyDesc, quote(title), fz.inTrash))
	}

	return strings.Join(keySearches, fmt.Sprintf(" %s ", fz.joiner.Stringer()))
}

func joinLists(exprJoiner string, expressions []string) string {
	reduced := strings.TrimSpace(strings.Join(expressions, exprJoiner))
	if reduced != "" {
		reduced = fmt.Sprintf("(%s)", reduced)
	}

	return reduced
}

func (mq *matchQuery) Stringer() string {
	overallSearchList := []string{}

	mimeTranslations := []string{}
	for _, fzPair := range mq.mimeQuerySearches {
		query := mimeQueryStringify(&fzPair)
		if query == "" {
			continue
		}

		mimeTranslations = append(mimeTranslations, query)
	}

	titleTranslations := []string{}
	for _, titleFzPair := range mq.titleSearches {
		titleQuery := titleQueryStringify(&titleFzPair)
		if titleQuery == "" {
			continue
		}

		titleTranslations = append(titleTranslations, titleQuery)
	}

	ownerTranslations := []string{}
	for _, ownerFzPair := range mq.ownerSearches {
		ownerQuery := ownerQueryStringify(&ownerFzPair)
		if ownerQuery == "" {
			continue
		}

		ownerTranslations = append(ownerTranslations, ownerQuery)
	}

	exprPairs := []struct {
		joiner   string
		elements []string
	}{
		{" and ", mimeTranslations},
		{" and ", titleTranslations},
		{" and ", ownerTranslations},
	}

	for _, exprPair := range exprPairs {
		expr := joinLists(exprPair.joiner, exprPair.elements)
		if expr == "" {
			continue
		}
		overallSearchList = append(overallSearchList, expr)
	}

	return strings.Join(overallSearchList, " and ")
}
