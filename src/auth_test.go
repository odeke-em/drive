// Copyright 2026 Google Inc. All Rights Reserved.
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
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/odeke-em/drive/config"
)

const (
	testAuthCode     = "4/0AbC-dEf_gHi"
	testRefreshToken = "1//refresh-token"
	testAuthTimeout  = 10 * time.Second
)

// fakeAuthServer stands in for the token endpoint of the authorization server.
type fakeAuthServer struct {
	*httptest.Server

	mu           sync.Mutex
	exchanges    []url.Values
	refreshToken string
}

func newFakeAuthServer(t *testing.T) *fakeAuthServer {
	fs := &fakeAuthServer{refreshToken: testRefreshToken}
	fs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/token" {
			http.NotFound(w, req)
			return
		}
		if err := req.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fs.mu.Lock()
		fs.exchanges = append(fs.exchanges, req.PostForm)
		refreshToken := fs.refreshToken
		fs.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if req.PostForm.Get("code") != testAuthCode {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "invalid_grant"}`)
			return
		}
		fmt.Fprintf(w, `{"access_token": "access-token", "token_type": "Bearer", "expires_in": 3600, "refresh_token": %q}`, refreshToken)
	}))
	t.Cleanup(fs.Close)
	return fs
}

func (fs *fakeAuthServer) config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scopes:       []string{DriveScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:   fs.URL + "/auth",
			TokenURL:  fs.URL + "/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

func (fs *fakeAuthServer) recordedExchanges() []url.Values {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return append([]url.Values(nil), fs.exchanges...)
}

// promptWatcher is the output of the authorization flow, from
// which the authorization URL that the user is told to visit is picked.
type promptWatcher struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	authURL chan *url.URL
	found   bool
}

var authURLRegexp = regexp.MustCompile(`https?://\S+response_type=code\S*`)

func newPromptWatcher() *promptWatcher {
	return &promptWatcher{authURL: make(chan *url.URL, 1)}
}

func (pw *promptWatcher) Write(b []byte) (int, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.buf.Write(b)
	if !pw.found {
		if match := authURLRegexp.Find(pw.buf.Bytes()); match != nil {
			if u, err := url.Parse(string(match)); err == nil {
				pw.found = true
				pw.authURL <- u
			}
		}
	}
	return len(b), nil
}

func (pw *promptWatcher) String() string {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	return pw.buf.String()
}

func (pw *promptWatcher) waitForAuthURL(t *testing.T) *url.URL {
	t.Helper()
	select {
	case u := <-pw.authURL:
		return u
	case <-time.After(testAuthTimeout):
		t.Fatalf("timed out waiting for the authorization URL, output so far:\n%s", pw.String())
		return nil
	}
}

type refreshTokenResult struct {
	refreshToken string
	err          error
}

// authFlow is a running retrieveRefreshToken.
type authFlow struct {
	t       *testing.T
	out     *promptWatcher
	stdin   *io.PipeWriter
	config  *oauth2.Config
	authURL *url.URL
	result  chan refreshTokenResult
}

func startAuthFlow(t *testing.T, ctx context.Context, config *oauth2.Config) *authFlow {
	t.Helper()

	stdinR, stdinW := io.Pipe()
	t.Cleanup(func() { stdinW.Close() })

	af := &authFlow{
		t:      t,
		out:    newPromptWatcher(),
		stdin:  stdinW,
		config: config,
		result: make(chan refreshTokenResult, 1),
	}
	go func() {
		refreshToken, err := retrieveRefreshToken(ctx, config, stdinR, af.out)
		af.result <- refreshTokenResult{refreshToken, err}
	}()

	af.authURL = af.out.waitForAuthURL(t)
	return af
}

func (af *authFlow) redirectURL(params url.Values) string {
	return af.authURL.Query().Get("redirect_uri") + "/?" + params.Encode()
}

// grantParams are the parameters that the authorization
// server appends to the redirect URL after the user's approval.
func (af *authFlow) grantParams() url.Values {
	return url.Values{
		"state": {af.authURL.Query().Get("state")},
		"code":  {testAuthCode},
		"scope": {DriveScope},
	}
}

// browse does what the browser does once redirected by the authorization server.
func (af *authFlow) browse(params url.Values) (statusCode int, body string) {
	af.t.Helper()

	res, err := http.Get(af.redirectURL(params))
	if err != nil {
		af.t.Fatalf("redirect to the loopback interface: %v", err)
	}
	defer res.Body.Close()

	blob, err := io.ReadAll(res.Body)
	if err != nil {
		af.t.Fatalf("reading the response of the loopback interface: %v", err)
	}
	return res.StatusCode, string(blob)
}

