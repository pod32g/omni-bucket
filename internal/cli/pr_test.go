package cli_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/cli"
)

func TestPRListCommandRendersTable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"values":[{"id":7,"title":"Fix bug","state":"OPEN"}]}`)
	}))
	defer srv.Close()

	cli.SetClientFactory(func() (*bitbucket.Client, error) {
		c := bitbucket.NewClient("e", "t")
		c.BaseURL = srv.URL
		return c, nil
	})
	defer cli.ResetClientFactory()

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pr", "list", "--repo", "ws/repo"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Fix bug") || !strings.Contains(out, "7") {
		t.Fatalf("got %q", out)
	}
}

func TestPRListCommandJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"values":[{"id":7,"title":"Fix bug","state":"OPEN"}]}`)
	}))
	defer srv.Close()

	cli.SetClientFactory(func() (*bitbucket.Client, error) {
		c := bitbucket.NewClient("e", "t")
		c.BaseURL = srv.URL
		return c, nil
	})
	defer cli.ResetClientFactory()

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pr", "list", "--repo", "ws/repo", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"title": "Fix bug"`) {
		t.Fatalf("got %q", buf.String())
	}
}

func TestAuthStatusCommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"username":"bob","display_name":"Bob Jones"}`)
	}))
	defer srv.Close()

	cli.SetClientFactory(func() (*bitbucket.Client, error) {
		c := bitbucket.NewClient("e", "t")
		c.BaseURL = srv.URL
		return c, nil
	})
	defer cli.ResetClientFactory()

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "status"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Bob Jones") {
		t.Fatalf("got %q", buf.String())
	}
}
