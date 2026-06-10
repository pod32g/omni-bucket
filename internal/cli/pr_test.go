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
	"github.com/pod32g/omni-bucket/internal/config"
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

func TestAuthLoginNonInteractive(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"username":"bob","display_name":"Bob"}`)
	}))
	defer srv.Close()

	cli.SetCredClientFactory(func(email, token string) *bitbucket.Client {
		c := bitbucket.NewClient(email, token)
		c.BaseURL = srv.URL
		return c
	})
	defer cli.ResetCredClientFactory()

	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "login", "--email", "bob@x.com", "--token", "secret"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Logged in as Bob") {
		t.Fatalf("output = %q", buf.String())
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Email != "bob@x.com" || cfg.Token != "secret" {
		t.Fatalf("config not saved: %+v", cfg)
	}
}

func TestPRListRejectsInvalidState(t *testing.T) {
	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"pr", "list", "--repo", "ws/repo", "--state", "bogus"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid state")
	}
	if !strings.Contains(err.Error(), "unknown state") {
		t.Fatalf("error = %v", err)
	}
}

func TestAuthLogoutClearsCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	t.Setenv("OMNI_BUCKET_CONFIG", path)
	seed := &config.Config{Method: "oauth", DefaultWorkspace: "keepme",
		OAuth: &config.OAuthConfig{ClientID: "id", RefreshToken: "rt"}}
	if err := seed.Save(); err != nil {
		t.Fatal(err)
	}
	root := cli.NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "logout"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != "" || got.OAuth != nil || got.Email != "" || got.Token != "" {
		t.Fatalf("creds not cleared: %+v", got)
	}
	if got.DefaultWorkspace != "keepme" {
		t.Fatalf("workspace should be preserved, got %q", got.DefaultWorkspace)
	}
}

func TestAuthStatusShowsMethod(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "config.yml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"username":"bob","display_name":"Bob"}`)
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
	if !strings.Contains(buf.String(), "Bob") || !strings.Contains(buf.String(), "token") {
		t.Fatalf("got %q", buf.String())
	}
}