func (af *authFlow) paste(line string) {
	af.t.Helper()
	if _, err := io.WriteString(af.stdin, line+"\n"); err != nil {
		af.t.Fatalf("pasting %q: %v", line, err)
	}
}

func (af *authFlow) wait() (string, error) {
	af.t.Helper()
	select {
	case res := <-af.result:
		return res.refreshToken, res.err
	case <-time.After(testAuthTimeout):
		af.t.Fatalf("timed out waiting for the authorization to finish, output so far:\n%s", af.out.String())
		return "", nil
	}
}

func (af *authFlow) stillWaiting() bool {
	select {
	case res := <-af.result:
		af.result <- res
		return false
	case <-time.After(100 * time.Millisecond):
		return true
	}
}

func TestRetrieveRefreshTokenAuthURL(t *testing.T) {
	fs := newFakeAuthServer(t)
	af := startAuthFlow(t, context.Background(), fs.config())

	if got, want := af.authURL.Scheme+"://"+af.authURL.Host+af.authURL.Path, fs.URL+"/auth"; got != want {
		t.Errorf("authorization endpoint: got %q want %q", got, want)
	}

	query := af.authURL.Query()
	wantParams := map[string]string{
		"client_id":             "client-id",
		"response_type":         "code",
		"access_type":           "offline",
		"scope":                 DriveScope,
		"code_challenge_method": "S256",
	}
	for key, want := range wantParams {
		if got := query.Get(key); got != want {
			t.Errorf("%s: got %q want %q", key, got, want)
		}
	}
	if len(query.Get("state")) < 32 {
		t.Errorf("state %q is too short to be unguessable", query.Get("state"))
	}
	if len(query.Get("code_challenge")) != 43 {
		t.Errorf("code_challenge %q isn't a base64url encoded SHA-256", query.Get("code_challenge"))
	}

	redirectURI := query.Get("redirect_uri")
	if strings.Contains(redirectURI, "oob") {
		t.Fatalf("redirect_uri %q is still out-of-band", redirectURI)
	}
	ru, err := url.Parse(redirectURI)
	if err != nil {
		t.Fatalf("redirect_uri %q: %v", redirectURI, err)
	}
	if ru.Scheme != "http" || ru.Path != "" || ru.RawQuery != "" {
		t.Errorf("redirect_uri %q isn't a bare http URL", redirectURI)
	}
	if ip := net.ParseIP(ru.Hostname()); ip == nil || !ip.IsLoopback() {
		t.Errorf("redirect_uri %q doesn't point to a loopback IP address", redirectURI)
	}
	if ru.Port() == "" || ru.Port() == "0" {
		t.Errorf("redirect_uri %q has no usable port", redirectURI)
	}
	if got := af.config.RedirectURL; got != redirectURI {
		t.Errorf("config's redirect URL: got %q want %q", got, redirectURI)
	}
	if !strings.Contains(af.out.String(), redirectURI) {
		t.Errorf("the redirect URL %q isn't mentioned in the prompt:\n%s", redirectURI, af.out.String())
	}

	af.browse(af.grantParams())
	if _, err := af.wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetrieveRefreshTokenFromLoopbackRedirect(t *testing.T) {
	fs := newFakeAuthServer(t)
	af := startAuthFlow(t, context.Background(), fs.config())

	// Requests that aren't the authorization response are of no consequence.
	if statusCode, _ := af.browse(url.Values{}); statusCode != http.StatusNotFound {
		t.Errorf("request without parameters: got status %d want %d", statusCode, http.StatusNotFound)
	}
	res, err := http.Get(af.authURL.Query().Get("redirect_uri") + "/favicon.ico?code=" + testAuthCode)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("request to another path: got status %d want %d", res.StatusCode, http.StatusNotFound)
	}
	if !af.stillWaiting() {
		t.Fatal("the authorization finished without an authorization response")
	}

	statusCode, body := af.browse(af.grantParams())
	if statusCode != http.StatusOK {
		t.Errorf("got status %d want %d", statusCode, http.StatusOK)
	}
	if !strings.Contains(body, "Authorization received") {
		t.Errorf("unexpected page shown in the browser:\n%s", body)
	}

	refreshToken, err := af.wait()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshToken != testRefreshToken {
		t.Errorf("refresh token: got %q want %q", refreshToken, testRefreshToken)
	}
	if !strings.Contains(af.out.String(), "Authorization received from the browser") {
		t.Errorf("the user isn't told that the authorization was received:\n%s", af.out.String())
	}

	exchanges := fs.recordedExchanges()
	if len(exchanges) != 1 {
		t.Fatalf("got %d token exchanges want 1", len(exchanges))
	}
	form := exchanges[0]
	wantForm := map[string]string{
		"grant_type":    "authorization_code",
		"code":          testAuthCode,
		"client_id":     "client-id",
		"client_secret": "client-secret",
		"redirect_uri":  af.authURL.Query().Get("redirect_uri"),
	}
	for key, want := range wantForm {
		if got := form.Get(key); got != want {
			t.Errorf("token exchange %s: got %q want %q", key, got, want)
		}
	}
	if got, want := pkceChallenge(form.Get("code_verifier")), af.authURL.Query().Get("code_challenge"); got != want {
		t.Errorf("code_verifier %q doesn't match the code_challenge: got %q want %q", form.Get("code_verifier"), got, want)
	}

	// The listener mustn't outlive the authorization.
	if res, err := http.Get(af.redirectURL(af.grantParams())); err == nil {
		res.Body.Close()
		t.Errorf("still listening on %s after the authorization", af.authURL.Query().Get("redirect_uri"))
	}
}

