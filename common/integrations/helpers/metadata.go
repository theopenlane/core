package helpers

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
)

// StringValue extracts a string value from a map and returns it trimmed
func StringValue(data map[string]any, key string) string {
	if len(data) == 0 {
		return ""
	}

	value, ok := data[key]
	if !ok {
		return ""
	}

	return StringFromAny(value)
}

// FirstStringValue returns the first non-empty string for the provided keys.
func FirstStringValue(data map[string]any, keys ...string) string {
	if len(data) == 0 {
		return ""
	}
	for _, key := range keys {
		if value := StringValue(data, key); value != "" {
			return value
		}
	}
	return ""
}

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
