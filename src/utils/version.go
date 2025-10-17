package utils

import "fmt"

func ParseSemanticVersion(version string) (major, minor, patch int, err error) {
	_, err = fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	return major, minor, patch, err
}