func TestRetrieveRefreshTokenStdinClosed(t *testing.T) {
	// e.g. `drive init < /dev/null`: the loopback redirect has to still work.
	fs := newFakeAuthServer(t)
	af := startAuthFlow(t, context.Background(), fs.config())
	af.stdin.Close()

	if !af.stillWaiting() {
		_, err := af.wait()
		t.Fatalf("the authorization finished once stdin was closed, err: %v", err)
	}

	af.browse(af.grantParams())
	refreshToken, err := af.wait()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshToken != testRefreshToken {
		t.Errorf("refresh token: got %q want %q", refreshToken, testRefreshToken)
	}
}

func TestRetrieveRefreshTokenStateMismatch(t *testing.T) {
	fs := newFakeAuthServer(t)
	af := startAuthFlow(t, context.Background(), fs.config())

	forged := []url.Values{
		{"state": {"forged"}, "code": {"forged-code"}},
		{"code": {"forged-code"}},
		// Neither can the authorization be aborted by a third party.
		{"state": {"forged"}, "error": {"access_denied"}},
		{"error": {"access_denied"}},
	}
	for _, params := range forged {
		if statusCode, _ := af.browse(params); statusCode != http.StatusBadRequest {
			t.Errorf("%v: got status %d want %d", params, statusCode, http.StatusBadRequest)
		}
	}
	af.paste(af.redirectURL(forged[0]))

	if !af.stillWaiting() {
		_, err := af.wait()
		t.Fatalf("the authorization finished on a forged response, err: %v", err)
	}
	if exchanges := fs.recordedExchanges(); len(exchanges) != 0 {
		t.Fatalf("a forged code was exchanged: %v", exchanges)
	}

	// The genuine response is still accepted afterwards.
	af.browse(af.grantParams())
	refreshToken, err := af.wait()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshToken != testRefreshToken {
		t.Errorf("refresh token: got %q want %q", refreshToken, testRefreshToken)
	}
	if exchanges := fs.recordedExchanges(); len(exchanges) != 1 || exchanges[0].Get("code") != testAuthCode {
		t.Errorf("unexpected token exchanges: %v", exchanges)
	}
}

func TestRetrieveRefreshTokenPasted(t *testing.T) {
	cases := []struct {
		name  string
		paste func(af *authFlow) string
	}{
		{
			name:  "redirect URL",
			paste: func(af *authFlow) string { return af.redirectURL(af.grantParams()) },
		},
		{
			name: "redirect URL without the slash",
			paste: func(af *authFlow) string {
				return af.authURL.Query().Get("redirect_uri") + "?" + af.grantParams().Encode()
			},
		},
		{
			name:  "query only",
			paste: func(af *authFlow) string { return "?" + af.grantParams().Encode() },
		},
		{
			name:  "code only",
			paste: func(af *authFlow) string { return testAuthCode },
		},
		{
			name:  "escaped code only",
			paste: func(af *authFlow) string { return url.QueryEscape(testAuthCode) },
		},
		{
			name:  "surrounded by spaces",
			paste: func(af *authFlow) string { return "  " + af.redirectURL(af.grantParams()) + " \r" },
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fs := newFakeAuthServer(t)
			af := startAuthFlow(t, context.Background(), fs.config())

			// Stray input mustn't end the authorization.
			af.paste("")
			af.paste("http://127.0.0.1:1/")
			af.paste(tt.paste(af))

			refreshToken, err := af.wait()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if refreshToken != testRefreshToken {
				t.Errorf("refresh token: got %q want %q", refreshToken, testRefreshToken)
			}
			if strings.Contains(af.out.String(), "from the browser") {
				t.Errorf("a pasted response was reported as received from the browser:\n%s", af.out.String())
			}

			exchanges := fs.recordedExchanges()
			if len(exchanges) != 1 {
				t.Fatalf("got %d token exchanges want 1", len(exchanges))
			}
			if got, want := pkceChallenge(exchanges[0].Get("code_verifier")), af.authURL.Query().Get("code_challenge"); got != want {
				t.Errorf("code_verifier doesn't match the code_challenge: got %q want %q", got, want)
			}
		})
	}
}

