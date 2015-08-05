// Copyright 2015 Google Inc. All Rights Reserved.
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
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	spinner "github.com/odeke-em/cli-spinner"
)

const (
	MimeTypeJoiner      = "-"
	RemoteDriveRootPath = "My Drive"

	FmtTimeString = "2006-01-02T15:04:05.000Z"
)

const (
	DefaultMaxProcs     = 4
	MaxFailedRetryCount = uint32(20) // Arbitrary value
)

var BytesPerKB = float64(1024)

var (
	MsgInternalServerError = "googleapi: Error 500: Intneral Error, internalError"
	MsgUserLimitExceeded   = "googleapi: Error 403: User rate limit exceeded, userRateLimitExceeded"
)

type desktopEntry struct {
	name string
	url  string
	icon string
}

type playable struct {
	play  func()
	pause func()
	reset func()
	stop  func()
}

func noop() {
}

type tuple struct {
	first  interface{}
	second interface{}
	last   interface{}
}

func retryableErrorCheck(v interface{}) (ok, retryable bool) {
	pr, pOk := v.(*tuple)
	if pr == nil || !pOk {
		return false, true
	}

	err, ok := pr.first.(error)
	if !ok || err == nil {
		return false, false
	}

	switch err.Error() {
	case MsgUserLimitExceeded:
		return false, true
	case MsgInternalServerError:
		return false, true
	default:
		return false, false
	}

	return false, true
}

func noopPlayable() *playable {
	return &playable{
		play:  noop,
		pause: noop,
		reset: noop,
		stop:  noop,
	}
}

func parseTime(ts string, round bool) (t time.Time) {
	t, _ = time.Parse(FmtTimeString, ts)
	if !round {
		return
	}
	return t.Round(time.Second)
}

func parseTimeAndRound(ts string) (t time.Time) {
	return parseTime(ts, true)
}

func internalIgnores() (ignores []string) {
	if runtime.GOOS == OSLinuxKey {
		ignores = append(ignores, "\\.\\s*desktop$")
	}
	return ignores
}

func newPlayable(freq int64) *playable {
	spin := spinner.New(freq)

	play := func() {
		spin.Start()
	}

	return &playable{
		play:  play,
		stop:  spin.Stop,
		reset: spin.Reset,
		pause: spin.Stop,
	}
}

func (g *Commands) playabler() *playable {
	if !g.opts.canPrompt() {
		return noopPlayable()
	}
	return newPlayable(10)
}

func rootLike(p string) bool {
	return p == "/" || p == "" || p == "root"
}

func remoteRootLike(p string) bool {
	return p == RemoteDriveRootPath
}

type byteDescription func(b int64) string

func memoizeBytes() byteDescription {
	cache := map[int64]string{}
	suffixes := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	maxLen := len(suffixes) - 1

	var cacheMu sync.Mutex

	return func(b int64) string {
		cacheMu.Lock()
		defer cacheMu.Unlock()

		description, ok := cache[b]
		if ok {
			return description
		}

		bf := float64(b)
		i := 0
		description = ""
		for {
			if bf/BytesPerKB < 1 || i >= maxLen {
				description = fmt.Sprintf("%.2f%s", bf, suffixes[i])
				break
			}
			bf /= BytesPerKB
			i += 1
		}
		cache[b] = description
		return description
	}
}

var prettyBytes = memoizeBytes()

func sepJoin(sep string, args ...string) string {
	return strings.Join(args, sep)
}

func sepJoinNonEmpty(sep string, args ...string) string {
	nonEmpties := NonEmptyStrings(args...)
	return sepJoin(sep, nonEmpties...)
}

func isHidden(p string, ignore bool) bool {
	if strings.HasPrefix(p, ".") {
		return !ignore
	}
	return false
}

func prompt(r *os.File, w *os.File, promptText ...interface{}) (input string) {

	fmt.Fprint(w, promptText...)

	flushTTYin()

	fmt.Fscanln(r, &input)
	return
}

func nextPage() bool {
	input := prompt(os.Stdin, os.Stdout, "---More---")
	if len(input) >= 1 && strings.ToLower(input[:1]) == QuitShortKey {
		return false
	}
	return true
}

func promptForChanges(args ...interface{}) bool {
	argv := []interface{}{
		"Proceed with the changes? [Y/n]:",
	}
	if len(args) >= 1 {
		argv = args
	}

	input := prompt(os.Stdin, os.Stdout, argv...)

	if input == "" {
		input = YesShortKey
	}

	return strings.ToUpper(input) == YesShortKey
}

func (f *File) toDesktopEntry(urlMExt *urlMimeTypeExt) *desktopEntry {
	name := f.Name
	if urlMExt.ext != "" {
		name = sepJoin("-", f.Name, urlMExt.ext)
	}
	return &desktopEntry{
		name: name,
		url:  urlMExt.url,
		icon: urlMExt.mimeType,
	}
}

