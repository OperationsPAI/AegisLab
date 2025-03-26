package utils

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
