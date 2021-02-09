// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloudprint

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"

	"code.google.com/p/goauth2/oauth"
)

const cloudprintURL = "https://www.google.com/cloudprint"

type multipartRequest struct {
	buf bytes.Buffer
	*multipart.Writer
}

func (mr *multipartRequest) init() {
	mr.Writer = multipart.NewWriter(&mr.buf)
}

type jsonStatus struct {
	Success bool
	Message string
}

func jsonRPC(auth *Auth, verb, url string, mr *multipartRequest, dst interface{}) error {
	var reader io.Reader
	if mr != nil {
		if auth.ProxyID != "" {
			mr.WriteField("proxy", auth.ProxyID)
		}
		mr.Writer.Close()
		reader = &mr.buf
	}
	req, err := http.NewRequest(verb, url, reader)
	if err != nil {
		return err
	}
	if i := strings.Index(url, "?"); i >= 0 {
		url = url[:i]
	}
	if mr != nil {
		req.Header.Set("Content-Type", mr.FormDataContentType())
	}
	if auth.ProxyID != "" {
		req.Header.Set("X-CloudPrint-Proxy", auth.ProxyID)
	}

	var client *http.Client
	if auth.Token.AccessToken != "" {
		var tr oauth.Transport
		tr.Config = &oauth.Config{
			ClientId:     auth.APIClientID,
			ClientSecret: auth.APIClientSecret,
			Scope:        "https://www.googleapis.com/auth/cloudprint https://www.googleapis.com/auth/googletalk",
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://accounts.google.com/o/oauth2/token",
			RedirectURL:  "oob",
		}
		tr.Token = &auth.Token
		client = tr.Client()
	} else {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %v", verb, url, err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return fmt.Errorf("%s %s: reading HTTP response: %v", verb, url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: %s\n%s", verb, url, resp.Status, buf.Bytes())
	}

	fmt.Printf("%s\n", buf.Bytes())

	var js jsonStatus
	if err := json.Unmarshal(buf.Bytes(), &js); err != nil {
		return fmt.Errorf("%s %s: invalid JSON response: %v", verb, url, err)
	}
	if !js.Success {
		suffix := ""
		if js.Message != "" {
			msg := js.Message
			if len(msg) > 2 && 'A' <= msg[0] && msg[0] <= 'Z' && 'a' <= msg[1] && msg[1] <= 'z' {
				msg = string(msg[0]+'a'-'A') + msg[1:]
			}
			suffix = ": " + msg
		} else {
			suffix = "\n" + buf.String()
		}
		return fmt.Errorf("%s %s: server rejected request%s", verb, url, suffix)
	}

	if dst != nil {
		if err := json.Unmarshal(buf.Bytes(), dst); err != nil {
			return fmt.Errorf("%s %s: invalid JSON response: %v", verb, url, err)
		}
	}

	return nil
}

func httpGET(auth *Auth, url string) ([]byte, error) {
	verb := "GET"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if i := strings.Index(url, "?"); i >= 0 {
		url = url[:i]
	}
	req.Header.Set("X-CloudPrint-Proxy", auth.ProxyID)

	var client *http.Client
	if auth.Token.AccessToken != "" {
		var tr oauth.Transport
		tr.Config = &oauth.Config{
			ClientId:     auth.APIClientID,
			ClientSecret: auth.APIClientSecret,
			Scope:        "https://www.googleapis.com/auth/cloudprint https://www.googleapis.com/auth/googletalk",
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://accounts.google.com/o/oauth2/token",
			RedirectURL:  "oob",
		}
		tr.Token = &auth.Token
		client = tr.Client()
	} else {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %v", verb, url, err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: reading HTTP response: %v", verb, url, err)
	}

	// TODO: Check 200

	return buf.Bytes(), nil
}

// RandomID returns a new random ID, for use as a proxy ID or printer ID.
func RandomID() (string, error) {
	buf := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("RandomID: reading rand.Reader: %v", err)
	}
	return fmt.Sprintf("%x", buf), nil
}

// AuthUser launches a web browser to ask the user to authenticate
// and records the resulting token.
func AuthUser(auth Auth) (Auth, error) {
	// TODO: Stop using package log, package httptest.
	config := &oauth.Config{
		ClientId:     auth.APIClientID,
		ClientSecret: auth.APIClientSecret,
		Scope:        "https://www.googleapis.com/auth/cloudprint https://www.googleapis.com/auth/googletalk",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
	}
	ch := make(chan string)
	randState, err := RandomID()
	if err != nil {
		return Auth{}, err
	}
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState)
	go openURL(authURL)
	log.Printf("Authorize this app at: %s", authURL)
	code := <-ch
	log.Printf("Got code: %s", code)

	t := &oauth.Transport{
		Config:    config,
		Transport: http.DefaultTransport,
	}
	_, err = t.Exchange(code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}

	auth.Token = *t.Token
	return auth, nil
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.")
}

type JobStatus string

const (
	JobQueued     JobStatus = "QUEUED"
	JobInProgress JobStatus = "IN_PROGRESS"
	JobDone       JobStatus = "DONE"
	JobError      JobStatus = "ERROR"
)

type PrinterStatus string

const (
	PrinterOnline  PrinterStatus = "ONLINE"
	PrinterUnknown PrinterStatus = "UNKNOWN"
	PrinterOffline PrinterStatus = "OFFLINE"
	PrinterDormant PrinterStatus = "DORMANT"
)
