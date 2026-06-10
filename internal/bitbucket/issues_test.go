package bitbucket_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

func TestIssuesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/issues" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"values":[{"id":5,"title":"Crash","state":"new","kind":"bug"}]}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL

	var got int
	for is, err := range c.Issues.List(context.Background(), "ws/repo", bitbucket.IssueListOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		got = is.ID
	}
	if got != 5 {
		t.Fatalf("id = %d, want 5", got)
	}
}
