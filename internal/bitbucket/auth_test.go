package bitbucket

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuthAuthorizeSetsHeader(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	b := &BasicAuth{Email: "e@x.com", Token: "tok"}
	if err := b.Authorize(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("e@x.com:tok"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthNotSentToForeignPaginationHost(t *testing.T) {
	var gotAuthOnB string
	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthOnB = r.Header.Get("Authorization")
		fmt.Fprint(w, `{"values":[{"id":2}]}`)
	}))
	defer srvB.Close()
	var srvA *httptest.Server
	srvA = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"values":[{"id":1}],"next":%q}`, srvB.URL+"/page2")
	}))
	defer srvA.Close()

	c := NewClient("e@x.com", "tok")
	c.BaseURL = srvA.URL
	for _, err := range c.PullRequests.List(context.Background(), "ws/repo", ListOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
	}
	if gotAuthOnB != "" {
		t.Fatalf("auth leaked to foreign host: %q", gotAuthOnB)
	}
}
