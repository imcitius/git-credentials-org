package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Defaults DefaultsConfig            `toml:"defaults"`
	Hosts    map[string]HostConfig     `toml:"hosts"`
	Backends map[string]BackendConfig  `toml:"backends"`
}

type DefaultsConfig struct {
	Backend  string `toml:"backend"`
	LogLevel string `toml:"log_level"`
}

type HostConfig struct {
	Provider string `toml:"provider"`
	Backend  string `toml:"backend"`
}

type BackendConfig struct {
	Vault   string `toml:"vault"`
	Account string `toml:"account"`
}

func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "git-credentials-org", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "git-credentials-org", "config.toml")
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Defaults: DefaultsConfig{
			Backend:  "keychain",
			LogLevel: "warn",
		},
		Hosts:    make(map[string]HostConfig),
		Backends: make(map[string]BackendConfig),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}

// BackendForHost returns the backend name to use for a given host,
// falling back to the default backend.
func (c *Config) BackendForHost(host string) string {
	if hc, ok := c.Hosts[host]; ok && hc.Backend != "" {
		return hc.Backend
	}
	return c.Defaults.Backend
}

// ProviderForHost returns the provider name for a given host.
// Returns empty string if no provider is explicitly configured.
func (c *Config) ProviderForHost(host string) string {
	if hc, ok := c.Hosts[host]; ok {
		return hc.Provider
	}
	return ""
}
