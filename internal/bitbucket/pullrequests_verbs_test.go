package bitbucket_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

func TestPullRequestsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/pullrequests/3" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"id":3,"title":"Add feature","state":"OPEN"}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL
	pr, err := c.PullRequests.Get(context.Background(), "ws/repo", 3)
	if err != nil {
		t.Fatal(err)
	}
	if pr.Title != "Add feature" {
		t.Fatalf("title = %q", pr.Title)
	}
}

func TestPullRequestsApprove(t *testing.T) {
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		if r.Header.Get("Content-Type") != "" {
			t.Errorf("approve must not send Content-Type, got %q", r.Header.Get("Content-Type"))
		}
		fmt.Fprint(w, `{"approved":true}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL
	if err := c.PullRequests.Approve(context.Background(), "ws/repo", 3); err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || path != "/repositories/ws/repo/pullrequests/3/approve" {
		t.Fatalf("got %s %s", method, path)
	}
}

func TestPullRequestsMerge(t *testing.T) {
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		fmt.Fprint(w, `{"id":3,"state":"MERGED"}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL
	pr, err := c.PullRequests.Merge(context.Background(), "ws/repo", 3)
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || path != "/repositories/ws/repo/pullrequests/3/merge" {
		t.Fatalf("got %s %s", method, path)
	}
	if pr.State != "MERGED" {
		t.Fatalf("state = %q", pr.State)
	}
}

func TestPullRequestsCreate(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		fmt.Fprint(w, `{"id":9,"title":"My PR","state":"OPEN"}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL
	pr, err := c.PullRequests.Create(context.Background(), "ws/repo", bitbucket.CreatePullRequest{
		Title:       "My PR",
		Source:      "feature/x",
		Destination: "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pr.ID != 9 {
		t.Fatalf("id = %d", pr.ID)
	}
	if gotBody["title"] != "My PR" {
		t.Fatalf("request title = %v", gotBody["title"])
	}
	if _, hasDesc := gotBody["description"]; hasDesc {
		t.Fatal("description key should be absent when Description is empty")
	}
}
