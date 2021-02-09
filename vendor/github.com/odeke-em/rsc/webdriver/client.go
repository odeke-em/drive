// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Conn struct {
	prefix string
}

type Session struct {
	ID           string `json:"id"`
	Capabilities `json:"capabilities"`
	c            *Conn
}

func Dial(addr string) (*Conn, error) {
	c := &Conn{
		prefix: "http://" + addr,
	}
	// Fetch /status to sanity check address.
	if err := c.get("/status", nil); err != nil {
		return nil, fmt.Errorf("cannot Dial %q: %v", addr, err)
	}
	return c, nil
}

type response struct {
	SessionID *string         `json:"sessionId"`
	Status    int             `json:"status"`
	Value     json.RawMessage `json:"value"`
}

type Failure struct {
	Message    string       `json:"message"`
	Screen     string       `json:"screen"`     // base64-encoded screenshot of page
	Class      string       `json:"class"`      // class name for exception thrown
	StackTrace []StackFrame `json:"stackTrace"` // stack trace for exception
}

type StackFrame struct {
	File   string `json:"fileName"`   // name of source file
	Line   int    `json:"lineNumber"` // line number in source file
	Class  string `json:"className"`  // class active in this frame
	Method string `json:"methodName"` // method active in this frame
}

type Error struct {
	ErrorCode ErrorCode
	Failure   Failure
}

type ErrorCode int

const (
	Success                    ErrorCode = 0
	NoSuchDriver               ErrorCode = 6
	NoSuchElement              ErrorCode = 7
	NoSuchFrame                ErrorCode = 8
	UnknownCommand             ErrorCode = 9
	StaleElementReference      ErrorCode = 10
	ElementNotVisible          ErrorCode = 11
	InvalidElementState        ErrorCode = 12
	UnknownError               ErrorCode = 13
	ElementIsNotSelectable     ErrorCode = 15
	JavaScriptError            ErrorCode = 17
	XPathLookupError           ErrorCode = 19
	TimeoutError               ErrorCode = 21
	NoSuchWindow               ErrorCode = 23
	InvalidCookieDomain        ErrorCode = 24
	UnableToSetCookie          ErrorCode = 25
	UnexpectedAlertOpen        ErrorCode = 26
	NoAlertOpenError           ErrorCode = 27
	ScriptTimeoutError         ErrorCode = 28
	InvalidElementCoordinates  ErrorCode = 29
	IMENotAvailable            ErrorCode = 30
	IMEEngineActivationFailed  ErrorCode = 31
	InvalidSelector            ErrorCode = 32
	SessionNotCreatedException ErrorCode = 33
	MoveTargetOutOfBounds      ErrorCode = 34
)

var errorText = [...]string{
	Success:                    "success",
	NoSuchDriver:               "session is unstarted or terminated",
	NoSuchElement:              "no such element",
	NoSuchFrame:                "no such frame",
	UnknownCommand:             "unknown command or resource",
	StaleElementReference:      "stale element reference",
	ElementNotVisible:          "element not visible",
	InvalidElementState:        "invalid element state",
	UnknownError:               "unknown server-side error",
	ElementIsNotSelectable:     "element is not selectable",
	JavaScriptError:            "javascript error",
	XPathLookupError:           "XPath lookup error",
	TimeoutError:               "timeout expired",
	NoSuchWindow:               "no such window",
	InvalidCookieDomain:        "invalid cookie domain",
	UnableToSetCookie:          "unable to set cookie",
	UnexpectedAlertOpen:        "unexpected alert open",
	NoAlertOpenError:           "no alert open",
	ScriptTimeoutError:         "script timeout expired",
	InvalidElementCoordinates:  "invalid element coordinates",
	IMENotAvailable:            "IME not available",
	IMEEngineActivationFailed:  "IME engine could not be started",
	InvalidSelector:            "invalid selector",
	SessionNotCreatedException: "session could not be created",
	MoveTargetOutOfBounds:      "move target out of bounds",
}

