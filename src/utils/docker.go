package utils

import (
	"fmt"
	"strings"

	"github.com/distribution/reference"
)

func ParseFullImageRefernce(fullRef string) (registry string, namespace string, repository string, tag string, err error) {
	namedRef, err := reference.ParseNamed(fullRef)
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid image reference format: %w", err)
	}

	registry = reference.Domain(namedRef)

	tag = "latest"
	if tagged, ok := namedRef.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	path := reference.Path(namedRef)
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		namespace = ""
		repository = parts[0]
	} else {
		namespace = parts[0]
		repository = strings.Join(parts[1:], "/")
	}

	return registry, namespace, repository, tag, nil
}