func TestRetrieveRefreshTokenDenied(t *testing.T) {
	cases := []struct {
		name string
		deny func(af *authFlow, params url.Values)
	}{
		{
			name: "loopback redirect",
			deny: func(af *authFlow, params url.Values) {
				statusCode, body := af.browse(params)
				if statusCode != http.StatusBadRequest {
					t.Errorf("got status %d want %d", statusCode, http.StatusBadRequest)
				}
				if !strings.Contains(body, "Authorization failed") || !strings.Contains(body, "access_denied") {
					t.Errorf("unexpected page shown in the browser:\n%s", body)
				}
			},
		},
		{
			name: "pasted",
			deny: func(af *authFlow, params url.Values) { af.paste(af.redirectURL(params)) },
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fs := newFakeAuthServer(t)
			af := startAuthFlow(t, context.Background(), fs.config())

			tt.deny(af, url.Values{
				"state": {af.authURL.Query().Get("state")},
				"error": {"access_denied"},
			})

			refreshToken, err := af.wait()
			if err == nil || !strings.Contains(err.Error(), "access_denied") {
				t.Fatalf("got error %v, want one mentioning access_denied", err)
			}
			if refreshToken != "" {
				t.Errorf("got refresh token %q on a denied authorization", refreshToken)
			}
			if exchanges := fs.recordedExchanges(); len(exchanges) != 0 {
				t.Errorf("unexpected token exchanges: %v", exchanges)
			}
		})
	}
}

func TestRetrieveRefreshTokenErrorPageIsEscaped(t *testing.T) {
	fs := newFakeAuthServer(t)
	af := startAuthFlow(t, context.Background(), fs.config())

	_, body := af.browse(url.Values{
		"state": {af.authURL.Query().Get("state")},
		"error": {"<script>alert(1)</script>"},
	})
	if strings.Contains(body, "<script>") {
		t.Errorf("the error is reflected unescaped:\n%s", body)
	}
	if _, err := af.wait(); err == nil {
		t.Error("expected an error")
	}
}

func TestRetrieveRefreshTokenExchangeFails(t *testing.T) {
	fs := newFakeAuthServer(t)
	af := startAuthFlow(t, context.Background(), fs.config())

	af.paste("not-the-code-that-was-issued")
	refreshToken, err := af.wait()
	if err == nil || !strings.Contains(err.Error(), "invalid_grant") {
		t.Fatalf("got error %v, want one mentioning invalid_grant", err)
	}
	if refreshToken != "" {
		t.Errorf("got refresh token %q on a failed exchange", refreshToken)
	}
}

func TestRetrieveRefreshTokenNoneIssued(t *testing.T) {
	fs := newFakeAuthServer(t)
	fs.refreshToken = ""
	af := startAuthFlow(t, context.Background(), fs.config())

	af.browse(af.grantParams())
	if _, err := af.wait(); err != errNoRefreshToken {
		t.Fatalf("got error %v want %v", err, errNoRefreshToken)
	}
}

func TestRetrieveRefreshTokenCanceled(t *testing.T) {
	fs := newFakeAuthServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	af := startAuthFlow(t, ctx, fs.config())

	cancel()
	if _, err := af.wait(); err != context.Canceled {
		t.Fatalf("got error %v want %v", err, context.Canceled)
	}
	if res, err := http.Get(af.redirectURL(af.grantParams())); err == nil {
		res.Body.Close()
		t.Errorf("still listening on %s after the cancellation", af.authURL.Query().Get("redirect_uri"))
	}
}

