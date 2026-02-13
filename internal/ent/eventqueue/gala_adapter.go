package eventqueue

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/events/soiree"
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

// NewGalaHeadersFromMutationProperties builds gala headers from soiree properties.
func NewGalaHeadersFromMutationProperties(props soiree.Properties) gala.Headers {
	properties := normalizeSoireeProperties(props)
	eventID := strings.TrimSpace(properties[soiree.PropertyEventID])

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
	props soiree.Properties,
) (gala.Envelope, error) {
	if runtime == nil {
		return gala.Envelope{}, gala.ErrRuntimeRequired
	}

	headers := NewGalaHeadersFromMutationProperties(props)
	galaPayload := NewMutationGalaPayload(payload)

	encodedPayload, schemaVersion, err := runtime.Registry().EncodePayload(ctx, topic.Name, galaPayload)
	if err != nil {
		return gala.Envelope{}, err
	}

	snapshot, err := runtime.ContextManager().Capture(ctx)
	if err != nil {
		return gala.Envelope{}, err
	}

	eventID := gala.EventID(headers.IdempotencyKey)
	if eventID == "" {
		eventID = gala.NewEventID()
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

// normalizeSoireeProperties converts soiree properties to string values for gala headers.
func normalizeSoireeProperties(props soiree.Properties) map[string]string {
	if len(props) == 0 {
		return nil
	}

	normalized := lo.MapEntries(props, func(key string, value any) (string, string) {
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
		return nil
	}

	return normalized
}
