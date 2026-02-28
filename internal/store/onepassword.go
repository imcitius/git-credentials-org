package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type OnePasswordStore struct {
	vault   string
	account string
}

func NewOnePasswordStore(vault, account string) *OnePasswordStore {
	return &OnePasswordStore{vault: vault, account: account}
}

func (o *OnePasswordStore) Name() string {
	return "onepassword"
}

func (o *OnePasswordStore) itemTitle(namespace string) string {
	return "git-credential-org: " + namespace
}

func (o *OnePasswordStore) Get(namespace string) (*Credential, error) {
	title := o.itemTitle(namespace)

	args := []string{"item", "get", title, "--vault", o.vault, "--format", "json"}
	if o.account != "" {
		args = append(args, "--account", o.account)
	}

	out, err := o.run(args...)
	if err != nil {
		if strings.Contains(err.Error(), "isn't an item") ||
			strings.Contains(err.Error(), "not found") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("1password get %q: %w", namespace, err)
	}

	return o.parseItemJSON(out)
}

func (o *OnePasswordStore) Store(namespace string, cred *Credential) error {
	title := o.itemTitle(namespace)

	existing, err := o.Get(namespace)
	if err != nil && err != ErrNotFound {
		return err
	}

	if existing != nil {
		return o.editItem(title, cred)
	}
	return o.createItem(title, cred)
}

func (o *OnePasswordStore) Erase(namespace string) error {
	title := o.itemTitle(namespace)

	args := []string{"item", "delete", title, "--vault", o.vault}
	if o.account != "" {
		args = append(args, "--account", o.account)
	}

	_, err := o.run(args...)
	if err != nil {
		if strings.Contains(err.Error(), "isn't an item") ||
			strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("1password erase %q: %w", namespace, err)
	}
	return nil
}

func (o *OnePasswordStore) createItem(title string, cred *Credential) error {
	args := []string{
		"item", "create",
		"--category", "login",
		"--title", title,
		"--vault", o.vault,
		"--", // separator for field assignments
		fmt.Sprintf("username=%s", cred.Username),
		fmt.Sprintf("password=%s", cred.Password),
	}
	if o.account != "" {
		// insert before the -- separator
		args = append(args[:7], append([]string{"--account", o.account}, args[7:]...)...)
	}

	_, err := o.run(args...)
	if err != nil {
		return fmt.Errorf("1password create item: %w", err)
	}
	return nil
}

func (o *OnePasswordStore) editItem(title string, cred *Credential) error {
	args := []string{
		"item", "edit", title,
		"--vault", o.vault,
		"--", // separator for field assignments
		fmt.Sprintf("username=%s", cred.Username),
		fmt.Sprintf("password=%s", cred.Password),
	}
	if o.account != "" {
		args = append(args[:5], append([]string{"--account", o.account}, args[5:]...)...)
	}

	_, err := o.run(args...)
	if err != nil {
		return fmt.Errorf("1password edit item: %w", err)
	}
	return nil
}

func (o *OnePasswordStore) run(args ...string) ([]byte, error) {
	cmd := exec.Command("op", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

type opItem struct {
	Fields []opField `json:"fields"`
}

type opField struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

func (o *OnePasswordStore) parseItemJSON(data []byte) (*Credential, error) {
	var item opItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("parsing 1password item: %w", err)
	}

	cred := &Credential{}
	for _, f := range item.Fields {
		switch f.ID {
		case "username":
			cred.Username = f.Value
		case "password":
			cred.Password = f.Value
		}
	}

	if cred.Username == "" && cred.Password == "" {
		return nil, ErrNotFound
	}

	return cred, nil
}
