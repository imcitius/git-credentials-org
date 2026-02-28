package provider

import "fmt"

type Generic struct{}

func (g *Generic) Name() string { return "generic" }

func (g *Generic) DefaultUsername() string { return "" }

func (g *Generic) TokenPrompt(namespace string) string {
	return fmt.Sprintf("Enter credentials for %s.\nUsername: ", namespace)
}

func (g *Generic) DetectHost(_ string) bool { return true }
