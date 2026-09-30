package providerkit

import (
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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

// DirectoryAccountPayloadSets wraps directory account envelopes in the complete-snapshot account payload set
func DirectoryAccountPayloadSets(accounts []types.MappingEnvelope) []types.IngestPayloadSet {
	return []types.IngestPayloadSet{
		{
			Schema:           entityops.SchemaDirectoryAccount.Name,
			Envelopes:        accounts,
			SnapshotComplete: true,
		},
	}
}

// DirectoryGroupPayloadSets wraps directory group and membership envelopes in payload sets sharing one snapshot completeness flag
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
