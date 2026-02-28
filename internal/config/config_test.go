package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Load() should not error on missing file, got: %v", err)
	}

	if cfg.Defaults.Backend != "keychain" {
		t.Errorf("default backend = %q, want %q", cfg.Defaults.Backend, "keychain")
	}
}

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[defaults]
backend = "onepassword"
log_level = "debug"

[hosts."gitlab.com"]
provider = "gitlab"
backend = "keychain"

[hosts."github.com"]
provider = "github"

[backends.onepassword]
vault = "DevVault"
account = "team.1password.com"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Defaults.Backend != "onepassword" {
		t.Errorf("defaults.backend = %q, want %q", cfg.Defaults.Backend, "onepassword")
	}

	if cfg.Defaults.LogLevel != "debug" {
		t.Errorf("defaults.log_level = %q, want %q", cfg.Defaults.LogLevel, "debug")
	}

	if hc, ok := cfg.Hosts["gitlab.com"]; !ok {
		t.Error("missing hosts.gitlab.com")
	} else {
		if hc.Provider != "gitlab" {
			t.Errorf("gitlab.com provider = %q, want %q", hc.Provider, "gitlab")
		}
		if hc.Backend != "keychain" {
			t.Errorf("gitlab.com backend = %q, want %q", hc.Backend, "keychain")
		}
	}

	if bc, ok := cfg.Backends["onepassword"]; !ok {
		t.Error("missing backends.onepassword")
	} else {
		if bc.Vault != "DevVault" {
			t.Errorf("onepassword vault = %q, want %q", bc.Vault, "DevVault")
		}
	}
}

func TestBackendForHost(t *testing.T) {
	cfg := &Config{
		Defaults: DefaultsConfig{Backend: "keychain"},
		Hosts: map[string]HostConfig{
			"gitlab.com": {Backend: "onepassword", Provider: "gitlab"},
		},
	}

	if got := cfg.BackendForHost("gitlab.com"); got != "onepassword" {
		t.Errorf("BackendForHost(gitlab.com) = %q, want %q", got, "onepassword")
	}

	if got := cfg.BackendForHost("github.com"); got != "keychain" {
		t.Errorf("BackendForHost(github.com) = %q, want %q", got, "keychain")
	}
}

func TestProviderForHost(t *testing.T) {
	cfg := &Config{
		Hosts: map[string]HostConfig{
			"gitlab.com": {Provider: "gitlab"},
		},
	}

	if got := cfg.ProviderForHost("gitlab.com"); got != "gitlab" {
		t.Errorf("ProviderForHost(gitlab.com) = %q, want %q", got, "gitlab")
	}

	if got := cfg.ProviderForHost("unknown.com"); got != "" {
		t.Errorf("ProviderForHost(unknown.com) = %q, want %q", got, "")
	}
}
