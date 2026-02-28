package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/imcitius/git-credentials-org/internal/config"
	"github.com/imcitius/git-credentials-org/internal/handler"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	verbose := os.Getenv("GIT_CREDENTIALS_ORG_DEBUG") == "1"
	configPath := os.Getenv("GIT_CREDENTIALS_ORG_CONFIG")
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	// Check for --verbose flag
	args := os.Args[1:]
	for i, a := range args {
		if a == "--verbose" || a == "-v" {
			verbose = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	operation := args[0]

	switch operation {
	case "get", "store", "erase":
		runCredentialOp(operation, configPath, verbose)
	case "install":
		runInstall()
	case "list":
		fmt.Fprintln(os.Stderr, "list: not yet implemented (requires backend-specific enumeration)")
		os.Exit(1)
	case "version", "--version":
		fmt.Printf("git-credentials-org %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown operation: %s\n", operation)
		printUsage()
		os.Exit(1)
	}
}

func runCredentialOp(op, configPath string, verbose bool) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	h := handler.New(cfg, verbose)

	switch op {
	case "get":
		if err := h.Get(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "store":
		if err := h.Store(os.Stdin); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "erase":
		if err := h.Erase(os.Stdin); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runInstall() {
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error finding executable path: %v\n", err)
		os.Exit(1)
	}
	self, _ = filepath.EvalSymlinks(self)

	// Ensure config directory exists
	configDir := filepath.Dir(config.DefaultConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating config directory: %v\n", err)
		os.Exit(1)
	}

	// Write default config if it doesn't exist
	configPath := config.DefaultConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := `[defaults]
backend = "keychain"
log_level = "warn"

[hosts."gitlab.com"]
provider = "gitlab"

[hosts."github.com"]
provider = "github"
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing default config: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Created default config at %s\n", configPath)
	}

	// Configure git
	cmds := [][]string{
		{"git", "config", "--global", "credential.helper", self},
		{"git", "config", "--global", "credential.useHttpPath", "true"},
	}

	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error running %s: %v\n", strings.Join(c, " "), err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "Installed git-credentials-org as global credential helper.\n")
	fmt.Fprintf(os.Stderr, "  helper: %s\n", self)
	fmt.Fprintf(os.Stderr, "  config: %s\n", configPath)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `git-credentials-org %s

A Git credential helper that resolves credentials by organization namespace.

Usage:
  git-credentials-org <get|store|erase>   Git credential helper operations
  git-credentials-org install             Configure git to use this helper
  git-credentials-org version             Print version
  git-credentials-org help                Print this help

Flags:
  --verbose, -v    Enable debug logging (also: GIT_CREDENTIALS_ORG_DEBUG=1)

Environment:
  GIT_CREDENTIALS_ORG_CONFIG   Path to config file (default: ~/.config/git-credentials-org/config.toml)
  GIT_CREDENTIALS_ORG_DEBUG    Set to "1" to enable debug logging
`, version)
}
