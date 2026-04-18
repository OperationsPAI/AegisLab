package cmd

import (
	"fmt"
	"strings"
)

// imageRefParts represents a parsed container image reference split into its
// four database columns.
type imageRefParts struct {
	Registry   string `json:"registry"`
	Namespace  string `json:"namespace"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

// String renders the parts back to a canonical "<registry>/<namespace>/<repository>:<tag>"
// form, omitting the namespace segment when empty.
func (p imageRefParts) String() string {
	if p.Namespace == "" {
		return fmt.Sprintf("%s/%s:%s", p.Registry, p.Repository, p.Tag)
	}
	return fmt.Sprintf("%s/%s/%s:%s", p.Registry, p.Namespace, p.Repository, p.Tag)
}

// parseImageRef splits a container image reference into (registry, namespace,
// repository, tag).
//
// Rules:
//   - A tag is REQUIRED. References without ":<tag>" are rejected.
//   - A registry is OPTIONAL. If the first path segment does not look like a
//     host (no dot, no colon, and not "localhost"), the registry defaults to
//     "docker.io".
//   - Nested namespaces are preserved: in "docker.io/foo/bar/baz:tag" the
//     registry is "docker.io", namespace "foo/bar", repository "baz".
//   - Digest refs ("@sha256:...") are rejected. The backend row stores a tag,
//     not a digest; accepting digests would silently drop the digest.
func parseImageRef(ref string) (imageRefParts, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return imageRefParts{}, fmt.Errorf("image reference is empty")
	}
	if strings.Contains(ref, "@") {
		return imageRefParts{}, fmt.Errorf("digest references are not supported; provide a ':<tag>' reference instead")
	}

	lastSlash := strings.LastIndex(ref, "/")
	lastColon := strings.LastIndex(ref, ":")
	if lastColon <= lastSlash || lastColon == -1 {
		return imageRefParts{}, fmt.Errorf("image reference %q is missing ':<tag>'", ref)
	}
	nameAndRepo := ref[:lastColon]
	tag := ref[lastColon+1:]
	if tag == "" {
		return imageRefParts{}, fmt.Errorf("image reference %q has an empty tag", ref)
	}

	parts := strings.Split(nameAndRepo, "/")
	if len(parts) == 0 || parts[0] == "" {
		return imageRefParts{}, fmt.Errorf("image reference %q is malformed", ref)
	}

	var registry string
	var pathParts []string
	if looksLikeRegistry(parts[0]) && len(parts) > 1 {
		registry = parts[0]
		pathParts = parts[1:]
	} else {
		registry = "docker.io"
		pathParts = parts
	}

	if len(pathParts) == 0 || pathParts[len(pathParts)-1] == "" {
		return imageRefParts{}, fmt.Errorf("image reference %q is missing a repository", ref)
	}

	repository := pathParts[len(pathParts)-1]
	namespace := ""
	if len(pathParts) > 1 {
		namespace = strings.Join(pathParts[:len(pathParts)-1], "/")
	}

	return imageRefParts{
		Registry:   registry,
		Namespace:  namespace,
		Repository: repository,
		Tag:        tag,
	}, nil
}

func looksLikeRegistry(segment string) bool {
	if segment == "localhost" {
		return true
	}
	return strings.ContainsAny(segment, ".:")
}
