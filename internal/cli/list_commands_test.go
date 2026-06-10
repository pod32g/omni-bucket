package cli_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/cli"
)

// pointFactoryAt makes the CLI build a client aimed at the given test server.
func pointFactoryAt(t *testing.T, url string) {
	t.Helper()
	cli.SetClientFactory(func() (*bitbucket.Client, error) {
		c := bitbucket.NewClient("e", "t")
		c.BaseURL = url
		return c, nil
	})
	t.Cleanup(cli.ResetClientFactory)
}

func TestRepoListCommand(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	t.Setenv("BITBUCKET_WORKSPACE", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/myws" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"values":[{"full_name":"myws/api","name":"api","is_private":true,"mainbranch":{"name":"main"}}]}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"repo", "list", "--workspace", "myws"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "myws/api") || !strings.Contains(out, "private") || !strings.Contains(out, "main") {
		t.Fatalf("got %q", out)
	}
}

func TestRepoListCommandJSON(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	t.Setenv("BITBUCKET_WORKSPACE", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"values":[{"full_name":"myws/api","is_private":false}]}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"repo", "list", "--workspace", "myws", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"full_name": "myws/api"`) {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPipelineListCommand(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/pipelines/" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"values":[{"build_number":42,"state":{"name":"COMPLETED","result":{"name":"SUCCESSFUL"}}}]}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pipeline", "list", "--repo", "ws/repo"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "42") || !strings.Contains(out, "SUCCESSFUL") {
		t.Fatalf("got %q", out)
	}
}

func TestIssueListCommand(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/issues" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"values":[{"id":5,"title":"Crash","state":"new","kind":"bug"}]}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"issue", "list", "--repo", "ws/repo"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "5") || !strings.Contains(out, "Crash") || !strings.Contains(out, "bug") {
		t.Fatalf("got %q", out)
	}
}
