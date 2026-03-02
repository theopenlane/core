package catalog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers/catalog"
)

// TestCatalogCompleteness asserts that every visible provider spec has a corresponding builder
// and that every builder has a corresponding provider spec. This catches catalog drift:
// adding a JSON spec without a builder, or adding a builder without a JSON spec.
func TestCatalogCompleteness(t *testing.T) {
	loader := config.NewFSLoader(config.ProvidersFS, "providers")
	specs, err := loader.Load()
	require.NoError(t, err, "loading provider specs")
	require.NotEmpty(t, specs, "expected at least one provider spec")

	builders := catalog.Builders()
	require.NotEmpty(t, builders, "expected at least one builder")

	builderTypes := make(map[types.ProviderType]struct{}, len(builders))
	for _, b := range builders {
		builderTypes[b.Type()] = struct{}{}
	}

	specTypes := make(map[types.ProviderType]struct{}, len(specs))
	for providerType, spec := range specs {
		// specs with visible=false are intentional config-only entries (e.g. awsauditmanager,
		// awssecurityhub) that share an underlying builder with a different provider type
		if spec.Visible != nil && !*spec.Visible {
			continue
		}
		specTypes[providerType] = struct{}{}
	}

	// every visible spec must have a builder
	for providerType := range specTypes {
		assert.Contains(t, builderTypes, providerType,
			"provider spec %q has no corresponding builder in catalog.Builders()", providerType)
	}

	// every builder must have a spec (prevents orphaned builders)
	for providerType := range builderTypes {
		assert.Contains(t, specTypes, providerType,
			"builder for provider %q has no corresponding JSON spec in config/providers/", providerType)
	}
}

// TestProviderTypeUniqueness asserts that all builders in the catalog return distinct type identifiers.
// Duplicate provider types cause silent overwrite in the registry.
func TestProviderTypeUniqueness(t *testing.T) {
	builders := catalog.Builders()
	require.NotEmpty(t, builders)

	seen := make(map[types.ProviderType]int, len(builders))
	for i, b := range builders {
		seen[b.Type()]++
		assert.Equal(t, 1, seen[b.Type()],
			"builder at index %d returns duplicate provider type %q", i, b.Type())
	}
}
