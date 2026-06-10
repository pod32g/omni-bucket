package bitbucket_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

func TestPullRequestsListFollowsNext(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repositories/ws/repo/pullrequests" {
			fmt.Fprintf(w, `{"values":[{"id":1,"title":"first","state":"OPEN"}],"next":%q}`, srv.URL+"/page2")
			return
		}
		fmt.Fprint(w, `{"values":[{"id":2,"title":"second","state":"OPEN"}]}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e@x.com", "tok")
	c.BaseURL = srv.URL

	var ids []int
	for pr, err := range c.PullRequests.List(context.Background(), "ws/repo", bitbucket.ListOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, pr.ID)
	}
	if !reflect.DeepEqual(ids, []int{1, 2}) {
		t.Fatalf("ids = %v, want [1 2]", ids)
	}
}

func TestPullRequestsListRespectsLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"values":[{"id":1},{"id":2},{"id":3}]}`)
	}))
	defer srv.Close()

	c := bitbucket.NewClient("e", "t")
	c.BaseURL = srv.URL

	var count int
	for _, err := range c.PullRequests.List(context.Background(), "ws/repo", bitbucket.ListOptions{Limit: 2}) {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}
