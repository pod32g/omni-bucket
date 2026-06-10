package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds persisted CLI settings.
type Config struct {
	Email            string `yaml:"email"`
	Token            string `yaml:"token"`
	DefaultWorkspace string `yaml:"default_workspace,omitempty"`
}

// Path returns the config file location. OMNI_BUCKET_CONFIG overrides it
// (used in tests); otherwise it is <user-config-dir>/omni-bucket/config.yml.
func Path() (string, error) {
	if p := os.Getenv("OMNI_BUCKET_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "omni-bucket", "config.yml"), nil
}

// Load reads config from disk. A missing file yields an empty Config, no error.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Save writes the config to disk atomically with 0600 permissions.
func (c *Config) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".config-*.yml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, p)
}

// Resolved merges credentials, with env vars taking precedence over the file.
func (c *Config) Resolved() (email, token, workspace string) {
	email, token, workspace = c.Email, c.Token, c.DefaultWorkspace
	if v := os.Getenv("BITBUCKET_EMAIL"); v != "" {
		email = v
	}
	if v := os.Getenv("BITBUCKET_API_TOKEN"); v != "" {
		token = v
	}
	if v := os.Getenv("BITBUCKET_WORKSPACE"); v != "" {
		workspace = v
	}
	return email, token, workspace
}
