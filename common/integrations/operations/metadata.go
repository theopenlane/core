package operations

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
)

// sanitizeDescriptors is a generic helper that filters and assigns provider type to descriptors
func sanitizeDescriptors[T any](provider types.ProviderType, descriptors []T, isValid func(T) bool, getProvider func(T) types.ProviderType, setProvider func(*T, types.ProviderType)) []T {
	if len(descriptors) == 0 {
		return nil
	}

	out := lo.FilterMap(descriptors, func(descriptor T, _ int) (T, bool) {
		if !isValid(descriptor) {
			var zero T
			return zero, false
		}
		if getProvider(descriptor) == types.ProviderUnknown {
			setProvider(&descriptor, provider)
		}

		return descriptor, true
	})
	if len(out) == 0 {
		return nil
	}

	return out
}

// SanitizeOperationDescriptors filters and cleans a slice of OperationDescriptor
func SanitizeOperationDescriptors(provider types.ProviderType, descriptors []types.OperationDescriptor) []types.OperationDescriptor {
	return sanitizeDescriptors(
		provider,
		descriptors,
		func(d types.OperationDescriptor) bool { return d.Run != nil && d.Name != "" },
		func(d types.OperationDescriptor) types.ProviderType { return d.Provider },
		func(d *types.OperationDescriptor, p types.ProviderType) { d.Provider = p },
	)
}

// SanitizeClientDescriptors filters out invalid client descriptors and assigns provider type
func SanitizeClientDescriptors(provider types.ProviderType, descriptors []types.ClientDescriptor) []types.ClientDescriptor {
	return sanitizeDescriptors(
		provider,
		descriptors,
		func(d types.ClientDescriptor) bool { return d.Build != nil },
		func(d types.ClientDescriptor) types.ProviderType { return d.Provider },
		func(d *types.ClientDescriptor, p types.ProviderType) { d.Provider = p },
	)
}
