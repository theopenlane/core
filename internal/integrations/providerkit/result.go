package providerkit

import (
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// EncodeResult serializes an operation result, mapping encode failure to the caller's error
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

// MarshalEnvelopeVariant serializes a provider payload into a variant-specific mapping envelope
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

// RawEnvelopeVariant wraps a serialized provider payload in a variant-specific mapping envelope
func RawEnvelopeVariant(variant string, resource string, payload json.RawMessage) types.MappingEnvelope {
	return types.MappingEnvelope{
		Variant:  variant,
		Resource: resource,
		Payload:  payload,
	}
}

// DirectoryAccountPayloadSets wraps directory account envelopes in a complete-snapshot payload set
func DirectoryAccountPayloadSets(accounts []types.MappingEnvelope) []types.IngestPayloadSet {
	return []types.IngestPayloadSet{
		{
			Schema:           entityops.SchemaDirectoryAccount.Name,
			Envelopes:        accounts,
			SnapshotComplete: true,
		},
	}
}

// DirectoryGroupPayloadSets wraps group and membership envelopes with one completeness flag
func DirectoryGroupPayloadSets(groups, memberships []types.MappingEnvelope, complete bool) []types.IngestPayloadSet {
	return []types.IngestPayloadSet{
		{
			Schema:           entityops.SchemaDirectoryGroup.Name,
			Envelopes:        groups,
			SnapshotComplete: complete,
		},
		{
			Schema:           entityops.SchemaDirectoryMembership.Name,
			Envelopes:        memberships,
			SnapshotComplete: complete,
		},
	}
}
