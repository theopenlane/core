package eventqueue

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/workflows"
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
func NewMutationGalaEnvelope(
	ctx context.Context,
	runtime *gala.Runtime,
	topic gala.Topic[MutationGalaPayload],
	payload *events.MutationPayload,
	metadata MutationGalaMetadata,
) (gala.Envelope, error) {
	if runtime == nil {
		return gala.Envelope{}, gala.ErrRuntimeRequired
	}

	headers := NewGalaHeadersFromMutationMetadata(metadata)
	galaPayload := NewMutationGalaPayload(payload)

	encodedPayload, schemaVersion, err := runtime.Registry().EncodePayload(topic.Name, galaPayload)
	if err != nil {
		return gala.Envelope{}, err
	}

	snapshot, err := runtime.ContextManager().Capture(projectGalaFlagsFromWorkflowContext(ctx))
	if err != nil {
		return gala.Envelope{}, err
	}

	eventID := gala.EventID(headers.IdempotencyKey)
	if eventID == "" {
		eventID = gala.NewEventID()
		headers.IdempotencyKey = string(eventID)
	}

	return gala.Envelope{
		ID:              eventID,
		Topic:           topic.Name,
		SchemaVersion:   schemaVersion,
		OccurredAt:      time.Now().UTC(),
		Headers:         headers,
		Payload:         encodedPayload,
		ContextSnapshot: snapshot,
	}, nil
}

// projectGalaFlagsFromWorkflowContext maps known workflow context markers into Gala context flags.
func projectGalaFlagsFromWorkflowContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if workflows.IsWorkflowBypass(ctx) {
		ctx = gala.WithFlag(ctx, gala.ContextFlagWorkflowBypass)
	}

	if workflows.AllowWorkflowEventEmission(ctx) {
		ctx = gala.WithFlag(ctx, gala.ContextFlagWorkflowAllowEventEmission)
	}

	return ctx
}

// mutationMetadataProperties builds listener fallback properties from payload proposed changes.
func mutationMetadataProperties(payload *events.MutationPayload) map[string]string {
	if payload == nil {
		return nil
	}

	normalized := lo.MapEntries(payload.ProposedChanges, func(key string, value any) (string, string) {
		stringValue, ok := events.ValueAsString(value)
		if !ok || strings.TrimSpace(key) == "" {
			return "", ""
		}

		return key, stringValue
	})

	normalized = lo.PickBy(normalized, func(key string, _ string) bool {
		return strings.TrimSpace(key) != ""
	})

	if len(normalized) == 0 {
		if payload.Mutation == nil {
			return nil
		}

		normalized = map[string]string{}
		for _, field := range payload.Mutation.Fields() {
			rawValue, ok := payload.Mutation.Field(field)
			if !ok || strings.TrimSpace(field) == "" {
				continue
			}

			stringValue, valueOK := events.ValueAsString(rawValue)
			if !valueOK {
				continue
			}

			normalized[field] = stringValue
		}
	}

	return normalizeMutationMetadataProperties(normalized)
}

// normalizeMutationMetadataProperties normalizes mutation metadata keys and values for Gala headers.
func normalizeMutationMetadataProperties(properties map[string]string) map[string]string {
	if len(properties) == 0 {
		return nil
	}

	normalized := lo.MapEntries(properties, func(key, value string) (string, string) {
		key = strings.TrimSpace(key)
		if key == "" {
			return "", ""
		}

		return key, value
	})
	normalized = lo.PickBy(normalized, func(key string, _ string) bool {
		return strings.TrimSpace(key) != ""
	})

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}