func (e ErrorCode) String() string {
	if e >= 0 && int(e) < len(errorText) && errorText[e] != "" {
		return errorText[e]
	}
	return fmt.Sprintf("ErrorCode(%d)", e)
}

func (e *Error) Error() string {
	if e.Failure.Message != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Failure.Message)
	}
	return e.ErrorCode.String()
}

type Element struct {
	ID string `json:"ELEMENT"`
	s  *Session
	c  *Conn
}

type Capabilities struct {
	Browser      string `json:"browserName,omitempty"`
	Version      string `json:"version,omitempty"`
	Platform     string `json:"platform,omitempty"`
	Javascript   bool   `json:"javascriptEnabled,omitempty"`
	Screenshot   bool   `json:"takesScreenshot,omitempty"`
	Alert        bool   `json:"handlesAlerts,omitempty"`
	Database     bool   `json:"databaseEnabled,omitempty"`
	Location     bool   `json:"locationContextEnabled,omitempty"`
	AppCache     bool   `json:"applicationCacheEnabled,omitempty"`
	Connectivity bool   `json:"browserConnectionEnabled,omitempty"`
	CSSSelectors bool   `json:"cssSelectorsEnabled,omitempty"`
	WebStorage   bool   `json:"webStorageEnabled,omitempty"`
	Rotate       bool   `json:"rotatable,omitempty"`
	InsecureSSL  bool   `json:"acceptSslCerts,omitempty"`
	NativeEvents bool   `json:"nativeEvents,omitempty"`
	// TODO: Proxy
}

type Cookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Path   string `json:"path,omitempty"`
	Domain string `json:"domain,omitempty"`
	Secure bool   `json:"secure,omitempty"`
	Expiry int64  `json:"expiry,omitempty"` // Unix seconds
}

func (c *Conn) do(verb, path string, body io.Reader, value interface{}) (*http.Response, *response, error) {
	req, err := http.NewRequest(verb, c.prefix+path, body)
	if err != nil {
		return nil, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		if len(data) == 0 {
			return resp, nil, fmt.Errorf("%s %s: %v", verb, path, resp.Status)
		}
		return resp, nil, fmt.Errorf("%s %s: %v\n%s", verb, path, resp.Status, data)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		return resp, nil, fmt.Errorf("%s %s: invalid Content-Type %s", verb, path, ct)
	}

	var jr response
	err = json.NewDecoder(resp.Body).Decode(&jr)
	if err != nil {
		return resp, nil, err
	}

	if jr.Status != 0 {
		var fail Failure
		json.Unmarshal(jr.Value, &fail)
		return resp, &jr, &Error{
			ErrorCode: ErrorCode(jr.Status),
			Failure:   fail,
		}
	}

	if value != nil {
		err = json.Unmarshal(jr.Value, value)
		if err != nil {
			return resp, &jr, fmt.Errorf("%s %s: decoding json: %v", verb, path, err)
		}
	}
	return resp, &jr, nil
}

func (c *Conn) get(path string, value interface{}) error {
	_, _, err := c.do("GET", path, nil, value)
	return err
}

func (c *Conn) post(path string, args, value interface{}) error {
	var jv []byte
	if args != nil {
		var err error
		jv, err = json.Marshal(args)
		if err != nil {
			return err
		}
	}
	_, _, err := c.do("POST", path, bytes.NewReader(jv), value)
	return err
}

func (c *Conn) delete(path string, value interface{}) error {
	_, _, err := c.do("DELETE", path, nil, value)
	return err
}

func (c *Conn) postRedirect(path string, args, value interface{}) (string, error) {
	jv, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	resp, hdr, err := c.do("POST", path, bytes.NewReader(jv), value)

	// NOTE: chromedriver responds with 200 OK instead of 303.
	if path == "/session" && resp.StatusCode == 200 && hdr.SessionID != nil {
		return "/session/" + *hdr.SessionID, nil
	}

	if resp.StatusCode != 303 {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("POST %s: expected 303 redirect, got %s", path, resp.Status)
	}
	url := resp.Header.Get("Location")
	if !strings.HasPrefix(url, "/") {
		return url, fmt.Errorf("POST %s: expected redirect to /..., got %s", path, url)
	}

	err = c.get(url, value)
	if err != nil {
		return url, err
	}

	return url, nil
}

