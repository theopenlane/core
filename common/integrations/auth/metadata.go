package auth

import (
	"maps"

	"github.com/theopenlane/core/common/integrations/types"
)

// ExtractMetadata decodes provider metadata from a credential payload into the target type.
func ExtractMetadata[T any](payload types.CredentialPayload) (T, error) {
	var result T
	data := payload.Data.ProviderData
	if len(data) == 0 {
		return result, nil
	}

	if err := DecodeProviderData(data, &result); err != nil {
		return result, err
	}

	return result, nil
}

// CloneMetadata creates a shallow copy of provider metadata, returning an empty map if nil.
func CloneMetadata(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}

	return maps.Clone(data)
}

// SetMetadataField sets a field in the metadata map only if the value is non-empty.
func SetMetadataField(meta map[string]any, key, value string) {
	if value != "" {
		meta[key] = value
	}
}
