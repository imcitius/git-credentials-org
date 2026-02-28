package provider

import (
	"fmt"
	"strings"
)

type GitLab struct{}

func (g *GitLab) Name() string { return "gitlab" }

func (g *GitLab) DefaultUsername() string { return "oauth2" }

func (g *GitLab) TokenPrompt(namespace string) string {
	return fmt.Sprintf("Enter GitLab Personal Access Token for %s: ", namespace)
}

func (g *GitLab) DetectHost(host string) bool {
	return host == "gitlab.com" || strings.Contains(host, "gitlab")
}
