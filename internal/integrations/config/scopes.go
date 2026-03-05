package config

import (
	"github.com/theopenlane/core/internal/integrations/types"
)

// MergeRequestedScopes merges the provider spec's default OAuth scopes with the caller-supplied
// scopes, deduplicating. Returns nil when the merged set is empty.
func MergeRequestedScopes(spec ProviderSpec, requested []string) []string {
	var base []string
	if spec.OAuth != nil {
		base = spec.OAuth.Scopes
	}

	merged := types.MergeScopes(base, requested...)
	if len(merged) == 0 {
		return nil
	}

	return merged
}
