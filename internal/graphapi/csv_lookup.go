package graphapi

import (
	"strings"

	"github.com/samber/lo"
)

// normalizeCSVReferenceKey normalizes input values for lookup comparisons.
func normalizeCSVReferenceKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// normalizeUniqueValues trims, normalizes, and de-duplicates values by normalized key.
func normalizeUniqueValues(values []string, normalize func(string) string) map[string]string {
	if normalize == nil {
		normalize = normalizeCSVReferenceKey
	}

	trimmed := lo.FilterMap(values, func(value string, _ int) (string, bool) {
		value = strings.TrimSpace(value)
		return value, value != ""
	})

	if len(trimmed) == 0 {
		return nil
	}

	unique := lo.UniqBy(trimmed, func(value string) string {
		return normalize(value)
	})

	normalized := lo.SliceToMap(unique, func(value string) (string, string) {
		return normalize(value), value
	})

	delete(normalized, "")

	return normalized
}
