package utils

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func ToSnakeCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
