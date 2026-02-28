package provider

import "testing"

func TestForHost(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		configured string
		wantName   string
	}{
		{name: "gitlab.com auto-detected", host: "gitlab.com", wantName: "gitlab"},
		{name: "github.com auto-detected", host: "github.com", wantName: "github"},
		{name: "unknown host falls back to generic", host: "git.example.com", wantName: "generic"},
		{name: "self-hosted gitlab", host: "gitlab.mycompany.com", wantName: "gitlab"},
		{name: "self-hosted github enterprise", host: "github.enterprise.com", wantName: "github"},
		{name: "explicit config overrides detection", host: "git.example.com", configured: "gitlab", wantName: "gitlab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ForHost(tt.host, tt.configured)
			if p.Name() != tt.wantName {
				t.Errorf("ForHost(%q, %q).Name() = %q, want %q", tt.host, tt.configured, p.Name(), tt.wantName)
			}
		})
	}
}

func TestGitLabDefaults(t *testing.T) {
	g := &GitLab{}
	if g.DefaultUsername() != "oauth2" {
		t.Errorf("DefaultUsername() = %q, want %q", g.DefaultUsername(), "oauth2")
	}
}

func TestGitHubDefaults(t *testing.T) {
	g := &GitHub{}
	if g.DefaultUsername() != "x-access-token" {
		t.Errorf("DefaultUsername() = %q, want %q", g.DefaultUsername(), "x-access-token")
	}
}

func TestGenericDefaults(t *testing.T) {
	g := &Generic{}
	if g.DefaultUsername() != "" {
		t.Errorf("DefaultUsername() = %q, want %q", g.DefaultUsername(), "")
	}
}
