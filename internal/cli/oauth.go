package cli

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	oauthCallbackPort = "8765"
	oauthRedirectURI  = "http://127.0.0.1:8765/callback"
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

// oauthClientFromConfig builds a Bearer client whose token source persists
// rotated tokens back to the config file on refresh.
func oauthClientFromConfig(cfg *config.Config) (*bitbucket.Client, error) {
	if cfg.OAuth == nil || cfg.OAuth.RefreshToken == "" {
		return nil, fmt.Errorf("not authenticated, run `bb auth login --browser`")
	}
	oc := cfg.OAuth
	ts := bitbucket.NewOAuthTokenSource(
		oc.ClientID, oc.ClientSecret, oc.AccessToken, oc.RefreshToken, oc.Expiry, nil,
		func(access, refresh string, expiry time.Time) {
			persistOAuthTokens(access, refresh, expiry)
		},
	)
	return bitbucket.NewClientWithAuth(ts), nil
}

// persistOAuthTokens writes refreshed tokens back to the config file. It is
// best-effort: a failure here only means the next run refreshes again.
func persistOAuthTokens(access, refresh string, expiry time.Time) {
	cfg, err := config.Load()
	if err != nil || cfg.OAuth == nil {
		return
	}
	cfg.OAuth.AccessToken = access
	cfg.OAuth.RefreshToken = refresh
	cfg.OAuth.Expiry = expiry
	_ = cfg.Save()
}

// promptLine prints a label and reads one trimmed line from stdin.
func promptLine(cmd *cobra.Command, label string) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), label)
	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// runBrowserLogin performs the OAuth 2.0 authorization-code flow via a local
// callback server and stores the resulting tokens.
func runBrowserLogin(cmd *cobra.Command, clientID, clientSecret string) error {
	out := cmd.OutOrStdout()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if clientID == "" {
		clientID = os.Getenv("BITBUCKET_OAUTH_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("BITBUCKET_OAUTH_CLIENT_SECRET")
	}
	if clientID == "" && cfg.OAuth != nil {
		clientID = cfg.OAuth.ClientID
	}
	if clientSecret == "" && cfg.OAuth != nil {
		clientSecret = cfg.OAuth.ClientSecret
	}
	if clientID == "" {
		clientID, err = promptLine(cmd, "OAuth consumer Key (client_id): ")
		if err != nil {
			return err
		}
	}
	if clientSecret == "" {
		fmt.Fprint(out, "OAuth consumer Secret (input hidden): ")
		b, perr := term.ReadPassword(int(os.Stdin.Fd()))
		if perr != nil {
			return perr
		}
		fmt.Fprintln(out)
		clientSecret = strings.TrimSpace(string(b))
	}
	if clientID == "" || clientSecret == "" {
		return errors.New("client_id and client_secret are required")
	}

	state, err := randomState()
	if err != nil {
		return err
	}
	resultCh := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", newCallbackHandler(state, resultCh))
	ln, err := net.Listen("tcp", "127.0.0.1:"+oauthCallbackPort)
	if err != nil {
		return fmt.Errorf("cannot start callback server on port %s (already in use?): %w", oauthCallbackPort, err)
	}
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	authURL := buildAuthorizeURL(clientID, oauthRedirectURI, state)
	fmt.Fprintf(out, "Opening your browser to authorize. If it does not open, visit:\n%s\n", authURL)
	_ = openBrowser(authURL)

	var res callbackResult
	select {
	case res = <-resultCh:
	case <-time.After(2 * time.Minute):
		return errors.New("timed out waiting for the authorization callback")
	}
	if res.err != nil {
		return res.err
	}

	tok, err := bitbucket.ExchangeCode(cmd.Context(), nil, clientID, clientSecret, res.code, oauthRedirectURI, time.Now)
	if err != nil {
		return err
	}

	ts := bitbucket.NewOAuthTokenSource(clientID, clientSecret, tok.AccessToken, tok.RefreshToken, tok.Expiry, nil, nil)
	acct, err := bitbucket.NewClientWithAuth(ts).CurrentUser(cmd.Context())
	if err != nil {
		return fmt.Errorf("token verification failed: %w", err)
	}

	cfg.Email = ""
	cfg.Token = ""
	cfg.Method = "oauth"
	cfg.OAuth = &config.OAuthConfig{
		ClientID: clientID, ClientSecret: clientSecret,
		AccessToken: tok.AccessToken, RefreshToken: tok.RefreshToken, Expiry: tok.Expiry,
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "Logged in as %s (%s) via OAuth\n", acct.DisplayName, acct.Username)
	return nil
}