func TestRetrieveRefreshTokenDistinctSecrets(t *testing.T) {
	fs := newFakeAuthServer(t)
	seenStates := make(map[string]bool)
	seenChallenges := make(map[string]bool)

	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		af := startAuthFlow(t, ctx, fs.config())
		query := af.authURL.Query()
		cancel()
		af.wait()

		if seenStates[query.Get("state")] {
			t.Errorf("state %q was reused", query.Get("state"))
		}
		if seenChallenges[query.Get("code_challenge")] {
			t.Errorf("code_challenge %q was reused", query.Get("code_challenge"))
		}
		seenStates[query.Get("state")] = true
		seenChallenges[query.Get("code_challenge")] = true
	}
}

func TestAuthCodeFromPaste(t *testing.T) {
	const state = "the-state"

	cases := []struct {
		pasted   string
		wantCode string
		wantErr  error
	}{
		{pasted: "http://127.0.0.1:4242/?state=the-state&code=4%2F0AbC&scope=x", wantCode: "4/0AbC"},
		{pasted: "http://127.0.0.1:4242?code=4/0AbC&state=the-state", wantCode: "4/0AbC"},
		{pasted: "http://[::1]:4242/?state=the-state&code=4%2F0AbC#fragment", wantCode: "4/0AbC"},
		{pasted: "state=the-state&code=4%2F0AbC", wantCode: "4/0AbC"},
		{pasted: "4/0AbC-dEf_gHi", wantCode: "4/0AbC-dEf_gHi"},
		{pasted: "4%2F0AbC-dEf_gHi", wantCode: "4/0AbC-dEf_gHi"},
		{pasted: "\t4/0AbC \n", wantCode: "4/0AbC"},

		{pasted: "http://127.0.0.1:4242/?state=another-state&code=4%2F0AbC", wantErr: errAuthStateMismatch},
		{pasted: "http://127.0.0.1:4242/?code=4%2F0AbC", wantErr: errAuthStateMismatch},
		{pasted: "code=4%2F0AbC", wantErr: errAuthStateMismatch},
		{pasted: "http://127.0.0.1:4242/?state=the-state", wantErr: errNoAuthCode},
		{pasted: "http://127.0.0.1:4242/?state=the-state&code=", wantErr: errNoAuthCode},
		{pasted: "http://127.0.0.1:4242/", wantErr: errNoAuthCode},
		{pasted: "http://127.0.0.1:4242/?state=the-state&code=%zz", wantErr: errNoAuthCode},
		{pasted: "", wantErr: errNoAuthCode},
		{pasted: "   ", wantErr: errNoAuthCode},
	}

	for _, tt := range cases {
		code, err := authCodeFromPaste(tt.pasted, state)
		if err != tt.wantErr {
			t.Errorf("%q: got error %v want %v", tt.pasted, err, tt.wantErr)
		}
		if code != tt.wantCode {
			t.Errorf("%q: got code %q want %q", tt.pasted, code, tt.wantCode)
		}
	}

	_, err := authCodeFromPaste("http://127.0.0.1:4242/?state=the-state&error=access_denied", state)
	if err == nil || !strings.Contains(err.Error(), "access_denied") {
		t.Errorf("got error %v, want one mentioning access_denied", err)
	}
}

func TestPKCEChallenge(t *testing.T) {
	// The example of RFC 7636, Appendix B.
	const (
		verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		want     = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	)
	if got := pkceChallenge(verifier); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestNewAuthConfigIsNotOutOfBand(t *testing.T) {
	authConfig := newAuthConfig(&config.Context{ClientId: "client-id", ClientSecret: "client-secret"})
	if authConfig.RedirectURL != "" {
		t.Errorf("got a preset redirect URL %q, want it left to the loopback listener", authConfig.RedirectURL)
	}
	if authConfig.ClientID != "client-id" || authConfig.ClientSecret != "client-secret" {
		t.Errorf("unexpected client credentials: %q %q", authConfig.ClientID, authConfig.ClientSecret)
	}
}

func TestInitRequiresOwnAPIKeys(t *testing.T) {
	cases := []struct{ clientId, clientSecret string }{
		{"", ""},
		{"client-id", ""},
		{"", "client-secret"},
	}

	for _, tt := range cases {
		t.Setenv(GoogleApiClientIdEnvKey, tt.clientId)
		t.Setenv(GoogleApiClientSecretEnvKey, tt.clientSecret)

		g := &Commands{context: &config.Context{}}
		if err := g.Init(); err != ErrNoAPIKeys {
			t.Errorf("%+v: got error %v want %v", tt, err, ErrNoAPIKeys)
		}
		if g.context.RefreshToken != "" {
			t.Errorf("%+v: got an unexpected refresh token", tt)
		}
	}
}
