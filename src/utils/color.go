package utils

import "regexp"

var hexColorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func IsValidHexColor(hex string) bool {
	if hex == "" {
		return false
	}
	return hexColorRegex.MatchString(hex)
}
