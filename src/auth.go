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
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// The authorization code is received with the loopback IP address flow, see
// https://developers.google.com/identity/protocols/oauth2/native-app#redirect-uri_loopback
// It replaces the out-of-band (OOB) flow that Google has blocked, see
// https://developers.google.com/identity/protocols/oauth2/resources/oob-migration

var (
	errAuthStateMismatch = errors.New("the authorization response doesn't belong to this authorization request")
	errNoAuthCode        = errors.New("the authorization response has no authorization code")
	errNoRefreshToken    = errors.New("no refresh token was issued; revoke drive's access at https://myaccount.google.com/permissions then retry")
)

const authResponsePage = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>drive</title></head>
<body><h3>%s</h3><p>%s</p></body></html>
`

type authCodeResult struct {
	code string
	err  error
	// pasted is set if the result was typed in, instead of
	// being received from the browser's redirect.
	pasted bool
}

// listenOnLoopback listens on a free port of the loopback interface.
func listenOnLoopback() (ln net.Listener, err error) {
	for _, addr := range []string{"127.0.0.1:0", "[::1]:0"} {
		var lErr error
		if ln, lErr = net.Listen("tcp", addr); lErr == nil {
			return ln, nil
		}
		if err == nil {
			err = lErr
		}
	}
	return nil, err
}

func randomURLSafeString(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// pkceChallenge derives the S256 code challenge of RFC 7636 from verifier.
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// authCodeFromQuery extracts the authorization code from the parameters that
// the authorization server appended to the redirect URL.
func authCodeFromQuery(query url.Values, wantState string) (string, error) {
	// The state has to be checked first so that a response that
	// isn't ours can neither complete nor abort the authorization.
	if query.Get("state") != wantState {
		return "", errAuthStateMismatch
	}
	if errCode := query.Get("error"); errCode != "" {
		return "", fmt.Errorf("authorization failed: %s", errCode)
	}
	code := query.Get("code")
	if code == "" {
		return "", errNoAuthCode
	}
	return code, nil
}

// authCodeFromPaste extracts the authorization code from what the user
// pasted: either the URL that the browser got redirected to, or only the
// value of its code parameter.
func authCodeFromPaste(pasted, wantState string) (string, error) {
	pasted = strings.TrimSpace(pasted)

	rawQuery := pasted
	if i := strings.IndexByte(rawQuery, '?'); i >= 0 {
		rawQuery = rawQuery[i+1:]
	}
	if i := strings.IndexByte(rawQuery, '#'); i >= 0 {
		rawQuery = rawQuery[:i]
	}

	if !strings.Contains(rawQuery, "=") {
		// Only the code, which is still escaped if it was copied out of the URL.
		code := pasted
		if unescaped, err := url.PathUnescape(code); err == nil {
			code = unescaped
		}
		if code == "" || strings.Contains(code, "://") {
			return "", errNoAuthCode
		}
		return code, nil
	}

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", errNoAuthCode
	}
	return authCodeFromQuery(query, wantState)
}

// authCallbackHandler handles the request that the browser makes once the
// authorization server redirects it back to the loopback interface.
func authCallbackHandler(wantState string, results chan<- authCodeResult) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		if req.URL.Path != "/" || (query.Get("code") == "" && query.Get("error") == "") {
			// e.g. the browser asking for the favicon.
			http.NotFound(w, req)
			return
		}

		code, err := authCodeFromQuery(query, wantState)
		if err == errAuthStateMismatch {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, authResponsePage, "Authorization failed", html.EscapeString(err.Error()))
		} else {
			fmt.Fprintf(w, authResponsePage, "Authorization received",
				"You can close this tab and return to drive in your terminal.")
		}

		select {
		case results <- authCodeResult{code: code, err: err}:
		default:
			// Already got a result, e.g. on a page reload.
		}
	})
}

// readPastedAuthCode is the fallback for when the browser can't reach the
// loopback interface of this machine e.g. during an SSH session: the browser
// fails to load the page that it got redirected to, but its URL still carries
// the authorization code.
func readPastedAuthCode(in io.Reader, out io.Writer, wantState string, results chan<- authCodeResult) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		code, err := authCodeFromPaste(line, wantState)
		if err == errAuthStateMismatch || err == errNoAuthCode {
			fmt.Fprintf(out, "%v\nPaste the URL that the browser was redirected to: ", err)
			continue
		}

		select {
		case results <- authCodeResult{code: code, err: err, pasted: true}:
		default:
		}
		return
	}
}

// retrieveRefreshToken takes the user through the OAuth 2.0 authorization
// code flow, prompting on out and reading the manually pasted response, if
// any, from in. authConfig's redirect URL gets set to the loopback address
// that is listened on to receive the authorization response.
func retrieveRefreshToken(ctx context.Context, authConfig *oauth2.Config, in io.Reader, out io.Writer) (string, error) {
	ln, err := listenOnLoopback()
	if err != nil {
		return "", fmt.Errorf("cannot listen on the loopback interface to receive the authorization: %v", err)
	}
	defer ln.Close()

	authConfig.RedirectURL = "http://" + ln.Addr().String()

	state, err := randomURLSafeString(32)
	if err != nil {
		return "", err
	}
	// Bind the authorization code to this process as per RFC 7636, because
	// any other local process could get hold of the loopback redirect.
	verifier, err := randomURLSafeString(32)
	if err != nil {
		return "", err
	}

	authURL := authConfig.AuthCodeURL(state, oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))

	results := make(chan authCodeResult, 1)
	server := &http.Server{Handler: authCallbackHandler(state, results)}
	go server.Serve(ln)
	defer server.Close()

	fmt.Fprintf(out, "Visit this URL to authorize drive\n%s\n\n"+
		"Once you approve, the browser gets redirected to %s and drive carries on by itself.\n"+
		"If the browser is on another machine, that page won't load: copy its URL from the\n"+
		"address bar and paste it here: ", authURL, authConfig.RedirectURL)

	flushTTYin()
	go readPastedAuthCode(in, out, state, results)

	var res authCodeResult
	select {
	case res = <-results:
	case <-ctx.Done():
		fmt.Fprintln(out)
		return "", ctx.Err()
	}

	if !res.pasted {
		fmt.Fprintln(out, "\nAuthorization received from the browser.")
	}

	// Let the browser get its response before going away.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	server.Shutdown(shutdownCtx)
	cancel()

	if res.err != nil {
		return "", res.err
	}

	token, err := authConfig.Exchange(ctx, res.code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return "", err
	}
	if token.RefreshToken == "" {
		return "", errNoRefreshToken
	}
	return token.RefreshToken, nil
}
