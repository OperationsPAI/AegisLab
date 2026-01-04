package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

// HelmKeyPart represents a parsed key part that may contain an array index
type HelmKeyPart struct {
	Key     string
	IsArray bool
	Index   int
}

var envVarRegex = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// helmKeyRegex validates Helm value keys with support for:
// - Simple keys: "key"
// - Nested keys: "key.subkey"
// - Array indices: "key[0]", "key[123]"
// - Complex paths: "accounting.initContainers[0].image"
var helmKeyRegex = regexp.MustCompile(`^[a-zA-Z_-][a-zA-Z0-9_-]*(\[\d+\])?(\.[a-zA-Z_-][a-zA-Z0-9_-]*(\[\d+\])?)*$`)

// arrayIndexRegex matches array indices like [0], [123]
var arrayIndexRegex = regexp.MustCompile(`^(.+?)\[(\d+)\]$`)

// ConvertSimpleTypeToString converts simple types (string, int, float64, bool) to their string representation
func ConvertSimpleTypeToString(a any) (string, error) {
	switch v := a.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported type %T for conversion to string", a)
	}
}

// ConvertStringToSimpleType converts a string to a simple type (string, int, float64, bool)
func ConvertStringToSimpleType(s string) (any, error) {
	if s == "" {
		return s, nil
	}

	var value any

	// Check for leading zeros - if present, keep as string to preserve format
	// e.g., "023" should remain "023", not be converted to 23
	// Also handle negative numbers with leading zeros after the minus sign, e.g., "-023"
	if len(s) > 1 && s[0] == '0' && s[1] >= '0' && s[1] <= '9' {
		// Has leading zero (not "0" alone, not "0.xxx"), keep as string
		return s, nil
	}
	if len(s) > 2 && s[0] == '-' && s[1] == '0' && s[2] >= '0' && s[2] <= '9' {
		// Negative number with leading zero, e.g., "-023", keep as string
		return s, nil
	}

	if convertedValueI, err := strconv.Atoi(s); err == nil {
		value = convertedValueI
		return value, nil
	}

	if convertedValueF, err := strconv.ParseFloat(s, 64); err == nil {
		value = convertedValueF
		return value, nil
	}

	if convertedValueB, err := strconv.ParseBool(s); err == nil {
		value = convertedValueB
		return value, nil
	}

	value = s
	return value, nil
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

// GenerateULID generates a ULID string based on the provided time.
func GenerateULID(t *time.Time) string {
	if t == nil {
		now := time.Now()
		t = &now
	}

	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(*t), entropy)
	return id.String()
}

// IsValidEnvVar checks if the provided string is a valid environment variable name
func IsValidEnvVar(envVar string) error {
	if envVar == "" {
		return fmt.Errorf("environment variable cannot be empty")
	}
	if len(envVar) > 128 {
		return fmt.Errorf("environment variable name too long (max 128 characters)")
	}
	if ok := envVarRegex.MatchString(envVar); !ok {
		return fmt.Errorf("environment variable contains invalid characters")
	}
	return nil
}

// IsValidHelmValueKey checks if the provided string is a valid Helm Value Path key
func IsValidHelmValueKey(key string) error {
	if key == "" {
		return fmt.Errorf("Helm value key cannot be empty")
	}
	if ok := helmKeyRegex.MatchString(key); !ok {
		return fmt.Errorf("Helm value key contains invalid characters")
	}
	return nil
}

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// ParseHelmKey parses a Helm key like "accounting.initContainers[0].image"
// into structured parts that include array indices
func ParseHelmKey(key string) []HelmKeyPart {
	parts := strings.Split(key, ".")
	result := make([]HelmKeyPart, 0, len(parts))

	for _, part := range parts {
		if matches := arrayIndexRegex.FindStringSubmatch(part); matches != nil {
			// This part has an array index
			keyName := matches[1]
			index, _ := strconv.Atoi(matches[2])
			result = append(result, HelmKeyPart{
				Key:     keyName,
				IsArray: true,
				Index:   index,
			})
		} else {
			// Regular key without array index
			result = append(result, HelmKeyPart{
				Key:     part,
				IsArray: false,
				Index:   0,
			})
		}
	}

	return result
}

func ToSnakeCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
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