func (f *File) serializeAsDesktopEntry(destPath string, urlMExt *urlMimeTypeExt) (int, error) {
	deskEnt := f.toDesktopEntry(urlMExt)
	handle, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}

	defer func() {
		handle.Close()
		chmodErr := os.Chmod(destPath, 0755)

		if chmodErr != nil {
			fmt.Fprintf(os.Stderr, "%s: [desktopEntry]::chmod %v\n", destPath, chmodErr)
		}

		chTimeErr := os.Chtimes(destPath, f.ModTime, f.ModTime)
		if chTimeErr != nil {
			fmt.Fprintf(os.Stderr, "%s: [desktopEntry]::chtime %v\n", destPath, chTimeErr)
		}
	}()

	icon := strings.Replace(deskEnt.icon, UnescapedPathSep, MimeTypeJoiner, -1)

	return fmt.Fprintf(handle, "[Desktop Entry]\nIcon=%s\nName=%s\nType=%s\nURL=%s\n",
		icon, deskEnt.name, LinkKey, deskEnt.url)
}

func remotePathSplit(p string) (dir, base string) {
	// Avoiding use of filepath.Split because of bug with trailing "/" not being stripped
	sp := strings.Split(p, "/")
	spl := len(sp)
	dirL, baseL := sp[:spl-1], sp[spl-1:]
	dir = strings.Join(dirL, "/")
	base = strings.Join(baseL, "/")
	return
}

func commonPrefix(values ...string) string {
	vLen := len(values)
	if vLen < 1 {
		return ""
	}
	minIndex := 0
	min := values[0]
	minLen := len(min)

	for i := 1; i < vLen; i += 1 {
		st := values[i]
		if st == "" {
			return ""
		}
		lst := len(st)
		if lst < minLen {
			min = st
			minLen = lst
			minIndex = i + 0
		}
	}

	prefix := make([]byte, minLen)
	matchOn := true
	for i := 0; i < minLen; i += 1 {
		for j, other := range values {
			if minIndex == j {
				continue
			}
			if other[i] != min[i] {
				matchOn = false
				break
			}
		}
		if !matchOn {
			break
		}
		prefix[i] = min[i]
	}
	return string(prefix)
}

func ReadFullFile(p string) (clauses []string, err error) {
	return readFile_(p, nil)
}

