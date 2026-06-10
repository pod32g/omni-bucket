package cli

import (
	"errors"
	"strings"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/config"
)

// newClientFn builds the authenticated API client. It is a package variable so
// tests can substitute a client pointed at an httptest server.
var newClientFn = defaultNewClient

// SetClientFactory overrides the client factory (test helper).
func SetClientFactory(fn func() (*bitbucket.Client, error)) { newClientFn = fn }

// ResetClientFactory restores the default factory (test helper).
func ResetClientFactory() { newClientFn = defaultNewClient }

// newCredClientFn builds a client from explicit credentials (used by `auth
// login`, which has credentials before any config is saved). Overridable in tests.
var newCredClientFn = bitbucket.NewClient

// SetCredClientFactory overrides the credential client factory (test helper).
func SetCredClientFactory(fn func(email, token string) *bitbucket.Client) { newCredClientFn = fn }

// ResetCredClientFactory restores the default credential factory (test helper).
func ResetCredClientFactory() { newCredClientFn = bitbucket.NewClient }

func defaultNewClient() (*bitbucket.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	email, token, _ := cfg.Resolved()
	if email == "" || token == "" {
		return nil, errors.New("not authenticated, run `bb auth login`")
	}
	return bitbucket.NewClient(email, token), nil
}

// resolveRepo turns --repo (and an optional default workspace) into a
// "workspace/repo" slug. A bare repo name is prefixed with the default
// workspace from --workspace or config.
func resolveRepo(cfg *config.Config) (string, error) {
	if flags.repo == "" {
		return "", errors.New("no repository specified; use --repo workspace/repo")
	}
	if strings.Contains(flags.repo, "/") {
		return flags.repo, nil
	}
	// Workspace precedence: --workspace flag > BITBUCKET_WORKSPACE env > config file.
	_, _, ws := cfg.Resolved()
	if flags.workspace != "" {
		ws = flags.workspace
	}
	if ws == "" {
		return "", errors.New("repo has no workspace; use --repo workspace/repo or set a default workspace")
	}
	return ws + "/" + flags.repo, nil
}