type Status struct {
	Build struct { // build information
		Version  string `json:"version"`  // release label
		Revision string `json:"revision"` // revision of local source control client
		Time     string `json:"time"`     // build timestamp (format unspecified)
	} `json:"build"`
	OS struct {
		Arch    string `json:"arch"`    // system architecture
		Name    string `json:"name"`    // name of operating system
		Version string `json:"version"` // version of operating system
	} `json:"os"`
}

func (c *Conn) Status() (*Status, error) {
	var st Status
	if err := c.get("/status", &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (c *Conn) NewSession(cap *Capabilities) (*Session, error) {
	var args struct {
		Desired Capabilities `json:"desiredCapabilities"`
	}
	if cap != nil {
		args.Desired = *cap
	}
	s := &Session{
		c: c,
	}
	url, err := c.postRedirect("/session", &args, &s.Capabilities)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(url, "/session/") {
		return nil, fmt.Errorf("unexpected session redirect to %s", url)
	}
	s.ID = url[len("/session/"):]
	return s, nil
}

func (c *Conn) Sessions() ([]*Session, error) {
	var sess []*Session
	if err := c.get("/sessions", &sess); err != nil {
		return nil, err
	}
	for _, s := range sess {
		s.c = c
	}
	return sess, nil
}

func (s *Session) path(suffix string) string {
	return "/session/" + s.ID + suffix
}

func (s *Session) Delete() error {
	// DELETE /session/id
	return s.c.delete(s.path(""), nil)
}

type Timeout string

const (
	ScriptTimeout       Timeout = "script"
	ImplicitTimeout     Timeout = "implicit"
	PageLoadTimeout     Timeout = "page load"
	AsyncScriptTimeout          = "/async_script"
	ImplicitWaitTimeout Timeout = "/implicit_wait"
)

func (s *Session) SetTimeout(t Timeout, d time.Duration) error {
	// TODO: What's the difference between posting "implicit" and using implicit_wait?
	var args struct {
		MS   int64  `json:"ms"`
		Type string `json:"type,omitempty"`
	}
	url := s.path("/timeouts")
	if strings.HasPrefix(string(t), "/") {
		url += string(t)
	} else {
		args.Type = string(t)
	}
	args.MS = int64(d / time.Millisecond)
	return s.c.post(url, &args, nil)
}

type Window struct {
	c  *Conn
	s  *Session
	ID string
}

func (w *Window) path(suffix string) string {
	return w.s.path("/window/" + w.ID + suffix)
}

func (s *Session) Window() (*Window, error) {
	w := &Window{
		s: s,
		c: s.c,
	}
	err := s.c.get(s.path("/window_handle"), &w.ID)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Session) Windows() ([]*Window, error) {
	var ids []string
	err := s.c.get(s.path("/window_handles"), &ids)
	if err != nil {
		return nil, err
	}
	var ws []*Window
	for _, id := range ids {
		ws = append(ws, &Window{
			s:  s,
			c:  s.c,
			ID: id,
		})
	}
	return ws, nil
}

func (s *Session) URL() (string, error) {
	var url string
	err := s.c.get(s.path("/url"), &url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *Session) SetURL(url string) error {
	var args struct {
		URL string `json:"url"`
	}
	args.URL = url
	return s.c.post(s.path("/url"), &args, nil)
}

func (s *Session) Forward() error {
	return s.c.post(s.path("/forward"), nil, nil)
}

func (s *Session) Back() error {
	return s.c.post(s.path("/back"), nil, nil)
}

func (s *Session) Refresh() error {
	return s.c.post(s.path("/refresh"), nil, nil)
}

/*
func (s *Session) Execute(javscript string) error {
	// POST execute
}

func (s *Session) ExecuteAsync(javscript string) error {
	// POST execute_async
}

func (s *Session) ScreenShot() ([]byte, error) {
	// GET screenshot + decode
}

func (s *Session) IMEs() ([]*IME, error) {
	// GET ime/available_engines
}

func (s *Session) IME() (*IME, error) {
	// GET ime/active_engine
}

func (s *session) SetIME(ime *IME) error {
	// POST ime/activate
	// POST ime/deactivate
}

func (s *Session) IMEActive() (bool, error) {
	// GET ime/activated
}

func (s *Session) SetFrame(xxx) error {
	// POST frame
}
*/

func (s *Session) SetWindow(w *Window) error {
	var args struct {
		Name string `json:"name"`
	}
	return s.c.post(s.path("/window"), &args, nil)
}

func (s *Session) CloseWindow() error {
	return s.c.delete(s.path("/window"), nil)
}

func (w *Window) Resize(dx, dy int) error {
	var args struct {
		DX int `json:"width"`
		DY int `json:"height"`
	}
	args.DX = dx
	args.DY = dy
	return w.c.post(w.path("/size"), &args, nil)
}

func (w *Window) Size() (dx, dy int, err error) {
	var args struct {
		DX int `json:"width"`
		DY int `json:"height"`
	}
	if err := w.c.get(w.path("/size"), &args); err != nil {
		return 0, 0, err
	}
	return args.DX, args.DY, nil
}

func (w *Window) Maximize() error {
	return w.c.post(w.path("/maximize"), nil, nil)
}

func (w *Window) Position() (x, y int, err error) {
	var args struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := w.c.get(w.path("/position"), &args); err != nil {
		return 0, 0, err
	}
	return args.X, args.Y, nil
}

// TODO: Cookie

func (s *Session) Source() (string, error) {
	var src string
	err := s.c.get(s.path("/source"), &src)
	return src, err
}

func (s *Session) Title() (string, error) {
	var title string
	err := s.c.get(s.path("/title"), &title)
	return title, err
}

type Strategy string

const (
	ByClassName       Strategy = "class name"
	ByCSSSelector     Strategy = "css selector"
	ByID              Strategy = "id"
	ByName            Strategy = "name"
	ByLinkText        Strategy = "link text"
	ByPartialLinkText Strategy = "partial link text"
	ByTagName         Strategy = "tag name"
	ByXPath           Strategy = "xpath"
)

func (s *Session) Element(by Strategy, desc string) (*Element, error) {
	var args struct {
		Using string `json:"using"`
		Value string `json:"value"`
	}
	args.Using = string(by)
	args.Value = desc

	var elem Element
	err := s.c.post(s.path("/element"), &args, &elem)
	if err != nil {
		return nil, err
	}
	elem.s = s
	elem.c = s.c
	return &elem, nil
}

func (s *Session) Elements(by Strategy, desc string) ([]*Element, error) {
	var args struct {
		Using string `json:"using"`
		Value string `json:"value"`
	}
	args.Using = string(by)
	args.Value = desc

	var elems []*Element
	err := s.c.post(s.path("/elements"), &args, &elems)
	if err != nil {
		return nil, err
	}
	for _, elem := range elems {
		elem.s = s
		elem.c = s.c
	}
	return elems, nil
}

func (s *Session) ActiveElement() (*Element, error) {
	var elem Element
	err := s.c.get(s.path("/element/active"), &elem)
	if err != nil {
		return nil, err
	}
	elem.s = s
	elem.c = s.c
	return &elem, nil
}

func (e *Element) path(suffix string) string {
	return e.s.path("/element/" + e.ID + suffix)
}

func (e *Element) Element(by Strategy, desc string) (*Element, error) {
	var args struct {
		Using string `json:"using"`
		Value string `json:"value"`
	}
	args.Using = string(by)
	args.Value = desc

	var elem Element
	err := e.c.post(e.path("/element"), &args, &elem)
	if err != nil {
		return nil, err
	}
	elem.s = e.s
	elem.c = e.c
	return &elem, nil
}

func (e *Element) Elements(by Strategy, desc string) ([]*Element, error) {
	var args struct {
		Using string `json:"using"`
		Value string `json:"value"`
	}
	args.Using = string(by)
	args.Value = desc

	var elems []*Element
	err := e.c.post(e.path("/elements"), &args, &elems)
	if err != nil {
		return nil, err
	}
	for _, elem := range elems {
		elem.s = e.s
		elem.c = e.c
	}
	return elems, nil
}

func (e *Element) Text() (string, error) {
	var str string
	err := e.c.get(e.path("/text"), &str)
	return str, err
}

func (e *Element) Click() error {
	return e.c.post(e.path("/click"), nil, nil)
}

func (e *Element) Submit() error {
	return e.c.post(e.path("/submit"), nil, nil)
}

// TODO: Value, Keys

func (e *Element) Name() (string, error) {
	var str string
	err := e.c.get(e.path("/name"), &str)
	return str, err
}

func (e *Element) Value(str string) error {
	var args struct {
		Value []string `json:"value"`
	}
	args.Value = strings.Split(str, "")
	return e.c.post(e.path("/value"), &args, nil)
}

func (e *Element) Clear() error {
	return e.c.post(e.path("/clear"), nil, nil)
}

func (e *Element) Selected() (bool, error) {
	var b bool
	err := e.c.get(e.path("/selected"), &b)
	return b, err
}

func (e *Element) Enabled() (bool, error) {
	var b bool
	err := e.c.get(e.path("/enabled"), &b)
	return b, err
}

func (e *Element) Attr(name string) (string, bool, error) {
	var ptr *string
	err := e.c.get(e.path("/attribute/"+name), &ptr)
	var str string
	if ptr != nil {
		str = *ptr
	}
	return str, ptr != nil, err
}

func (e *Element) Equal(other *Element) (bool, error) {
	var ok bool
	err := e.c.get(e.path("/equals/"+other.ID), &ok)
	return ok, err
}

func (e *Element) Displayed() (bool, error) {
	var ok bool
	err := e.c.get(e.path("/displayed"), &ok)
	return ok, err
}

func (e *Element) Location() (x, y int, err error) {
	var args struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := e.c.get(e.path("/location"), &args); err != nil {
		return 0, 0, err
	}
	return args.X, args.Y, nil
}

func (e *Element) Size() (x, y int, err error) {
	var args struct {
		X int `json:"width"`
		Y int `json:"height"`
	}
	if err := e.c.get(e.path("/size"), &args); err != nil {
		return 0, 0, err
	}
	return args.X, args.Y, nil
}

func (e *Element) CSS(property string) (string, error) {
	var str string
	if err := e.c.get(e.path("/css/"+property), &str); err != nil {
		return "", err
	}
	return str, nil
}

func (s *Session) Orientation() (string, error) {
	var str json.RawMessage
	if err := s.c.get(s.path("/orientation"), &str); err != nil {
		return "", err
	}
	return string(str), nil
}

func (s *Session) SetOrientation(orientation string) error {
	var args struct {
		O string `json:"orientation"`
	}
	args.O = orientation
	return s.c.post(s.path("/orientation"), &args, nil)
}

func (s *Session) AlertText() (string, error) {
	var str string
	err := s.c.get(s.path("/alert_text"), &str)
	return str, err
}

func (s *Session) AlertOK() error {
	err := s.c.post(s.path("/accept_alert"), nil, nil)
	return err
}

func (s *Session) AlertCancel() error {
	err := s.c.post(s.path("/dismiss_alert"), nil, nil)
	return err
}

// TODO: Mouse and Touch

// TODO Location

// TODO local storage

// TODO session storage

// TODO log

// TODO application cache
