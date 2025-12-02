package helpers

import (
	"fmt"
	"strings"

	"github.com/theopenlane/shared/integrations/types"
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

	return strings.TrimSpace(fmt.Sprint(value))
}

// SanitizeOperationDescriptors filters and cleans a slice of OperationDescriptor
func SanitizeOperationDescriptors(provider types.ProviderType, descriptors []types.OperationDescriptor) []types.OperationDescriptor {
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.Run == nil || descriptor.Name == "" {
			continue
		}

		if descriptor.Provider == types.ProviderUnknown {
			descriptor.Provider = provider
		}

		out = append(out, descriptor)
	}

	return out
}

// SanitizeClientDescriptors filters out invalid client descriptors and assigns provider type
func SanitizeClientDescriptors(provider types.ProviderType, descriptors []types.ClientDescriptor) []types.ClientDescriptor {
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.Build == nil {
			continue
		}

		if descriptor.Provider == types.ProviderUnknown {
			descriptor.Provider = provider
		}

		out = append(out, descriptor)
	}

	return out
}
