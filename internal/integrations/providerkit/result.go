package providerkit

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// EncodeResult serializes an operation result and maps any encode failure to the caller-supplied error
func EncodeResult(value any, encodeErr error) (json.RawMessage, error) {
	raw, err := jsonx.ToRawMessage(value)
	if err != nil {
		return nil, encodeErr
	}

	return raw, nil
}

// MarshalEnvelope serializes a provider payload into one mapping envelope
func MarshalEnvelope(resource string, payload any, encodeErr error) (types.MappingEnvelope, error) {
	return MarshalEnvelopeVariant("", resource, payload, encodeErr)
}

// MarshalEnvelopeVariant serializes a provider payload into one mapping envelope for a specific variant
func MarshalEnvelopeVariant(variant string, resource string, payload any, encodeErr error) (types.MappingEnvelope, error) {
	raw, err := jsonx.ToRawMessage(payload)
	if err != nil {
		return types.MappingEnvelope{}, encodeErr
	}

	return RawEnvelopeVariant(variant, resource, raw), nil
}

// RawEnvelope wraps an already-serialized provider payload in a mapping envelope
func RawEnvelope(resource string, payload json.RawMessage) types.MappingEnvelope {
	return RawEnvelopeVariant("", resource, payload)
}

// RawEnvelopeVariant wraps an already-serialized provider payload in a variant-specific mapping envelope
func RawEnvelopeVariant(variant string, resource string, payload json.RawMessage) types.MappingEnvelope {
	return types.MappingEnvelope{
		Variant:  variant,
		Resource: resource,
		Payload:  payload,
	}
}
