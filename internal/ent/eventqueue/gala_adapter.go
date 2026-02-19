package eventqueue

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/gala"
)

// MutationGalaPayload is the JSON-safe mutation payload contract for gala.
type MutationGalaPayload struct {
	// MutationType is the ent schema type that emitted the mutation.
	MutationType string `json:"mutation_type"`
	// Operation is the mutation operation string.
	Operation string `json:"operation"`
	// EntityID is the mutated entity identifier.
	EntityID string `json:"entity_id"`
	// ChangedFields captures updated/cleared fields for the mutation.
	ChangedFields []string `json:"changed_fields,omitempty"`
	// ClearedFields captures fields explicitly cleared by the mutation.
	ClearedFields []string `json:"cleared_fields,omitempty"`
	// ChangedEdges captures changed edge names for workflow-eligible edges.
	ChangedEdges []string `json:"changed_edges,omitempty"`
	// AddedIDs captures edge IDs added by edge name.
	AddedIDs map[string][]string `json:"added_ids,omitempty"`
	// RemovedIDs captures edge IDs removed by edge name.
	RemovedIDs map[string][]string `json:"removed_ids,omitempty"`
	// ProposedChanges captures field-level proposed values.
	ProposedChanges map[string]any `json:"proposed_changes,omitempty"`
}

const (
	// mutationGalaPropertyEntityID is the standard mutation metadata key used for entity identifiers.
	mutationGalaPropertyEntityID = "ID"
)

// MutationGalaMetadata captures envelope metadata for Gala mutation dispatch.
type MutationGalaMetadata struct {
	// EventID is the stable idempotency/event identifier for this mutation dispatch.
	EventID string
	// Properties stores string-safe metadata for listener header fallbacks.
	Properties map[string]string
}

// galaEnvelopeRuntime captures the minimal Gala runtime surface needed for envelope construction.
type galaEnvelopeRuntime interface {
	Registry() *gala.Registry
	ContextManager() *gala.ContextManager
}

// NewMutationGalaPayload converts a mutation payload into a JSON-safe gala payload.
func NewMutationGalaPayload(payload *events.MutationPayload) MutationGalaPayload {
	if payload == nil {
		return MutationGalaPayload{}
	}

	return MutationGalaPayload{
		MutationType:    events.MutationType(payload),
		Operation:       payload.Operation,
		EntityID:        payload.EntityID,
		ChangedFields:   append([]string(nil), payload.ChangedFields...),
		ClearedFields:   append([]string(nil), payload.ClearedFields...),
		ChangedEdges:    append([]string(nil), payload.ChangedEdges...),
		AddedIDs:        events.CloneStringSliceMap(payload.AddedIDs),
		RemovedIDs:      events.CloneStringSliceMap(payload.RemovedIDs),
		ProposedChanges: events.CloneAnyMap(payload.ProposedChanges),
	}
}

// NewMutationGalaMetadata builds metadata for Gala mutation envelopes from mutation payload data.
func NewMutationGalaMetadata(eventID string, payload *events.MutationPayload) MutationGalaMetadata {
	properties := mutationMetadataProperties(payload)
	entityID := ""
	if payload != nil {
		entityID = strings.TrimSpace(payload.EntityID)
	}
	if entityID != "" {
		if properties == nil {
			properties = map[string]string{}
		}
		properties[mutationGalaPropertyEntityID] = entityID
	}

	return MutationGalaMetadata{
		EventID:    strings.TrimSpace(eventID),
		Properties: properties,
	}
}

// NewGalaHeadersFromMutationMetadata builds Gala headers from mutation metadata.
func NewGalaHeadersFromMutationMetadata(metadata MutationGalaMetadata) gala.Headers {
	properties := normalizeMutationMetadataProperties(metadata.Properties)
	eventID := strings.TrimSpace(metadata.EventID)

	return gala.Headers{
		IdempotencyKey: eventID,
		Properties:     properties,
	}
}

// NewMutationGalaEnvelope builds a gala envelope from legacy mutation emit inputs.
func NewMutationGalaEnvelope(ctx context.Context, g galaEnvelopeRuntime, topic gala.Topic[MutationGalaPayload], payload *events.MutationPayload, metadata MutationGalaMetadata) (envelope gala.Envelope, err error) {
	headers := NewGalaHeadersFromMutationMetadata(metadata)
	galaPayload := NewMutationGalaPayload(payload)

	encodedPayload, err := g.Registry().EncodePayload(topic.Name, galaPayload)
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

// mutationMetadataProperties builds listener fallback properties from payload proposed changes.
// Returns raw properties; normalization is deferred to NewGalaHeadersFromMutationMetadata.
func mutationMetadataProperties(payload *events.MutationPayload) map[string]string {
	if payload == nil {
		return nil
	}

	properties := lo.PickBy(lo.MapEntries(payload.ProposedChanges, func(key string, value any) (string, string) {
		stringValue, ok := events.ValueAsString(value)
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

// normalizeMutationMetadataProperties normalizes mutation metadata keys and values for Gala headers.
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
