// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdriver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var conn *Conn
var localAddr string

func init() {
	var err error
	conn, err = Dial("localhost:9515")
	if err != nil {
		panic(err)
	}

	srv := httptest.NewServer(http.DefaultServeMux)
	localAddr = srv.URL
}

func TestStatus(t *testing.T) {
	st, err := conn.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", *st)
}

func TestInvalidSession(t *testing.T) {
	err := conn.get("/session/not-a-session-id", nil)
	if err == nil {
		t.Fatalf("GET /session/not-a-session-id succeeded")
	}
	t.Logf("%s", err)
}

func TestNotFound(t *testing.T) {
	err := conn.get("/not-found", nil)
	if err == nil {
		t.Fatalf("GET /not-found succeeded")
	}
}

func TestNewSession(t *testing.T) {
	s, err := conn.NewSession(nil)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	t.Logf("ID: %s Cap: %+v", s.ID, s.Capabilities)

	/* Fails on chromedriver.
	sess, err := conn.Sessions()
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}

	if len(sess) < 1 {
		t.Fatalf("Sessions returned no sessions")
	}

	for _, s := range sess {
		_ = s
	}
	*/

	/* chromedriver responds with capabilities.
	o, err := s.Orientation()
	if err != nil {
		t.Fatalf("Orientation: %v", err)
	}
	t.Logf("orientation: %v", o)
	*/

	if err := s.Delete(); err != nil {
		t.Fatalf("Session.Delete: %v", err)
	}
}

func TestTimeouts(t *testing.T) {
	s, err := conn.NewSession(nil)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer s.Delete()

	var timeouts = []Timeout{
		// chromedriver rejects these
		//ScriptTimeout,
		//ImplicitTimeout,
		//PageLoadTimeout,
		AsyncScriptTimeout,
		ImplicitWaitTimeout,
	}

	for _, timeout := range timeouts {
		if err := s.SetTimeout(timeout, 1*time.Second); err != nil {
			t.Errorf("set %s timeout: %v", timeout, err)
		}
	}
}

func TestWindow(t *testing.T) {
	s, err := conn.NewSession(nil)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer s.Delete()

	w, err := s.Window()
	if err != nil {
		t.Fatalf("Window: %v", err)
	}
	t.Logf("ID: %v", w.ID)

	ws, err := s.Windows()
	if err != nil {
		t.Fatalf("Windows: %v", err)
	}

	found := false
	for _, w1 := range ws {
		t.Logf("ID: %v", w1.ID)
		if w1.ID == w.ID {
			found = true
		}
	}

	if !found {
		t.Fatalf("did not find current window in window list")
	}

	err = w.Resize(400, 300)
	if err != nil {
		t.Errorf("Resize: %v", err)
	}

	dx, dy, err := w.Size()
	if dx != 400 || dy != 300 || err != nil {
		t.Errorf("Size() = %d, %d, %v, want %d, %d, %v", dx, dy, err, 400, 300, nil)
	}

	x, y, err := w.Position()
	if err != nil {
		t.Errorf("Position: %v", err)
	}
	t.Logf("Position = %d, %d", x, y)

	err = w.Maximize()
	if err != nil {
		t.Errorf("Maximize: %v", err)
	}
	x, y, _ = w.Position()
	t.Logf("Position after Maximize = %d, %d", x, y)

	err = s.SetWindow(w)
	if err != nil {
		t.Errorf("SetWindow: %v", err)
	}

	err = s.CloseWindow()
	if err != nil {
		t.Errorf("CloseWindow: %v", err)
	}
}

var test2Body = []byte(`<html><head><title>Title here</title></head><body><a href="/test3">click me</a></body></html>`)

var test3Body = []byte(`<a href="javascript:alert('alert text')">alert me</a>`)

func init() {
	http.HandleFunc("/test2", func(w http.ResponseWriter, req *http.Request) {
		w.Write(test2Body)
	})
	http.HandleFunc("/test3", func(w http.ResponseWriter, req *http.Request) {
		w.Write(test3Body)
	})
}

