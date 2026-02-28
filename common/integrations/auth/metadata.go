package auth

import (
	"maps"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
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

// PersistMetadata merges the JSON-tagged fields of meta into a clone of base.
// Fields tagged with omitempty are excluded when zero-valued.
func PersistMetadata[T any](base map[string]any, meta T) (map[string]any, error) {
	overlay, err := jsonx.ToMap(meta)
	if err != nil {
		return nil, err
	}

	out := maps.Clone(base)
	if out == nil {
		out = map[string]any{}
	}

	maps.Copy(out, overlay)

	return out, nil
}
