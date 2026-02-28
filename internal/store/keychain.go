package store

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const keychainServicePrefix = "git-credential-org"
const keychainAccount = "credentials"

type KeychainStore struct{}

func NewKeychainStore() *KeychainStore {
	return &KeychainStore{}
}

func (k *KeychainStore) Name() string {
	return "keychain"
}

// serviceName returns a per-namespace service name so each entry is
// visually distinct in Keychain Access, e.g. "git-credential-org:gitlab.com/org1".
func (k *KeychainStore) serviceName(namespace string) string {
	return keychainServicePrefix + ":" + namespace
}

func (k *KeychainStore) Get(namespace string) (*Credential, error) {
	data, err := keyring.Get(k.serviceName(namespace), keychainAccount)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("keychain get %q: %w", namespace, err)
	}

	var cred Credential
	if err := json.Unmarshal([]byte(data), &cred); err != nil {
		return nil, fmt.Errorf("keychain unmarshal %q: %w", namespace, err)
	}

	return &cred, nil
}

func (k *KeychainStore) Store(namespace string, cred *Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("keychain marshal: %w", err)
	}

	if err := keyring.Set(k.serviceName(namespace), keychainAccount, string(data)); err != nil {
		return fmt.Errorf("keychain store %q: %w", namespace, err)
	}

	return nil
}

func (k *KeychainStore) Erase(namespace string) error {
	if err := keyring.Delete(k.serviceName(namespace), keychainAccount); err != nil {
		if err == keyring.ErrNotFound {
			return nil
		}
		return fmt.Errorf("keychain erase %q: %w", namespace, err)
	}
	return nil
}
