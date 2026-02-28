# git-credentials-org

A Git credential helper that resolves credentials by **organization namespace**, solving the multi-org authentication problem on shared Git hosting platforms (GitLab, GitHub).

## The Problem

When working across multiple organizations on the same platform (e.g., `gitlab.com/org1` and `gitlab.com/org2`), standard credential helpers can't reliably distinguish which credentials to use. This tool derives a namespace from the URL path (`host/first-path-segment`) and stores credentials per namespace.

## Features

- **Namespace-based resolution**: `gitlab.com/org1` and `gitlab.com/org2` use separate credential sets automatically
- **Store-once-don't-ask**: Credentials are prompted once, stored, and reused. Erased only when git reports auth failure.
- **Pluggable backends**: macOS Keychain and 1Password (via `op` CLI)
- **Platform-aware**: Knows GitLab uses `oauth2` username with PATs, GitHub uses `x-access-token`, etc.
- **Zero-config for simple setups**: Works with sensible defaults out of the box

## Installation

### Homebrew (macOS)

```bash
brew install imcitius/tap/git-credentials-org
```

### Build from source

```bash
go install github.com/imcitius/git-credentials-org/cmd/git-credentials-org@latest
```

Or clone and build:

```bash
git clone https://github.com/imcitius/git-credentials-org.git
cd git-credentials-org
go build -o git-credentials-org ./cmd/git-credentials-org
sudo mv git-credentials-org /usr/local/bin/
```

### Quick setup

After installing, run:

```bash
git-credentials-org install
```

This will:
1. Create a default config at `~/.config/git-credentials-org/config.toml`
2. Set `credential.helper` and `credential.useHttpPath` in your global `.gitconfig`

### Manual setup

Add to your `~/.gitconfig` (the empty `helper =` line clears any previously configured helpers):

```ini
[credential]
    helper =
    helper = /usr/local/bin/git-credentials-org
    useHttpPath = true
```

## Configuration

Config file location: `~/.config/git-credentials-org/config.toml`

```toml
[defaults]
backend = "keychain"       # "keychain" or "onepassword"
log_level = "warn"

# Per-host settings
[hosts."gitlab.com"]
provider = "gitlab"        # Enables GitLab-specific behavior (oauth2 username)
# backend = "onepassword"  # Override backend for this host

[hosts."github.com"]
provider = "github"

# Backend-specific settings
[backends.onepassword]
vault = "Development"      # 1Password vault name
# account = "my.1password.com"
```

### Environment variables

| Variable | Description |
|---|---|
| `GIT_CREDENTIALS_ORG_CONFIG` | Custom config file path |
| `GIT_CREDENTIALS_ORG_DEBUG` | Set to `1` for debug logging |

## How It Works

1. Git calls `git-credentials-org get` with `protocol`, `host`, and `path` on stdin
2. The helper derives a namespace: `gitlab.com` + `org1/project/repo.git` → `gitlab.com/org1`
3. Looks up credentials in the configured backend for that namespace
4. If found, returns them. If not, prompts the user once and stores them.
5. On `store` (successful auth): updates the backend with working credentials
6. On `erase` (failed auth): removes credentials so the next `get` will prompt again

## Backends

### macOS Keychain (default)

Uses the macOS Keychain via the system security framework. Credentials are stored as JSON under the service name `git-credentials-org` with the namespace as the account key.

No additional setup required.

### 1Password

Requires the [1Password CLI](https://developer.1password.com/docs/cli/) (`op`) to be installed.

Credentials are stored as Login items in the specified vault with title pattern `git-credentials-org: <namespace>`.

Supports:
- Interactive use (biometric unlock via `op`)
- Service accounts (`OP_SERVICE_ACCOUNT_TOKEN` environment variable)

## Debugging

```bash
# Enable verbose logging
GIT_CREDENTIALS_ORG_DEBUG=1 git pull

# Or use the flag
git-credentials-org --verbose get <<EOF
protocol=https
host=gitlab.com
path=myorg/myproject.git

EOF
```

## License

MIT
