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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/googleapi"

	expirableCache "github.com/odeke-em/cache"
	spinner "github.com/odeke-em/cli-spinner"
	"github.com/odeke-em/drive/config"
)

var (
	// ErrRejectedTerms is empty "" because messages might be too
	// verbose to affirm a rejection that a user has already seen
	ErrRejectedTerms = errors.New("")
)

const (
	MimeTypeJoiner  = "-"
	RemoteSeparator = "/"

	FmtTimeString           = "2006-01-02T15:04:05.000Z"
	MsgClashesFixedNowRetry = "Clashes were fixed, please retry the operation"
	MsgErrFileNotMutable    = "File not mutable"

	DriveIgnoreSuffix                 = ".driveignore"
	DriveIgnoreNegativeLookAheadToken = "!"
)

const (
	MaxFailedRetryCount = 20 // Arbitrary value
)

var DefaultMaxProcs = runtime.NumCPU()

var BytesPerKB = float64(1024)

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

type jobSt struct {
	id uint64
	do func() (interface{}, error)
}

func (js jobSt) Id() interface{} {
	return js.id
}

func (js jobSt) Do() (interface{}, error) {
	return js.do()
}

type changeJobSt struct {
	change   *Change
	verb     string
	throttle <-chan time.Time
	fn       func(*Change) error
}

func (cjs *changeJobSt) changeJober(g *Commands) func() (interface{}, error) {
	return func() (interface{}, error) {
		ch := cjs.change
		verb := cjs.verb

		canPrintSteps := g.opts.Verbose && g.opts.canPreview()
		if canPrintSteps {
			g.log.Logf("\033[01m%s::Started %s\033[00m\n", verb, ch.Path)
		}

		err := cjs.fn(ch)

		if canPrintSteps {
			g.log.Logf("\033[04m%s::Done %s\033[00m\n", verb, ch.Path)
		}

		<-cjs.throttle
		return ch.Path, err
	}
}

