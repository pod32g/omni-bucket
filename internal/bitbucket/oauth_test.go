package bitbucket

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fixedNow(ts time.Time) func() time.Time { return func() time.Time { return ts } }

func TestExchangeCode(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != "abc" {
			t.Errorf("form = %v", r.Form)
		}
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("id:secret"))
		if r.Header.Get("Authorization") != wantAuth {
			t.Errorf("auth = %q", r.Header.Get("Authorization"))
		}
		fmt.Fprint(w, `{"access_token":"at","refresh_token":"rt","expires_in":7200,"token_type":"bearer"}`)
	}))
	defer srv.Close()
	old := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = old }()

	tok, err := ExchangeCode(context.Background(), srv.Client(), "id", "secret", "abc", "http://localhost:8765/callback", fixedNow(now))
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "at" || tok.RefreshToken != "rt" {
		t.Fatalf("tok = %+v", tok)
	}
	if !tok.Expiry.Equal(now.Add(7200 * time.Second)) {
		t.Fatalf("expiry = %v", tok.Expiry)
	}
}

func TestOAuthTokenSourceNoRefreshWhenFresh(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("token endpoint must not be called when the access token is fresh")
	}))
	defer srv.Close()
	old := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = old }()

	ts := &OAuthTokenSource{
		AccessToken: "fresh", Expiry: now.Add(time.Hour),
		HTTPClient: srv.Client(), now: fixedNow(now),
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.bitbucket.org/2.0/user", nil)
	if err := ts.Authorize(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "Bearer fresh" {
		t.Fatalf("auth = %q", req.Header.Get("Authorization"))
	}
}

func TestOAuthTokenSourceRefreshesWhenExpired(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "old-rt" {
			t.Errorf("form = %v", r.Form)
		}
		fmt.Fprint(w, `{"access_token":"new-at","refresh_token":"new-rt","expires_in":7200,"token_type":"bearer"}`)
	}))
	defer srv.Close()
	old := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = old }()

	var gotAccess, gotRefresh string
	var gotExpiry time.Time
	ts := &OAuthTokenSource{
		ClientID: "id", ClientSecret: "secret",
		AccessToken: "stale", RefreshToken: "old-rt",
		Expiry:     now.Add(-time.Minute), // expired
		HTTPClient: srv.Client(), now: fixedNow(now),
		OnRefresh: func(a, r string, e time.Time) { gotAccess, gotRefresh, gotExpiry = a, r, e },
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.bitbucket.org/2.0/user", nil)
	if err := ts.Authorize(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "Bearer new-at" {
		t.Fatalf("auth = %q", req.Header.Get("Authorization"))
	}
	if gotAccess != "new-at" || gotRefresh != "new-rt" || !gotExpiry.Equal(now.Add(7200*time.Second)) {
		t.Fatalf("OnRefresh got %q/%q/%v", gotAccess, gotRefresh, gotExpiry)
	}
}

func TestOAuthTokenSourceRefreshError(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer srv.Close()
	old := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = old }()

	ts := &OAuthTokenSource{
		RefreshToken: "dead", Expiry: now.Add(-time.Hour),
		HTTPClient: srv.Client(), now: fixedNow(now),
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.bitbucket.org/2.0/user", nil)
	err := ts.Authorize(context.Background(), req)
	if err == nil {
		t.Fatal("expected an error when refresh fails")
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Fatalf("error should mention invalid_grant: %v", err)
	}
}

func TestOAuthTokenSourceRefreshesWithinBuffer(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"refreshed","refresh_token":"rt2","expires_in":7200,"token_type":"bearer"}`)
	}))
	defer srv.Close()
	old := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = old }()

	ts := &OAuthTokenSource{
		ClientID: "id", ClientSecret: "secret",
		AccessToken: "soon-stale", RefreshToken: "rt",
		Expiry:     now.Add(30 * time.Second), // inside the 60s buffer
		HTTPClient: srv.Client(), now: fixedNow(now),
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.bitbucket.org/2.0/user", nil)
	if err := ts.Authorize(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "Bearer refreshed" {
		t.Fatalf("auth = %q", req.Header.Get("Authorization"))
	}
}
