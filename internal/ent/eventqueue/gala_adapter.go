package eventqueue

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/mutations"
	"github.com/theopenlane/core/pkg/gala"
)

// MutationGalaPayload is the JSON-safe mutation payload contract for gala
type MutationGalaPayload struct {
	// MutationType is the ent schema type that emitted the mutation
	MutationType string `json:"mutation_type"`
	// Operation is the mutation operation string
	Operation string `json:"operation"`
	// EntityID is the mutated entity identifier
	EntityID string `json:"entity_id"`
	// ChangedFields captures updated/cleared fields for the mutation
	ChangedFields []string `json:"changed_fields,omitempty"`
	// ClearedFields captures fields explicitly cleared by the mutation
	ClearedFields []string `json:"cleared_fields,omitempty"`
	// ChangedEdges captures changed edge names for workflow-eligible edges
	ChangedEdges []string `json:"changed_edges,omitempty"`
	// AddedIDs captures edge IDs added by edge name
	AddedIDs map[string][]string `json:"added_ids,omitempty"`
	// RemovedIDs captures edge IDs removed by edge name
	RemovedIDs map[string][]string `json:"removed_ids,omitempty"`
	// ProposedChanges captures field-level proposed values
	ProposedChanges map[string]any `json:"proposed_changes,omitempty"`
}

// ChangeSet returns the payload mutation deltas as a shared change-set contract
func (payload MutationGalaPayload) ChangeSet() mutations.ChangeSet {
	return mutations.NewChangeSet(payload.ChangedFields, payload.ChangedEdges, payload.AddedIDs, payload.RemovedIDs, payload.ProposedChanges)
}

// SetChangeSet applies a shared change-set contract onto this payload
func (payload *MutationGalaPayload) SetChangeSet(changeSet mutations.ChangeSet) {
	if payload == nil {
		return
	}

	cloned := changeSet.Clone()
	payload.ChangedFields = cloned.ChangedFields
	payload.ChangedEdges = cloned.ChangedEdges
	payload.AddedIDs = cloned.AddedIDs
	payload.RemovedIDs = cloned.RemovedIDs
	payload.ProposedChanges = cloned.ProposedChanges
}

// MutationGalaMetadata captures envelope metadata for Gala mutation dispatch
type MutationGalaMetadata struct {
	// EventID is the stable idempotency/event identifier for this mutation dispatch
	EventID string
	// Properties stores string-safe metadata for listener header fallbacks
	Properties map[string]string
}

// NewMutationGalaMetadata builds metadata for Gala mutation envelopes from mutation payload data
func NewMutationGalaMetadata(eventID string, payload MutationGalaPayload) MutationGalaMetadata {
	properties := mutationMetadataProperties(payload)
	entityID := strings.TrimSpace(payload.EntityID)
	if entityID != "" {
		if properties == nil {
			properties = map[string]string{}
		}
		properties[MutationPropertyEntityID] = entityID
	}

	return MutationGalaMetadata{
		EventID:    strings.TrimSpace(eventID),
		Properties: properties,
	}
}

// NewGalaHeadersFromMutationMetadata builds Gala headers from mutation metadata
func NewGalaHeadersFromMutationMetadata(metadata MutationGalaMetadata) gala.Headers {
	properties := normalizeMutationMetadataProperties(metadata.Properties)
	eventID := strings.TrimSpace(metadata.EventID)

	return gala.Headers{
		IdempotencyKey: eventID,
		Properties:     properties,
	}
}

// NewMutationGalaEnvelope builds a gala envelope from mutation emit inputs
func NewMutationGalaEnvelope(ctx context.Context, g *gala.Gala, topic gala.Topic[MutationGalaPayload], payload MutationGalaPayload, metadata MutationGalaMetadata) (envelope gala.Envelope, err error) {
	headers := NewGalaHeadersFromMutationMetadata(metadata)

	encodedPayload, err := g.Registry().EncodePayload(topic.Name, payload)
	if err != nil {
		return envelope, err
	}

	snapshot, err := g.ContextManager().Capture(ctx)
	if err != nil {
		return envelope, err
	}

	eventID := gala.EventID(headers.IdempotencyKey)
	if eventID == "" {
		eventID = gala.NewEventID()
		headers.IdempotencyKey = string(eventID)
	}

	envelope = gala.Envelope{
		ID:              eventID,
		Topic:           topic.Name,
		OccurredAt:      time.Now().UTC(),
		Headers:         headers,
		Payload:         encodedPayload,
		ContextSnapshot: snapshot,
	}

	return envelope, nil
}

// mutationMetadataProperties builds listener fallback properties from payload proposed changes
// Returns raw properties; normalization is deferred to NewGalaHeadersFromMutationMetadata
func mutationMetadataProperties(payload MutationGalaPayload) map[string]string {
	properties := lo.PickBy(lo.MapEntries(payload.ProposedChanges, func(key string, value any) (string, string) {
		stringValue, ok := ValueAsString(value)
		if !ok {
			return "", ""
		}

		return strings.TrimSpace(key), stringValue
	}), func(key string, _ string) bool {
		return key != ""
	})

	if len(properties) > 0 {
		return properties
	}

	return nil
}

// normalizeMutationMetadataProperties normalizes mutation metadata keys and values for Gala headers
func normalizeMutationMetadataProperties(properties map[string]string) map[string]string {
	if len(properties) == 0 {
		return nil
	}

	normalized := lo.PickBy(lo.MapEntries(properties, func(key, value string) (string, string) {
		return strings.TrimSpace(key), value
	}), func(key string, _ string) bool {
		return key != ""
	})

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}
