package providerkit

import (
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
	"github.com/theopenlane/utils/mapx"
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

// RawEnvelopeWithFields wraps a serialized provider payload in a mapping envelope with the non-empty fields merged in,
// pruning empty values so mapping orValue fallbacks apply, mapping merge failure to the caller's error
func RawEnvelopeWithFields(resource string, payload json.RawMessage, fields map[string]any, encodeErr error) (types.MappingEnvelope, error) {
	patch, err := jsonx.ToRawMap(mapx.PruneMapZeroAny(fields))
	if err != nil {
		return types.MappingEnvelope{}, encodeErr
	}

	merged, _, err := jsonx.MergeObjectMap(payload, patch)
	if err != nil {
		return types.MappingEnvelope{}, encodeErr
	}

	return RawEnvelope(resource, merged), nil
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
