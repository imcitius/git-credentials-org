package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Credential struct {
	Protocol string
	Host     string
	Path     string
	Username string
	Password string
}

func Parse(r io.Reader) (*Credential, error) {
	cred := &Credential{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		switch key {
		case "protocol":
			cred.Protocol = value
		case "host":
			cred.Host = value
		case "path":
			cred.Path = value
		case "username":
			cred.Username = value
		case "password":
			cred.Password = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading credential input: %w", err)
	}

	return cred, nil
}

func Write(w io.Writer, cred *Credential) error {
	pairs := []struct{ key, val string }{
		{"protocol", cred.Protocol},
		{"host", cred.Host},
		{"path", cred.Path},
		{"username", cred.Username},
		{"password", cred.Password},
	}

	for _, p := range pairs {
		if p.val == "" {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s=%s\n", p.key, p.val); err != nil {
			return fmt.Errorf("writing credential output: %w", err)
		}
	}

	_, err := fmt.Fprintln(w)
	return err
}
