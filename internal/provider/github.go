package provider

import (
	"fmt"
	"strings"
)

type GitHub struct{}

func (g *GitHub) Name() string { return "github" }

func (g *GitHub) DefaultUsername() string { return "x-access-token" }

func (g *GitHub) TokenPrompt(namespace string) string {
	return fmt.Sprintf("Enter GitHub Personal Access Token for %s: ", namespace)
}

func (g *GitHub) DetectHost(host string) bool {
	return host == "github.com" || strings.Contains(host, "github")
}