func retryableErrorCheck(v interface{}) (ok, retryable bool) {
	pr, pOk := v.(*tuple)
	if pr == nil || !pOk {
		retryable = true
		return
	}

	if pr.last == nil {
		ok = true
		return
	}

	err, assertOk := pr.last.(*googleapi.Error)
	// In relation to https://github.com/google/google-api-go-client/issues/93
	// where not every error is of googleapi.Error instance e.g io timeout errors
	// etc, let's assume that non-nil errors are retryable

	if !assertOk {
		retryable = true
		return
	}

	if err == nil {
		ok = true
		return
	}

	if strings.EqualFold(err.Message, MsgErrFileNotMutable) {
		retryable = false
		return
	}

	statusCode := err.Code
	if statusCode >= 500 && statusCode <= 599 {
		retryable = true
		return
	}

	switch statusCode {
	case 401, 403:
		retryable = true

		// TODO: Add other errors
	}

	return
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
	switch p {
	case "My Drive", "Meine Ablage", "Mon Drive", "A miña unidade", "Mi unidad",
		"माझा ड्राईव्ह", "मेरी डिस्क", "Drive Saya", "Il mio Drive", "Mijn Drive", "我的雲端硬碟":
		// TODO: Crowd source more language translations here
		// as per https://github.com/odeke-em/drive/issues/1015
		return true
	default:
		return false
	}
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

func _centricPathJoin(sep string, segments ...string) string {
	// Always ensure that the first segment is the separator
	segments = append([]string{sep}, segments...)
	joins := sepJoinNonEmpty(sep, segments...)
	return path.Clean(joins)
}

func localPathJoin(segments ...string) string {
	return _centricPathJoin(UnescapedPathSep, segments...)
}

func remotePathJoin(segments ...string) string {
	return _centricPathJoin(RemoteSeparator, segments...)
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

func promptForChanges(args ...interface{}) Agreement {
	argv := []interface{}{
		"Proceed with the changes? [Y/n]: ",
	}
	if len(args) >= 1 {
		argv = args
	}

	input := prompt(os.Stdin, os.Stdout, argv...)

	if input == "" {
		input = YesShortKey
	}

	if strings.ToUpper(input) == YesShortKey {
		return Accepted
	}

	return Rejected
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

func fReadFile_(f io.Reader, ignorer func(string) bool) (clauses []string, err error) {
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
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

func readFileFromStdin(ignorer func(string) bool) (clauses []string, err error) {
	return fReadFile_(os.Stdin, ignorer)
}

func readFile_(p string, ignorer func(string) bool) (clauses []string, err error) {
	f, fErr := os.Open(p)
	if fErr != nil || f == nil {
		err = fErr
		return
	}

	defer f.Close()

	clauses, err = fReadFile_(f, ignorer)

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
	"ods": "application/vnd.oasis.opendocument.spreadsheet",
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

	// docs and docx should not collide if "docx?" is used so terminate with "$"
	"docx?$": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"ppt$":   "application/vnd.ms-powerpoint",
	"pptx$":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"tsv":    "text/tab-separated-values",
	"xlsx?":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",

	"ipynb": "application/vnd.google.colaboratory",
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

func anyMatch(ignore func(string) bool, args ...string) bool {
	if ignore == nil {
		return false
	}

	for _, arg := range args {
		if ignore(arg) {
			return true
		}
	}
	return false
}

func siftExcludes(clauses []string) (excludes, includes []string) {
	alreadySeenExclude := map[string]bool{}
	alreadySeenInclude := map[string]bool{}

	for _, clause := range clauses {
		memoizerPtr := &alreadySeenExclude
		ptr := &excludes
		// Because Go lacks the negative lookahead ?! capability
		// it is necessary to avoid
		if strings.HasPrefix(clause, DriveIgnoreNegativeLookAheadToken) {
			ptr = &includes
			rest := strings.Split(clause, DriveIgnoreNegativeLookAheadToken)
			if len(rest) > 1 {
				memoizerPtr = &alreadySeenInclude
				clause = sepJoin("", rest[1:]...)
			}
		}

		memoizer := *memoizerPtr
		if _, alreadySeen := memoizer[clause]; alreadySeen {
			continue
		}

		*ptr = append(*ptr, clause)
		memoizer[clause] = true
	}
	return
}

func ignorerByClause(clauses ...string) (ignorer func(string) bool, err error) {
	if len(clauses) < 1 {
		return nil, nil
	}

	excludes, includes := siftExcludes(clauses)

	var excludesRegComp, includesComp *regexp.Regexp
	if len(excludes) >= 1 {
		excRegComp, excRegErr := regexp.Compile(strings.Join(excludes, "|"))
		if excRegErr != nil {
			err = combineErrors(err, makeErrorWithStatus("excludeIgnoreRegErr", excRegErr, StatusIllogicalState))
			return
		}
		excludesRegComp = excRegComp
	}

	if len(includes) >= 1 {
		incRegComp, incRegErr := regexp.Compile(strings.Join(includes, "|"))
		if incRegErr != nil {
			err = combineErrors(err, makeErrorWithStatus("includeIgnoreRegErr", incRegErr, StatusIllogicalState))
			return
		}
		includesComp = incRegComp
	}

	ignorer = func(s string) bool {
		sb := []byte(s)
		if excludesRegComp != nil && excludesRegComp.Match(sb) {
			if includesComp != nil && includesComp.Match(sb) {
				return false
			}
			return true
		}
		return false
	}

	return ignorer, nil
}

func combineIgnores(ignoresPath string) (ignorer func(string) bool, err error) {
	clauses, err := readCommentedFile(ignoresPath, "#")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// TODO: Should internalIgnores only be added only
	// after all the exclusion and exclusion steps.
	clauses = append(clauses, internalIgnores()...)

	return ignorerByClause(clauses...)
}

var mimeTypeFromQuery = cacher(regMapper(regExtStrMap, map[string]string{
	"docs":                 "application/vnd.google-apps.document",
	"folder":               DriveFolderMimeType,
	"form":                 "application/vnd.google-apps.form",
	"mp4":                  "video/mp4",
	"slides?|presentation": "application/vnd.google-apps.presentation",
	"sheet":                "application/vnd.google-apps.spreadsheet",
	"script":               "application/vnd.google-apps.script",
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
	maxProcs, err := strconv.ParseInt(os.Getenv(DriveGoMaxProcsKey), 10, 0)
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
	/*
	   See
	      + https://github.com/golang/go/issues/11511
	      + https://github.com/odeke-em/drive/issues/250
	*/
	return "\"" + strings.Replace(strings.Replace(s, "\\", "\\\\", -1), "\"", "\\\"", -1) + "\""
}

func newExpirableCacheValue(v interface{}) *expirableCache.ExpirableValue {
	return expirableCache.NewExpirableValueWithOffset(v, uint64(time.Hour))
}

func combineErrors(prevErr, supplementaryErr error) error {
	if prevErr == nil && supplementaryErr == nil {
		return nil
	}

	if supplementaryErr == nil {
		return prevErr
	}

	newErr := reComposeError(prevErr, supplementaryErr.Error())
	if codedErr, ok := supplementaryErr.(*Error); ok {
		return makeError(newErr, ErrorStatus(codedErr.Code()))
	}

	return newErr
}

// copyErrStatus copies the error status code from fromErr
// to toErr only if toErr and fromErr are both non-nil.
func copyErrStatusCode(toErr, fromErr error) error {
	if toErr == nil || fromErr == nil {
		return toErr
	}
	codedErr, hasCode := fromErr.(*Error)
	if hasCode {
		toErr = makeError(toErr, ErrorStatus(codedErr.Code()))
	}
	return toErr
}

func reComposeError(prevErr error, messages ...string) error {
	if len(messages) < 1 {
		return prevErr
	}

	joinedMessage := messages[0]
	for i, n := 1, len(messages); i < n; i++ {
		joinedMessage = fmt.Sprintf("%s\n%s", joinedMessage, messages[i])
	}

	if prevErr == nil {
		if len(joinedMessage) < 1 {
			return nil
		}
	} else {
		joinedMessage = fmt.Sprintf("%v\n%s", prevErr, joinedMessage)
	}

	err := errors.New(joinedMessage)
	codedErr, hasCode := prevErr.(*Error)
	if !hasCode {
		return err
	}

	return makeError(err, ErrorStatus(codedErr.Code()))
}

func CopyOptionsFromKeysIfNotSet(fromPtr, toPtr *Options, alreadySetKeys map[string]bool) {
	from := *fromPtr
	fromValue := reflect.ValueOf(from)
	toValue := reflect.ValueOf(toPtr).Elem()

	fromType := reflect.TypeOf(from)

	for i, n := 0, fromType.NumField(); i < n; i++ {
		fromFieldT := fromType.Field(i)
		fromTag := fromFieldT.Tag.Get("cli")

		_, alreadySet := alreadySetKeys[fromTag]
		if alreadySet {
			continue
		}

		fromFieldV := fromValue.Field(i)
		toFieldV := toValue.Field(i)

		if !toFieldV.CanSet() {
			continue
		}

		toFieldV.Set(fromFieldV)
	}
}

type CliSifter struct {
	From           interface{}
	Defaults       map[string]interface{}
	AlreadyDefined map[string]bool
}

func SiftCliTags(cs *CliSifter) string {
	from := cs.From
	defaults := cs.Defaults
	alreadyDefined := cs.AlreadyDefined

	fromValue := reflect.ValueOf(from)
	fromType := reflect.TypeOf(from)

	mapping := map[string]string{}

	for i, n := 0, fromType.NumField(); i < n; i++ {
		fromFieldT := fromType.Field(i)
		fromTag := fromFieldT.Tag.Get("json")

		if fromTag == "" {
			continue
		}

		fromFieldV := fromValue.Field(i)

		elem := fromFieldV.Elem()

		if _, defined := alreadyDefined[fromTag]; !defined {
			if retr, defaultSet := defaults[fromTag]; defaultSet {
				elem = reflect.ValueOf(retr)
			}
		}

		stringified := ""
		switch elem.Kind() {
		case reflect.String:
			stringified = fmt.Sprintf("%q", elem)
		case reflect.Invalid:
			continue
		default:
			stringified = fmt.Sprintf("%v", elem.Interface())
		}

		mapping[fromTag] = stringified
	}

	joined := []string{}

	for k, v := range mapping {
		joined = append(joined, fmt.Sprintf("%q:%v", k, v))
	}

	stringified := sepJoin(",", joined...)

	return fmt.Sprintf("{%v}", stringified)
}

func decrementTraversalDepth(d int) int {
	// Anything less than 0 is a request for infinite traversal
	// 0 is the minimum positive traversal
	if d <= 0 {
		return d
	}

	return d - 1
}

type fsListingArg struct {
	context *config.Context
	parent  string
	hidden  bool
	ignore  func(string) bool
	depth   int
}

func list(flArg *fsListingArg) (fileChan chan *File, err error) {
	context := flArg.context
	p := flArg.parent
	hidden := flArg.hidden
	ignore := flArg.ignore
	depth := flArg.depth

	absPath := context.AbsPathOf(p)
	var f []os.FileInfo
	f, err = ioutil.ReadDir(absPath)
	fileChan = make(chan *File)
	if err != nil {
		close(fileChan)
		return
	}

	go func() {
		defer close(fileChan)

		depth = decrementTraversalDepth(depth)
		if depth == 0 {
			return
		}

		for _, file := range f {
			fileName := file.Name()
			if fileName == config.GDDirSuffix {
				continue
			}
			if isHidden(fileName, hidden) {
				continue
			}

			resPath := path.Join(absPath, fileName)
			if anyMatch(ignore, fileName, resPath) {
				continue
			}

			// TODO: (@odeke-em) decide on how to deal with isFifo
			if namedPipe(file.Mode()) {
				fmt.Fprintf(os.Stderr, "%s (%s) is a named pipe, not reading from it\n", p, resPath)
				continue
			}

			if !symlink(file.Mode()) {
				fileChan <- NewLocalFile(resPath, file)
			} else {
				var symResolvPath string
				symResolvPath, err = filepath.EvalSymlinks(resPath)
				if err != nil {
					continue
				}

				if anyMatch(ignore, symResolvPath) {
					continue
				}

				var symInfo os.FileInfo
				symInfo, err = os.Stat(symResolvPath)
				if err != nil {
					continue
				}

				lf := NewLocalFile(symResolvPath, symInfo)
				// Retain the original name as appeared in
				// the manifest instead of the resolved one
				lf.Name = fileName
				fileChan <- lf
			}
		}
	}()
	return
}

func resolver(g *Commands, byId bool, sources []string, fileOp func(*File) interface{}) (kvChan chan *keyValue) {
	resolve := g.rem.FindByPath
	if byId {
		resolve = g.rem.FindById
	}

	kvChan = make(chan *keyValue)

	go func() {
		defer close(kvChan)

		for _, source := range sources {
			f, err := resolve(source)

			kv := keyValue{key: source, value: err}
			if err == nil {
				kv.value = fileOp(f)
			}

			kvChan <- &kv
		}
	}()

	return kvChan
}

func NotExist(err error) bool {
	return os.IsNotExist(err) || err == ErrPathNotExists
}

func localOpToChangerTranslator(g *Commands, c *Change) func(*Change, []string) error {
	var fn func(*Change, []string) error = nil

	op := c.Op()
	switch op {
	case OpMod:
		fn = g.localMod
	case OpModConflict:
		fn = g.localMod
	case OpAdd:
		fn = g.localAdd
	case OpDelete:
		fn = g.localDelete
	case OpIndexAddition:
		fn = g.localAddIndex
	}
	return fn
}

func remoteOpToChangerTranslator(g *Commands, c *Change) func(*Change) error {
	var fn func(*Change) error = nil

	op := c.Op()
	switch op {
	case OpMod:
		fn = g.remoteMod
	case OpModConflict:
		fn = g.remoteMod
	case OpAdd:
		fn = g.remoteAdd
	case OpDelete:
		fn = g.remoteTrash
	}
	return fn
}

const (
	// Since we'll be sharing `TypeMask` with other bit flags
	// we'll need to avoid collisions with other args
	InTrash int = 1 << (31 - 1 - iota)
	Folder
	Shared
	Owners
	Minimal
	Starred
	NonFolder
	DiskUsageOnly
	CurrentVersion
)

func folderExplicitly(mask int) bool    { return (mask & Folder) == Folder }
func nonFolderExplicitly(mask int) bool { return (mask & NonFolder) == NonFolder }

type driveFileFilter func(*File) bool

func makeFileFilter(mask int) driveFileFilter {
	return func(f *File) bool {
		// TODO: Decide if nil files should be either a pass or fail?
		if f == nil {
			return true
		}
		truths := []bool{}
		if nonFolderExplicitly(mask) {
			truths = append(truths, !f.IsDir)
		}
		// Even though regular file & folder are mutually exclusive let's
		// compare them separately in case we want this logical inconsistency
		// instead of an if...else clause
		if folderExplicitly(mask) {
			truths = append(truths, f.IsDir)
		}

		return allTruthsHold(truths...)
	}
}

func allTruthsHold(truths ...bool) bool {
	for _, truth := range truths {
		if !truth {
			return false
		}
	}
	return true
}

func parseDate(dateStr string, fmtSpecifiers ...string) (*time.Time, error) {
	var err error

	for _, fmtSpecifier := range fmtSpecifiers {
		var t time.Time
		t, err = time.Parse(fmtSpecifier, dateStr)
		if err == nil {
			return &t, err
		}
	}

	return nil, err
}

func parseDurationOffsetFromNow(durationOffsetStr string) (*time.Time, error) {
	d, err := time.ParseDuration(durationOffsetStr)
	if err != nil {
		return nil, err
	}

	offsetFromNow := time.Now().Add(d)
	return &offsetFromNow, nil
}

// Debug returns true if DRIVE_DEBUG is set in the environment.
// Set it to anything non-empty, for example `DRIVE_DEBUG=true`.
func Debug() bool {
	return os.Getenv("DRIVE_DEBUG") != ""
}

func DebugPrintf(fmt_ string, args ...interface{}) {
	FDebugPrintf(os.Stdout, fmt_, args...)
}

// FDebugPrintf will only print output to the out writer if
// environment variable `DRIVE_DEBUG` is set. It prints out a header
// on a newline containing the introspection of the callsite,
// and then the formatted message you'd like,
// appending an obligatory newline at the end.
// The output will be of the form:
// [<FILE>:<FUNCTION>:<LINE_NUMBER>]
// <MSG>\n
func FDebugPrintf(f io.Writer, fmt_ string, args ...interface{}) {
	if !Debug() {
		return
	}
	if f == nil {
		f = os.Stdout
	}

	programCounter, file, line, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(programCounter)
	prefix := fmt.Sprintf("[\033[32m%s:%s:\033[33m%d\033[00m]\n%s\n", file, fn.Name(), line, fmt_)
	fmt.Fprintf(f, prefix, args...)
}
