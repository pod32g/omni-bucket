package cli_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pod32g/omni-bucket/internal/cli"
)

func TestPRViewCommand(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/pullrequests/3" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"id":3,"title":"Add feature","state":"OPEN"}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pr", "view", "3", "--repo", "ws/repo"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "#3") || !strings.Contains(buf.String(), "Add feature") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPRApproveCommand(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/repositories/ws/repo/pullrequests/3/approve" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"approved":true}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pr", "approve", "3", "--repo", "ws/repo"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Approved pull request #3") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPRMergeCommand(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/repositories/ws/repo/pullrequests/3/merge" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"id":3,"state":"MERGED"}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pr", "merge", "3", "--repo", "ws/repo"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Merged pull request #3") || !strings.Contains(buf.String(), "MERGED") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPRCreateCommand(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/repositories/ws/repo/pullrequests" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"id":9,"title":"My PR","state":"OPEN"}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pr", "create", "--repo", "ws/repo", "--title", "My PR", "--source", "feature/x", "--destination", "main"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Created pull request #9") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPRApproveCommandSurfacesAPIError(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"type":"error","error":{"message":"You do not have permission"}}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"pr", "approve", "3", "--repo", "ws/repo"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected an error from a 403 response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Fatalf("error = %v", err)
	}
}

func TestPRViewCommandRejectsNonPositiveID(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"pr", "view", "0", "--repo", "ws/repo"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected an error for id 0")
	}
	if !strings.Contains(err.Error(), "positive integer") {
		t.Fatalf("error = %v", err)
	}
}

func TestPRMergeCommandJSON(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":3,"state":"MERGED"}`)
	}))
	defer srv.Close()
	pointFactoryAt(t, srv.URL)

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pr", "merge", "3", "--repo", "ws/repo", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"state": "MERGED"`) {
		t.Fatalf("got %q", buf.String())
	}
}
