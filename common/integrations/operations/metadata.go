package operations

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
)

// SanitizeOperationDescriptors filters and cleans a slice of OperationDescriptor
func SanitizeOperationDescriptors(provider types.ProviderType, descriptors []types.OperationDescriptor) []types.OperationDescriptor {
	if len(descriptors) == 0 {
		return nil
	}

	out := lo.FilterMap(descriptors, func(descriptor types.OperationDescriptor, _ int) (types.OperationDescriptor, bool) {
		if descriptor.Run == nil || descriptor.Name == "" {
			return types.OperationDescriptor{}, false
		}
		if descriptor.Provider == types.ProviderUnknown {
			descriptor.Provider = provider
		}
		return descriptor, true
	})
	if len(out) == 0 {
		return nil
	}
	return out
}

// SanitizeClientDescriptors filters out invalid client descriptors and assigns provider type
func SanitizeClientDescriptors(provider types.ProviderType, descriptors []types.ClientDescriptor) []types.ClientDescriptor {
	if len(descriptors) == 0 {
		return nil
	}

	out := lo.FilterMap(descriptors, func(descriptor types.ClientDescriptor, _ int) (types.ClientDescriptor, bool) {
		if descriptor.Build == nil {
			return types.ClientDescriptor{}, false
		}
		if descriptor.Provider == types.ProviderUnknown {
			descriptor.Provider = provider
		}
		return descriptor, true
	})
	if len(out) == 0 {
		return nil
	}
	return out
}
