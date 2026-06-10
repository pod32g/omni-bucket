package bitbucket_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

func TestParseAPIErrorEnvelope(t *testing.T) {
	body := []byte(`{"type":"error","error":{"message":"Not found","detail":"no such repo"}}`)
	err := bitbucket.ParseAPIErrorForTest(404, body)
	var apiErr *bitbucket.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 || apiErr.Message != "Not found" || apiErr.Detail != "no such repo" {
		t.Fatalf("got %+v", apiErr)
	}
}

func TestParseAPIErrorFallback(t *testing.T) {
	err := bitbucket.ParseAPIErrorForTest(500, []byte("boom"))
	var apiErr *bitbucket.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 || apiErr.Message != "boom" {
		t.Fatalf("got %+v", apiErr)
	}
}

func TestClientSendsBasicAuth(t *testing.T) {
	var gotAuth, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		fmt.Fprint(w, `{"username":"bob","display_name":"Bob"}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("bob@x.com", "secret")
	c.BaseURL = srv.URL
	acct, err := c.CurrentUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("bob@x.com:secret"))
	if gotAuth != want {
		t.Fatalf("Authorization = %q, want %q", gotAuth, want)
	}
	if gotAccept != "application/json" {
		t.Fatalf("Accept = %q", gotAccept)
	}
	if acct.DisplayName != "Bob" {
		t.Fatalf("display name = %q", acct.DisplayName)
	}
}

func TestDoReturnsAPIErrorOnHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"type":"error","error":{"message":"Bad token"}}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL
	_, err := c.CurrentUser(context.Background())
	var apiErr *bitbucket.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != 401 || apiErr.Message != "Bad token" {
		t.Fatalf("got %+v", apiErr)
	}
}
