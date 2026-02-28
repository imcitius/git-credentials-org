package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Credential
	}{
		{
			name:  "full credential",
			input: "protocol=https\nhost=gitlab.com\npath=org1/project/repo.git\nusername=oauth2\npassword=glpat-abc123\n\n",
			want: Credential{
				Protocol: "https",
				Host:     "gitlab.com",
				Path:     "org1/project/repo.git",
				Username: "oauth2",
				Password: "glpat-abc123",
			},
		},
		{
			name:  "get request (no username/password)",
			input: "protocol=https\nhost=gitlab.com\npath=org1/project/repo.git\n\n",
			want: Credential{
				Protocol: "https",
				Host:     "gitlab.com",
				Path:     "org1/project/repo.git",
			},
		},
		{
			name:  "host with port",
			input: "protocol=https\nhost=gitlab.example.com:8443\npath=org1/repo.git\n\n",
			want: Credential{
				Protocol: "https",
				Host:     "gitlab.example.com:8443",
				Path:     "org1/repo.git",
			},
		},
		{
			name:  "empty input",
			input: "\n",
			want:  Credential{},
		},
		{
			name:  "unknown keys ignored",
			input: "protocol=https\nhost=gitlab.com\nunknown=value\n\n",
			want: Credential{
				Protocol: "https",
				Host:     "gitlab.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if *got != tt.want {
				t.Errorf("Parse() = %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name string
		cred Credential
		want []string
	}{
		{
			name: "full credential",
			cred: Credential{
				Protocol: "https",
				Host:     "gitlab.com",
				Username: "oauth2",
				Password: "glpat-abc",
			},
			want: []string{"protocol=https", "host=gitlab.com", "username=oauth2", "password=glpat-abc"},
		},
		{
			name: "skips empty fields",
			cred: Credential{
				Protocol: "https",
				Host:     "gitlab.com",
				Username: "oauth2",
				Password: "token",
			},
			want: []string{"protocol=https", "host=gitlab.com", "username=oauth2", "password=token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Write(&buf, &tt.cred); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			output := buf.String()
			for _, expected := range tt.want {
				if !strings.Contains(output, expected) {
					t.Errorf("Write() output missing %q, got:\n%s", expected, output)
				}
			}

			if !strings.HasSuffix(output, "\n\n") {
				t.Errorf("Write() output should end with blank line, got:\n%q", output)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	original := &Credential{
		Protocol: "https",
		Host:     "gitlab.com",
		Path:     "org1/repo.git",
		Username: "oauth2",
		Password: "glpat-xyz789",
	}

	var buf bytes.Buffer
	if err := Write(&buf, original); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	parsed, err := Parse(&buf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if *parsed != *original {
		t.Errorf("round-trip failed: got %+v, want %+v", *parsed, *original)
	}
}
