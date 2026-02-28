package provider

// Provider encapsulates host-specific credential behavior.
type Provider interface {
	// Name returns the provider identifier (e.g., "gitlab", "github").
	Name() string

	// DefaultUsername returns the default username for token-based auth
	// on this platform (e.g., "oauth2" for GitLab PATs).
	DefaultUsername() string

	// TokenPrompt returns the user-facing prompt text for entering a token.
	TokenPrompt(namespace string) string

	// DetectHost returns true if this provider should handle the given host.
	DetectHost(host string) bool
}

// ForHost returns the appropriate provider for a given host,
// preferring explicit configuration over auto-detection.
func ForHost(host, configured string) Provider {
	providers := []Provider{
		&GitLab{},
		&GitHub{},
	}

	if configured != "" {
		for _, p := range providers {
			if p.Name() == configured {
				return p
			}
		}
	}

	for _, p := range providers {
		if p.DetectHost(host) {
			return p
		}
	}

	return &Generic{}
}
