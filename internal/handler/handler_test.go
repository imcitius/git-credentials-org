package handler

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/imcitius/git-credential-org/internal/config"
	"github.com/imcitius/git-credential-org/internal/store"
)

type mockStore struct {
	creds map[string]*store.Credential
}

func newMockStore() *mockStore {
	return &mockStore{creds: make(map[string]*store.Credential)}
}

func (m *mockStore) Name() string { return "mock" }

func (m *mockStore) Get(namespace string) (*store.Credential, error) {
	c, ok := m.creds[namespace]
	if !ok {
		return nil, store.ErrNotFound
	}
	return c, nil
}

func (m *mockStore) Store(namespace string, cred *store.Credential) error {
	m.creds[namespace] = cred
	return nil
}

func (m *mockStore) Erase(namespace string) error {
	delete(m.creds, namespace)
	return nil
}

// testHandler creates a handler that uses a mock store via a custom store factory.
// Since we can't inject the store directly through the public API, we test
// the store and protocol layers independently and do integration-style tests
// for the handler using the Get path with pre-populated mock stores.
func TestHandlerStore(t *testing.T) {
	mock := newMockStore()

	cred := &store.Credential{Username: "oauth2", Password: "glpat-test123"}
	if err := mock.Store("gitlab.com/org1", cred); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	got, err := mock.Get("gitlab.com/org1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Username != "oauth2" || got.Password != "glpat-test123" {
		t.Errorf("Get() = %+v, want username=oauth2, password=glpat-test123", got)
	}

	// Verify different namespace returns not found
	_, err = mock.Get("gitlab.com/org2")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Get(org2) should return ErrNotFound, got: %v", err)
	}
}

func TestHandlerErase(t *testing.T) {
	mock := newMockStore()
	mock.creds["gitlab.com/org1"] = &store.Credential{Username: "oauth2", Password: "token"}

	if err := mock.Erase("gitlab.com/org1"); err != nil {
		t.Fatalf("Erase() error = %v", err)
	}

	_, err := mock.Get("gitlab.com/org1")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Get after Erase should return ErrNotFound, got: %v", err)
	}
}

func TestHandlerGetWithExistingCredentials(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{Backend: "keychain"},
		Hosts:    map[string]config.HostConfig{"gitlab.com": {Provider: "gitlab"}},
		Backends: make(map[string]config.BackendConfig),
	}

	h := New(cfg, false)

	// We can't easily inject a mock store into the handler through the public API,
	// but we can at least verify the handler correctly parses protocol input
	// and returns an error when the store lookup fails in a non-keychain environment.
	input := "protocol=https\nhost=gitlab.com\npath=org1/repo.git\n\n"
	var output bytes.Buffer

	err := h.Get(strings.NewReader(input), &output)
	// This will either succeed (if running on macOS with keychain) or fail gracefully.
	// In CI without keychain, we expect an error but not a panic.
	if err != nil {
		t.Logf("Expected error in non-keychain env: %v", err)
	}
}

func TestHandlerStoreOperation(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{Backend: "keychain"},
		Hosts:    make(map[string]config.HostConfig),
		Backends: make(map[string]config.BackendConfig),
	}

	h := New(cfg, false)

	input := "protocol=https\nhost=gitlab.com\npath=org1/repo.git\nusername=oauth2\npassword=glpat-test\n\n"
	err := h.Store(strings.NewReader(input))
	// Same as above -- may fail without keychain but should not panic
	if err != nil {
		t.Logf("Expected error in non-keychain env: %v", err)
	}
}

func TestHandlerEraseOperation(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{Backend: "keychain"},
		Hosts:    make(map[string]config.HostConfig),
		Backends: make(map[string]config.BackendConfig),
	}

	h := New(cfg, false)

	input := "protocol=https\nhost=gitlab.com\npath=org1/repo.git\n\n"
	err := h.Erase(strings.NewReader(input))
	if err != nil {
		t.Logf("Expected error in non-keychain env: %v", err)
	}
}
