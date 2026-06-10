package cli

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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
