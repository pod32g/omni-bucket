package bitbucket_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

func TestRepositoriesListByWorkspace(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"values":[{"full_name":"ws/api","name":"api","is_private":true}]}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL

	var names []string
	for repo, err := range c.Repositories.List(context.Background(), "ws", bitbucket.RepoListOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, repo.FullName)
	}
	if len(names) != 1 || names[0] != "ws/api" {
		t.Fatalf("got %v", names)
	}
}
