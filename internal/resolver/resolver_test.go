package resolver

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		name string
		host string
		path string
		want string
	}{
		{
			name: "gitlab with org and project",
			host: "gitlab.com",
			path: "org1/project/repo.git",
			want: "gitlab.com/org1",
		},
		{
			name: "gitlab with org only",
			host: "gitlab.com",
			path: "org1/repo.git",
			want: "gitlab.com/org1",
		},
		{
			name: "github with org and repo",
			host: "github.com",
			path: "mycompany/backend.git",
			want: "github.com/mycompany",
		},
		{
			name: "path with leading slash",
			host: "gitlab.com",
			path: "/org1/project/repo.git",
			want: "gitlab.com/org1",
		},
		{
			name: "empty path returns host only",
			host: "gitlab.com",
			path: "",
			want: "gitlab.com",
		},
		{
			name: "path is just .git",
			host: "gitlab.com",
			path: "repo.git",
			want: "gitlab.com/repo",
		},
		{
			name: "self-hosted with port",
			host: "git.example.com:8443",
			path: "team/project.git",
			want: "git.example.com:8443/team",
		},
		{
			name: "deeply nested path",
			host: "gitlab.com",
			path: "org1/group/subgroup/repo.git",
			want: "gitlab.com/org1",
		},
		{
			name: "path without .git suffix",
			host: "github.com",
			path: "org2/service",
			want: "github.com/org2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.host, tt.path)
			if got != tt.want {
				t.Errorf("Resolve(%q, %q) = %q, want %q", tt.host, tt.path, got, tt.want)
			}
		})
	}
}