func TestURL(t *testing.T) {
	s, err := conn.NewSession(nil)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer s.Delete()

	nav := localAddr + "/test1"
	if err := s.SetURL(nav); err != nil {
		t.Fatalf("SetURL: %v", err)
	}
	url, err := s.URL()
	if url != nav || err != nil {
		t.Fatalf("URL = %q, %v, want %q, %v", url, err, nav, nil)
	}

	nav2 := localAddr + "/test2"
	if err := s.SetURL(nav2); err != nil {
		t.Fatalf("second SetURL: %v", err)
	}
	url, err = s.URL()
	if url != nav2 || err != nil {
		t.Fatalf("after second SetURL, URL = %q, %v, want %q, %v", url, err, nav2, nil)
	}

	if err := s.Back(); err != nil {
		t.Fatalf("Back: %v", err)
	}
	url, err = s.URL()
	if url != nav || err != nil {
		t.Fatalf("after Back, URL = %q, %v, want %q, %v", url, err, nav, nil)
	}

	if err := s.Forward(); err != nil {
		t.Fatalf("Back: %v", err)
	}
	url, err = s.URL()
	if url != nav2 || err != nil {
		t.Fatalf("after Forward, URL = %q, %v, want %q, %v", url, err, nav2, nil)
	}

	if err := s.Refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	title, err := s.Title()
	if title != "Title here" || err != nil {
		t.Fatalf("Title = %q, %v, want %q, %v", title, err, "Title here", nil)
	}

	src, err := s.Source()
	if src != string(test2Body) || err != nil {
		t.Fatalf("Source = %q, %v, want %q, %v", src, err, test2Body, nil)
	}

	elem, err := s.Element(ByTagName, "body")
	if err != nil {
		t.Fatalf("by tag name: %v", err)
	}
	t.Logf("elem %s", elem.ID)

	elems, err := s.Elements(ByTagName, "body")
	if err != nil {
		t.Fatalf("by tag name: %v", err)
	}
	for i, elem := range elems {
		t.Logf("elem#%d %s", i, elem.ID)
	}

	/*
		elem, err = s.ActiveElement()
		if err != nil {
			t.Fatalf("ActiveElement: %v", err)
		}
		t.Logf("active %s", elem.ID)
	*/

	all, err := s.Element(ByTagName, "html")
	if err != nil {
		t.Fatalf("by tag name html: %v", err)
	}
	t.Logf("all %s", all.ID)

	elem, err = all.Element(ByTagName, "body")
	if err != nil {
		t.Fatalf("all by tag name: %v", err)
	}
	t.Logf("all elem %s", elem.ID)

	elems, err = all.Elements(ByTagName, "body")
	if err != nil {
		t.Fatalf("all by tag name: %v", err)
	}
	for i, elem := range elems {
		t.Logf("all elem#%d %s", i, elem.ID)
	}

	link, err := s.Element(ByLinkText, "click me")
	if err != nil {
		t.Fatalf("element by link text: %v", err)
	}

	text, err := link.Text()
	if text != "click me" || err != nil {
		t.Fatalf("link.Text() = %q, %v, want %q, %v", text, err, "click me", nil)
	}

	name, err := link.Name()
	if name != "a" || err != nil {
		t.Fatalf("link.Name() = %q, %v, want %q, %v", name, err, "a", nil)
	}

	enab, err := link.Enabled()
	if enab != true || err != nil {
		t.Fatalf("link.Enabled() = %v, %v, want %v, %v", enab, err, true, nil)
	}

	nav3 := localAddr + "/test3"
	href, ok, err := link.Attr("href")
	if href != nav3 || !ok || err != nil {
		t.Fatalf("link.Attr(\"href\") = %q, %v, %v, want %q, %v, %v", href, ok, err, nav3, true, nil)
	}

	str, ok, err := link.Attr("foo")
	if str != "" || ok || err != nil {
		t.Fatalf("link.Attr(\"foo\") = %q, %v, %v, want %q, %v, %v", str, ok, err, "", false, nil)
	}

	link1, err := s.Element(ByLinkText, "click me")
	if err != nil {
		t.Fatalf("element by link text1: %v", err)
	}

	t.Logf("link %s link1 %s", link.ID, link1.ID)

	ok, err = link.Equal(link1)
	if ok != true || err != nil {
		t.Fatalf("link.Equal(link1) = %v, %v want %v, %v", ok, err, true, nil)
	}

	ok, err = link.Displayed()
	if ok != true || err != nil {
		t.Fatalf("link.Displayed() = %v, %v want %v, %v", ok, err, true, nil)
	}

	x, y, err := link.Location()
	if x == 0 || y == 0 || err != nil {
		t.Fatalf("link.Location() = %d, %d, %v want >0, >0, nil", x, y, err)
	}
	t.Logf("link at %d, %d", x, y)

	x, y, err = link.Size()
	if x == 0 || y == 0 || err != nil {
		t.Fatalf("link.Size() = %d, %d, %v want >0, >0, nil", x, y, err)
	}
	t.Logf("link size %d, %d", x, y)

	color, err := link.CSS("color")
	if color == "" || err != nil {
		t.Fatalf("link.Color() = %q, %v want non-empty, %v", color, err, nil)
	}
	t.Logf("color %q", color)

	if err := link.Click(); err != nil {
		t.Fatalf("click link: %v", err)
	}

	url, err = s.URL()
	if url != nav3 || err != nil {
		t.Fatalf("after Click, URL = %q, %v, want %q, %v", url, err, nav3, nil)
	}

	link, err = s.Element(ByLinkText, "alert me")
	if err != nil {
		t.Fatalf("element by link alert text: %v", err)
	}

	text, err = s.AlertText()
	if text != "" || err == nil {
		t.Fatalf("AlertText() = %q, %v", text, err)
	}

	if err := link.Click(); err != nil {
		t.Fatalf("click link2: %v", err)
	}

	text, err = s.AlertText()
	if text != "alert text" || err != nil {
		t.Fatalf("AlertText() = %q, %v", text, err)
	}

	if err := s.AlertOK(); err != nil {
		t.Fatalf("alert ok: %v", err)
	}

	text, err = s.AlertText()
	if text != "" || err == nil {
		t.Fatalf("AlertText() = %q, %v", text, err)
	}

	// TODO: test Submit
}
