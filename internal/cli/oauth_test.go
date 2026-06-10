package cli

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pod32g/omni-bucket/internal/config"
)

func TestBuildAuthorizeURL(t *testing.T) {
	raw := buildAuthorizeURL("my-id", "http://localhost:8765/callback", "xyz")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("client_id") != "my-id" || q.Get("response_type") != "code" ||
		q.Get("state") != "xyz" || q.Get("redirect_uri") != "http://localhost:8765/callback" {
		t.Fatalf("query = %v", q)
	}
	if !strings.HasPrefix(raw, "https://bitbucket.org/site/oauth2/authorize?") {
		t.Fatalf("url = %q", raw)
	}
}

func TestRandomStateUnique(t *testing.T) {
	a, err := randomState()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := randomState()
	if len(a) != 32 || a == b {
		t.Fatalf("a=%q b=%q", a, b)
	}
}

func TestCallbackHandlerSuccess(t *testing.T) {
	ch := make(chan callbackResult, 1)
	h := newCallbackHandler("st", ch)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/callback?state=st&code=abc", nil)
	h(rec, req)
	if !strings.Contains(rec.Body.String(), "successful") {
		t.Fatalf("body = %q", rec.Body.String())
	}
	res := <-ch
	if res.err != nil || res.code != "abc" {
		t.Fatalf("res = %+v", res)
	}
}

func TestCallbackHandlerStateMismatch(t *testing.T) {
	ch := make(chan callbackResult, 1)
	h := newCallbackHandler("st", ch)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/callback?state=wrong&code=abc", nil)
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d", rec.Code)
	}
	res := <-ch
	if res.err == nil {
		t.Fatal("expected an error on state mismatch")
	}
}

func TestPersistOAuthTokens(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	seed := &config.Config{
		Method: "oauth",
		OAuth: &config.OAuthConfig{
			ClientID: "id", ClientSecret: "secret",
			AccessToken: "old-at", RefreshToken: "old-rt",
			Expiry: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		},
	}
	if err := seed.Save(); err != nil {
		t.Fatal(err)
	}
	newExp := time.Date(2026, 6, 10, 14, 0, 0, 0, time.UTC)
	persistOAuthTokens("new-at", "new-rt", newExp)

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.OAuth.AccessToken != "new-at" || got.OAuth.RefreshToken != "new-rt" || !got.OAuth.Expiry.Equal(newExp) {
		t.Fatalf("oauth = %+v", got.OAuth)
	}
	if got.OAuth.ClientID != "id" || got.OAuth.ClientSecret != "secret" {
		t.Fatalf("consumer creds clobbered: %+v", got.OAuth)
	}
}

func TestOAuthClientFromConfigRequiresRefreshToken(t *testing.T) {
	_, err := oauthClientFromConfig(&config.Config{Method: "oauth"})
	if err == nil {
		t.Fatal("expected error when OAuth block is missing")
	}
}

func TestCallbackHandlerOAuthError(t *testing.T) {
	ch := make(chan callbackResult, 1)
	h := newCallbackHandler("st", ch)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/callback?error=access_denied", nil)
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d", rec.Code)
	}
	res := <-ch
	if res.err == nil {
		t.Fatal("expected an error for the error= callback")
	}
}
