package utils

import (
	"aegis/config"
	"fmt"
	"regexp"
)

const (
	MetaPattern            = `^\^[a-zA-Z][a-zA-Z0-9_-]*\\d\+\$$`
	NsPrefixAndNumberRegex = `^([a-zA-Z][a-zA-Z0-9_-]*)(\d+)$`
)

func GetNsConfigMap() (map[string]map[string]any, error) {
	m := config.GetMap("injection.namespace_config")
	nsConfigMap := make(map[string]map[string]any, len(m))
	for ns, c := range m {
		config, ok := c.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid namespace config for %s", ns)
		}

		nsConfigMap[ns] = config
	}

	return nsConfigMap, nil
}

func GetNsCountMap() (map[string]int, error) {
	nsConfigMap, err := GetNsConfigMap()
	if err != nil {
		return nil, err
	}

	nsCountMap := make(map[string]int, len(nsConfigMap))
	for ns, config := range nsConfigMap {
		value, exists := config["count"]
		if !exists {
			return nil, fmt.Errorf("namespace %s does not have a count field", ns)
		}

		vInt, ok := value.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid namespace value for %s", ns)
		}

		nsCountMap[ns] = int(vInt)
	}

	return nsCountMap, nil
}

func GetAllNamespaces() ([]string, error) {
	nsCountMap, err := GetNsCountMap()
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(nsCountMap))
	for ns, count := range nsCountMap {
		for idx := range count {
			namespaces = append(namespaces, fmt.Sprintf("%s%d", ns, idx))
		}
	}

	return namespaces, nil
}

// ValidateNsPattern validates that a namespace pattern is a valid regex pattern for matching namespaces
// It checks that the pattern itself is a valid regex and follows the expected format for namespace patterns
// Valid patterns should match the format: ^<prefix>\d+$ where prefix can contain letters, hyphens, underscores
func ValidateNsPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("namespace pattern cannot be empty")
	}

	// First, validate that it's a valid regex
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("namespace pattern '%s' is not a valid regex: %w", pattern, err)
	}

	metaRe := regexp.MustCompile(MetaPattern)

	if !metaRe.MatchString(pattern) {
		return fmt.Errorf("namespace pattern '%s' must follow format ^<prefix>\\d+$ (e.g., '^ts\\d+$', '^exp-dev\\d+$')", pattern)
	}

	return nil
}

// ExtractNsPrefixAndNumber extracts the namespace prefix and number from a namespace string
func ExtractNsPrefixAndNumber(namespace string) (prefix string, number int, err error) {
	// Match pattern: prefix followed by one or more digits at the end
	re := regexp.MustCompile(NsPrefixAndNumberRegex)
	matches := re.FindStringSubmatch(namespace)

	if len(matches) != 3 {
		return "", 0, fmt.Errorf("namespace '%s' does not match pattern <prefix><number>", namespace)
	}

	prefix = matches[1]
	_, err = fmt.Sscanf(matches[2], "%d", &number)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse number from namespace '%s': %w", namespace, err)
	}

	return prefix, number, nil
}
