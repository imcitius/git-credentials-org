package store

import (
	"errors"
	"fmt"

	"github.com/imcitius/git-credentials-org/internal/config"
)

var ErrNotFound = errors.New("credentials not found")

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CredentialStore interface {
	Get(namespace string) (*Credential, error)
	Store(namespace string, cred *Credential) error
	Erase(namespace string) error
	Name() string
}

func New(backendName string, cfg *config.Config) (CredentialStore, error) {
	switch backendName {
	case "keychain":
		return NewKeychainStore(), nil
	case "onepassword", "1password":
		bc := cfg.Backends[backendName]
		vault := bc.Vault
		if vault == "" {
			vault = "Private"
		}
		return NewOnePasswordStore(vault, bc.Account), nil
	default:
		return nil, fmt.Errorf("unknown backend: %s", backendName)
	}
}
