package workflows

import "github.com/samber/lo"

// NormalizeStrings filters empty strings and returns unique values
func NormalizeStrings(input []string) []string {
	return lo.Uniq(lo.Filter(input, func(s string, _ int) bool { return s != "" }))
}
