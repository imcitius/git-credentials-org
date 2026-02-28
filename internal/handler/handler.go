package handler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/imcitius/git-credential-org/internal/config"
	"github.com/imcitius/git-credential-org/internal/protocol"
	"github.com/imcitius/git-credential-org/internal/provider"
	"github.com/imcitius/git-credential-org/internal/resolver"
	"github.com/imcitius/git-credential-org/internal/store"
)

type Handler struct {
	cfg     *config.Config
	verbose bool
}

func New(cfg *config.Config, verbose bool) *Handler {
	return &Handler{cfg: cfg, verbose: verbose}
}

func (h *Handler) Get(r io.Reader, w io.Writer) error {
	cred, err := protocol.Parse(r)
	if err != nil {
		return err
	}

	namespace := resolver.Resolve(cred.Host, cred.Path)
	h.log("get: namespace=%s (host=%s, path=%s)", namespace, cred.Host, cred.Path)

	backend, err := h.storeForHost(cred.Host)
	if err != nil {
		return err
	}

	stored, err := backend.Get(namespace)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("backend %s get: %w", backend.Name(), err)
	}

	if stored != nil {
		h.log("get: found credentials in %s for %s", backend.Name(), namespace)
		return protocol.Write(w, &protocol.Credential{
			Protocol: cred.Protocol,
			Host:     cred.Host,
			Username: stored.Username,
			Password: stored.Password,
		})
	}

	// No stored credentials -- prompt the user but do NOT persist yet.
	// Git will call "store" after verifying auth succeeded, or "erase" on failure.
	h.log("get: no credentials found, prompting user (will persist on 'store' callback)")
	prov := h.providerForHost(cred.Host)
	newCred, err := h.promptForCredentials(prov, namespace)
	if err != nil {
		return fmt.Errorf("prompting for credentials: %w", err)
	}

	return protocol.Write(w, &protocol.Credential{
		Protocol: cred.Protocol,
		Host:     cred.Host,
		Username: newCred.Username,
		Password: newCred.Password,
	})
}

func (h *Handler) Store(r io.Reader) error {
	cred, err := protocol.Parse(r)
	if err != nil {
		return err
	}

	if cred.Username == "" || cred.Password == "" {
		return nil
	}

	namespace := resolver.Resolve(cred.Host, cred.Path)
	h.log("store: upsert for namespace=%s", namespace)

	backend, err := h.storeForHost(cred.Host)
	if err != nil {
		return err
	}

	return backend.Store(namespace, &store.Credential{
		Username: cred.Username,
		Password: cred.Password,
	})
}

func (h *Handler) Erase(r io.Reader) error {
	cred, err := protocol.Parse(r)
	if err != nil {
		return err
	}

	namespace := resolver.Resolve(cred.Host, cred.Path)
	h.log("erase: removing namespace=%s", namespace)

	backend, err := h.storeForHost(cred.Host)
	if err != nil {
		return err
	}

	return backend.Erase(namespace)
}

func (h *Handler) storeForHost(host string) (store.CredentialStore, error) {
	backendName := h.cfg.BackendForHost(host)
	return store.New(backendName, h.cfg)
}

func (h *Handler) providerForHost(host string) provider.Provider {
	configured := h.cfg.ProviderForHost(host)
	return provider.ForHost(host, configured)
}

func (h *Handler) promptForCredentials(prov provider.Provider, namespace string) (*store.Credential, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot open terminal for prompting: %w", err)
	}
	defer tty.Close()

	fd := int(tty.Fd())

	defaultUser := prov.DefaultUsername()
	if defaultUser != "" {
		fmt.Fprintf(tty, "%s", prov.TokenPrompt(namespace))
		token, err := term.ReadPassword(fd)
		fmt.Fprintln(tty) // newline after masked input
		if err != nil {
			return nil, fmt.Errorf("reading token: %w", err)
		}
		return &store.Credential{
			Username: defaultUser,
			Password: strings.TrimSpace(string(token)),
		}, nil
	}

	// Generic: prompt for username (visible) and password (masked)
	reader := bufio.NewReader(tty)
	fmt.Fprintf(tty, "%s", prov.TokenPrompt(namespace))
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(tty, "Password: ")
	password, err := term.ReadPassword(fd)
	fmt.Fprintln(tty)
	if err != nil {
		return nil, fmt.Errorf("reading password: %w", err)
	}

	return &store.Credential{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(string(password)),
	}, nil
}

func (h *Handler) log(format string, args ...any) {
	if h.verbose {
		fmt.Fprintf(os.Stderr, "[git-credential-org] "+format+"\n", args...)
	}
}
