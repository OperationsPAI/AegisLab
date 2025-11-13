package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func IsValidEnvVar(envVar string) error {
	if envVar == "" {
		return fmt.Errorf("environment variable cannot be empty")
	}

	if len(envVar) > 128 {
		return fmt.Errorf("environment variable name too long (max 128 characters)")
	}

	envVarPattern := `^[A-Z_][A-Z0-9_]*$`
	matched, err := regexp.MatchString(envVarPattern, envVar)
	if err != nil {
		return fmt.Errorf("pattern match error: %v", err)
	}
	if !matched {
		return fmt.Errorf("environment variable must contain only uppercase letters, numbers, and underscores, and start with a letter or underscore")
	}

	return nil
}

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

// GenerateColorFromKey generates a consistent color based on a key string
func GenerateColorFromKey(key string) string {
	// Predefined color palette with good visibility and contrast
	colors := []string{
		"#f44336", // Red
		"#e91e63", // Pink
		"#9c27b0", // Purple
		"#673ab7", // Deep Purple
		"#3f51b5", // Indigo
		"#2196f3", // Blue
		"#03a9f4", // Light Blue
		"#00bcd4", // Cyan
		"#009688", // Teal
		"#4caf50", // Green
		"#8bc34a", // Light Green
		"#cddc39", // Lime
		"#ffeb3b", // Yellow
		"#ffc107", // Amber
		"#ff9800", // Orange
		"#ff5722", // Deep Orange
		"#795548", // Brown
		"#607d8b", // Blue Grey
	}

	// Simple hash function to get consistent color for same key
	hash := 0
	for _, char := range key {
		hash = (hash*31 + int(char)) % len(colors)
	}

	return colors[hash]
}

func ToSingular(plural string) string {
	if len(plural) < 1 {
		return plural
	}

	irregular := map[string]string{
		"people": "person",
		"men":    "man",
		"women":  "woman",
		"data":   "datum",
		"feet":   "foot",
	}
	if s, ok := irregular[plural]; ok {
		return s
	}

	if strings.HasSuffix(plural, "s") && len(plural) > 1 {
		if strings.HasSuffix(plural, "ss") {
			return plural
		}

		if strings.HasSuffix(plural, "ies") && len(plural) > 3 {
			return plural[:len(plural)-3] + "y"
		}

		if !strings.HasSuffix(plural, "es") {
			return plural[:len(plural)-1] // 移除末尾的 's'
		}
	}

	if strings.HasSuffix(plural, "es") && len(plural) > 2 {
		return plural[:len(plural)-2]
	}

	return plural
}