func readFile_(p string, ignorer func(string) bool) (clauses []string, err error) {
	f, fErr := os.Open(p)
	if fErr != nil || f == nil {
		err = fErr
		return
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	for {
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		line = strings.Trim(line, " ")
		line = strings.Trim(line, "\n")
		if ignorer != nil && ignorer(line) {
			continue
		}
		clauses = append(clauses, line)
	}
	return
}

func readCommentedFile(p, comment string) (clauses []string, err error) {
	ignorer := func(line string) bool {
		return strings.HasPrefix(line, comment) || len(line) < 1
	}

	return readFile_(p, ignorer)
}

func chunkInt64(v int64) chan int {
	var maxInt int
	maxInt = 1<<31 - 1
	maxIntCast := int64(maxInt)

	chunks := make(chan int)

	go func() {
		q, r := v/maxIntCast, v%maxIntCast
		for i := int64(0); i < q; i += 1 {
			chunks <- maxInt
		}

		if r > 0 {
			chunks <- int(r)
		}

		close(chunks)
	}()

	return chunks
}

func nonEmptyStrings(fn func(string) string, v ...string) (splits []string) {
	for _, elem := range v {
		if fn != nil {
			elem = fn(elem)
		}
		if elem != "" {
			splits = append(splits, elem)
		}
	}
	return
}

func NonEmptyStrings(v ...string) (splits []string) {
	return nonEmptyStrings(nil, v...)
}

func NonEmptyTrimmedStrings(v ...string) (splits []string) {
	return nonEmptyStrings(strings.TrimSpace, v...)
}

var regExtStrMap = map[string]string{
	"csv":   "text/csv",
	"html?": "text/html",
	"te?xt": "text/plain",
	"xml":   "text/xml",

	"gif":   "image/gif",
	"png":   "image/png",
	"svg":   "image/svg+xml",
	"jpe?g": "image/jpeg",

	"odt": "application/vnd.oasis.opendocument.text",
	"odm": "application/vnd.oasis.opendocument.text-master",
	"ott": "application/vnd.oasis.opendocument.text-template",
	"ods": "application/vnd.oasis.opendocument.sheet",
	"ots": "application/vnd.oasis.opendocument.spreadsheet-template",
	"odg": "application/vnd.oasis.opendocument.graphics",
	"otg": "application/vnd.oasis.opendocument.graphics-template",
	"oth": "application/vnd.oasis.opendocument.text-web",
	"odp": "application/vnd.oasis.opendocument.presentation",
	"otp": "application/vnd.oasis.opendocument.presentation-template",
	"odi": "application/vnd.oasis.opendocument.image",
	"odb": "application/vnd.oasis.opendocument.database",
	"oxt": "application/vnd.openofficeorg.extension",

	"rtf": "application/rtf",
	"pdf": "application/pdf",

	"json": "application/json",
	"js":   "application/x-javascript",

	"apk":   "application/vnd.android.package-archive",
	"bin":   "application/octet-stream",
	"tiff?": "image/tiff",
	"tgz":   "application/x-compressed",
	"zip":   "application/zip",

	"mp3": "audio/mpeg",

	"docx?": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"pptx?": "application/vnd.openxmlformats-officedocument.wordprocessingml.presentation",
	"xlsx?": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

func regMapper(srcMaps ...map[string]string) map[*regexp.Regexp]string {
	regMap := make(map[*regexp.Regexp]string)
	for _, srcMap := range srcMaps {
		for regStr, resolve := range srcMap {
			regExComp, err := regexp.Compile(regStr)
			if err == nil {
				regMap[regExComp] = resolve
			}
		}
	}
	return regMap
}

func cacher(regMap map[*regexp.Regexp]string) func(string) string {
	var cache = make(map[string]string)
	var cacheMu sync.Mutex

	return func(ext string) string {
		cacheMu.Lock()
		defer cacheMu.Unlock()

		memoized, ok := cache[ext]
		if ok {
			return memoized
		}

		bExt := []byte(ext)
		for regEx, mimeType := range regMap {
			if regEx != nil && regEx.Match(bExt) {
				memoized = mimeType
				break
			}
		}

		cache[ext] = memoized
		return memoized
	}
}

func anyMatch(pat *regexp.Regexp, args ...string) bool {
	if pat == nil {
		return false
	}

	for _, arg := range args {
		if pat.Match([]byte(arg)) {
			return true
		}
	}
	return false
}

var mimeTypeFromQuery = cacher(regMapper(regExtStrMap, map[string]string{
	"docs":         "application/vnd.google-apps.document",
	"folder":       DriveFolderMimeType,
	"form":         "application/vnd.google-apps.form",
	"mp4":          "video/mp4",
	"presentation": "application/vnd.google-apps.presentation",
	"sheet":        "application/vnd.google-apps.spreadsheet",
	"script":       "application/vnd.google-apps.script",
}))

var mimeTypeFromExt = cacher(regMapper(regExtStrMap))

func guessMimeType(p string) string {
	resolvedMimeType := mimeTypeFromExt(p)
	return resolvedMimeType
}

func CrudAtoi(ops ...string) CrudValue {
	opValue := None

	for _, op := range ops {
		if len(op) < 1 {
			continue
		}

		first := op[0]
		if first == 'c' || first == 'C' {
			opValue |= Create
		} else if first == 'r' || first == 'R' {
			opValue |= Read
		} else if first == 'u' || first == 'U' {
			opValue |= Update
		} else if first == 'd' || first == 'D' {
			opValue |= Delete
		}
	}

	return opValue
}

func httpOk(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	return _hasAnyAtExtreme(value, strings.HasPrefix, prefixes)
}

func hasAnySuffix(value string, prefixes ...string) bool {
	return _hasAnyAtExtreme(value, strings.HasSuffix, prefixes)
}

func _hasAnyAtExtreme(value string, fn func(string, string) bool, queries []string) bool {
	for _, query := range queries {
		if fn(value, query) {
			return true
		}
	}
	return false
}

func maxProcs() int {
	maxProcs, err := strconv.ParseInt(os.Getenv(GoMaxProcsKey), 10, 0)
	if err != nil {
		return DefaultMaxProcs
	}

	maxProcsInt := int(maxProcs)
	if maxProcsInt < 1 {
		return DefaultMaxProcs
	}

	return maxProcsInt
}

func customQuote(s string) string {
	return "\"" + strings.Replace(strings.Replace(s, "\\", "\\\\", -1), "\"", "\\\"", -1) + "\""
}

type expirableCacheValue struct {
	value     interface{}
	entryTime time.Time
}

func (e *expirableCacheValue) Expired(q time.Time) bool {
	if e == nil {
		return true
	}

	return e.entryTime.Before(q)
}

func (e *expirableCacheValue) Value() interface{} {
	if e == nil {
		return nil
	}

	return e.value
}

func newExpirableCacheValueWithOffset(v interface{}, offset time.Duration) *expirableCacheValue {
	return &expirableCacheValue{
		value:     v,
		entryTime: time.Now().Add(offset),
	}
}

var newExpirableCacheValue = func(v interface{}) *expirableCacheValue {
	return newExpirableCacheValueWithOffset(v, time.Hour)
}
