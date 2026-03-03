package catalog_test

import (
	"context"
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

	allSpecTypes := make(map[types.ProviderType]struct{}, len(specs))
	visibleSpecTypes := make(map[types.ProviderType]struct{}, len(specs))
	for providerType, spec := range specs {
		allSpecTypes[providerType] = struct{}{}
		if spec.Visible == nil || *spec.Visible {
			visibleSpecTypes[providerType] = struct{}{}
		}
	}

	// every visible spec must have a builder
	for providerType := range visibleSpecTypes {
		assert.Contains(t, builderTypes, providerType,
			"provider spec %q has no corresponding builder in catalog.Builders()", providerType)
	}

	// every builder must have a spec (prevents orphaned builders), including hidden specs
	for providerType := range builderTypes {
		assert.Contains(t, allSpecTypes, providerType,
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

// TestCatalogProviderConformance asserts each catalog builder can build from its spec and
// that published client/operation descriptors are structurally valid for registry ingestion.
func TestCatalogProviderConformance(t *testing.T) {
	loader := config.NewFSLoader(config.ProvidersFS, "providers")
	specs, err := loader.Load()
	require.NoError(t, err, "loading provider specs")

	builders := catalog.Builders()
	require.NotEmpty(t, builders, "expected at least one builder")

	for _, builder := range builders {
		providerType := builder.Type()

		spec, ok := specs[providerType]
		require.Truef(t, ok, "missing spec for provider %q", providerType)

		provider, err := builder.Build(context.Background(), spec)
		require.NoErrorf(t, err, "builder failed for provider %q", providerType)
		require.NotNilf(t, provider, "builder returned nil provider for %q", providerType)
		require.Equalf(t, providerType, provider.Type(), "provider type mismatch for %q", providerType)

		if operationProvider, ok := provider.(types.OperationProvider); ok {
			seen := map[types.OperationName]struct{}{}
			for i, descriptor := range operationProvider.Operations() {
				assert.NotEmptyf(t, descriptor.Name, "provider %q operation %d missing name", providerType, i)
				assert.NotNilf(t, descriptor.Run, "provider %q operation %q missing run function", providerType, descriptor.Name)
				if descriptor.Provider != types.ProviderUnknown {
					assert.Equalf(t, providerType, descriptor.Provider, "provider %q operation %q has mismatched provider", providerType, descriptor.Name)
				}
				if _, dup := seen[descriptor.Name]; dup {
					assert.Failf(t, "duplicate operation name", "provider %q has duplicate operation name %q", providerType, descriptor.Name)
				}
				seen[descriptor.Name] = struct{}{}
			}
		}

		if clientProvider, ok := provider.(types.ClientProvider); ok {
			seen := map[types.ClientName]struct{}{}
			for i, descriptor := range clientProvider.ClientDescriptors() {
				assert.NotEmptyf(t, descriptor.Name, "provider %q client %d missing name", providerType, i)
				assert.NotNilf(t, descriptor.Build, "provider %q client %q missing build function", providerType, descriptor.Name)
				if descriptor.Provider != types.ProviderUnknown {
					assert.Equalf(t, providerType, descriptor.Provider, "provider %q client %q has mismatched provider", providerType, descriptor.Name)
				}
				if _, dup := seen[descriptor.Name]; dup {
					assert.Failf(t, "duplicate client name", "provider %q has duplicate client name %q", providerType, descriptor.Name)
				}
				seen[descriptor.Name] = struct{}{}
			}
		}
	}
}
