package bitbucket_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

func TestPipelinesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/pipelines/" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("sort") != "-created_on" {
			t.Errorf("sort = %q, want -created_on", r.URL.Query().Get("sort"))
		}
		fmt.Fprint(w, `{"values":[{"build_number":42,"state":{"name":"COMPLETED"}}]}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL

	var got int
	for p, err := range c.Pipelines.List(context.Background(), "ws/repo", bitbucket.PipelineListOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		got = p.BuildNumber
	}
	if got != 42 {
		t.Fatalf("build number = %d, want 42", got)
	}
}
