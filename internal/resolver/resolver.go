package resolver

import (
	"strings"
)

// Resolve derives a namespace from a host and path.
// For "gitlab.com" + "org1/project/repo.git" it returns "gitlab.com/org1".
// For self-hosted instances without a path prefix, it returns just the host.
func Resolve(host, path string) string {
	if path == "" {
		return host
	}

	path = strings.TrimPrefix(path, "/")
	firstSegment, _, _ := strings.Cut(path, "/")
	firstSegment = strings.TrimSuffix(firstSegment, ".git")

	if firstSegment == "" {
		return host
	}

	return host + "/" + firstSegment
}
