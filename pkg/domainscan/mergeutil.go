package domainscan

// mergeStrings unions two string slices, dropping empty values and duplicates
// while preserving first-seen order
func mergeStrings(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	merged := make([]string, 0, len(a)+len(b))

	for _, s := range append(append([]string{}, a...), b...) {
		if s == "" {
			continue
		}

		if _, ok := seen[s]; ok {
			continue
		}

		seen[s] = struct{}{}
		merged = append(merged, s)
	}

	return merged
}
