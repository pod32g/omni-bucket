package cli

import (
	"testing"

	"github.com/pod32g/omni-bucket/internal/config"
)

func TestResolveRepo(t *testing.T) {
	t.Setenv("BITBUCKET_WORKSPACE", "")
	tests := []struct {
		name      string
		repo      string
		flagWS    string
		cfg       *config.Config
		want      string
		wantError bool
	}{
		{name: "empty repo errors", repo: "", cfg: &config.Config{}, wantError: true},
		{name: "full slug passes through", repo: "ws/repo", cfg: &config.Config{}, want: "ws/repo"},
		{name: "bare name uses flag workspace", repo: "repo", flagWS: "fw", cfg: &config.Config{}, want: "fw/repo"},
		{name: "bare name uses config workspace", repo: "repo", cfg: &config.Config{DefaultWorkspace: "cw"}, want: "cw/repo"},
		{name: "flag beats config", repo: "repo", flagWS: "fw", cfg: &config.Config{DefaultWorkspace: "cw"}, want: "fw/repo"},
		{name: "bare name no workspace errors", repo: "repo", cfg: &config.Config{}, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags.repo = tt.repo
			flags.workspace = tt.flagWS
			defer func() { flags.repo = ""; flags.workspace = "" }()
			got, err := resolveRepo(tt.cfg)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
