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

// SubtractArrays returns a new slice containing elements in 'a' that are not in 'b'.
func SubtractArrays[T comparable](a, b []T) []T {
	set := make(map[T]struct{})
	for _, v := range b {
		set[v] = struct{}{}
	}

	result := make([]T, 0, len(a))
	for _, v := range a {
		if _, exists := set[v]; !exists {
			result = append(result, v)
		}
	}
	return result
}

// Union returns the union of two slices, removing duplicates.
func Union[T comparable](a, b []T) []T {
	set := make(map[T]struct{})
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		set[s] = struct{}{}
	}

	result := make([]T, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}
