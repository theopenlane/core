package keystore

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
)

// groupDescriptors groups a keyed descriptor map by provider using a key extraction function.
func groupDescriptors[K comparable, T any](data map[K]T, providerOf func(K) types.ProviderType) map[types.ProviderType][]T {
	grouped := map[types.ProviderType][]T{}
	for key, descriptor := range data {
		provider := providerOf(key)
		grouped[provider] = append(grouped[provider], descriptor)
	}

	for provider, descriptors := range grouped {
		copied := make([]T, len(descriptors))
		copy(copied, descriptors)
		grouped[provider] = copied
	}

	return grouped
}

// flattenDescriptors converts a grouped provider descriptor map into a flat slice.
func flattenDescriptors[T any](grouped map[types.ProviderType][]T) []T {
	if len(grouped) == 0 {
		return nil
	}

	return lo.Flatten(lo.Values(grouped))
}
