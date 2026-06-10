package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

const (
	oauthCallbackPort = "8765"
	oauthRedirectURI  = "http://localhost:8765/callback"
)

// buildAuthorizeURL builds the Bitbucket authorize URL for the code grant.
func buildAuthorizeURL(clientID, redirectURI, state string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("state", state)
	if redirectURI != "" {
		q.Set("redirect_uri", redirectURI)
	}
	return bitbucket.AuthorizeEndpoint + "?" + q.Encode()
}

// randomState returns a hex CSRF state token.
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// callbackResult is delivered once by the callback handler.
type callbackResult struct {
	code string
	err  error
}

// newCallbackHandler validates state, extracts the code, writes a closeable
// page, and delivers the result (non-blocking) on resultCh.
func newCallbackHandler(wantState string, resultCh chan<- callbackResult) http.HandlerFunc {
	send := func(r callbackResult) {
		select {
		case resultCh <- r:
		default:
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			http.Error(w, "authorization failed: "+e, http.StatusBadRequest)
			send(callbackResult{err: fmt.Errorf("authorization denied: %s", e)})
			return
		}
		if q.Get("state") != wantState {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			send(callbackResult{err: fmt.Errorf("state mismatch (possible CSRF)")})
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			send(callbackResult{err: fmt.Errorf("no authorization code in callback")})
			return
		}
		fmt.Fprintln(w, "Login successful. You can close this tab and return to the terminal.")
		send(callbackResult{code: code})
	}
}

// openBrowser attempts to open url in the default browser.
func openBrowser(target string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{target}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", target}
	default:
		cmd, args = "xdg-open", []string{target}
	}
	return exec.Command(cmd, args...).Start()
}
