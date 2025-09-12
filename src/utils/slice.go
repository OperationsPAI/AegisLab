package utils

import "strings"

func FilterEmptyStrings(strs []string) []string {
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func Union(a, b []string) []string {
	set := make(map[string]struct{})
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		set[s] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}
