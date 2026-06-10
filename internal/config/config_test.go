package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pod32g/omni-bucket/internal/config"
)

func TestSaveLoadRoundTripWith0600Perms(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	t.Setenv("OMNI_BUCKET_CONFIG", path)

	c := &config.Config{Email: "e@x.com", Token: "tok", DefaultWorkspace: "ws"}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perms = %v, want 0600", info.Mode().Perm())
	}
	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "e@x.com" || got.Token != "tok" || got.DefaultWorkspace != "ws" {
		t.Fatalf("got %+v", got)
	}
}

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	t.Setenv("OMNI_BUCKET_CONFIG", filepath.Join(t.TempDir(), "nope.yml"))
	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "" || got.Token != "" {
		t.Fatalf("expected empty config, got %+v", got)
	}
}

func TestResolvedEnvOverridesFile(t *testing.T) {
	c := &config.Config{Email: "file@x.com", Token: "filetok", DefaultWorkspace: "fws"}
	t.Setenv("BITBUCKET_EMAIL", "env@x.com")
	t.Setenv("BITBUCKET_API_TOKEN", "envtok")
	email, token, ws := c.Resolved()
	if email != "env@x.com" || token != "envtok" || ws != "fws" {
		t.Fatalf("got %s/%s/%s", email, token, ws)
	}
}

func TestResolvedWorkspaceEnvOverride(t *testing.T) {
	c := &config.Config{DefaultWorkspace: "fws"}
	t.Setenv("BITBUCKET_WORKSPACE", "envws")
	_, _, ws := c.Resolved()
	if ws != "envws" {
		t.Fatalf("ws = %q, want envws", ws)
	}
}
