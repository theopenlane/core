package eventqueue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/gala"
)

// galaAdapterTestActor is a fixture context value for snapshot capture validation.
type galaAdapterTestActor struct {
	ID string `json:"id"`
}

// TestNewMutationGalaPayload verifies mutation payload conversion into JSON-safe gala payload fields.
func TestNewMutationGalaPayload(t *testing.T) {
	t.Parallel()

	payload := &events.MutationPayload{
		MutationType:  "organization",
		Operation:     "UPDATE",
		EntityID:      "org_123",
		ChangedFields: []string{"name"},
		ClearedFields: []string{"description"},
		ChangedEdges:  []string{"delegate"},
		AddedIDs: map[string][]string{
			"delegate": {"user_1"},
		},
		RemovedIDs: map[string][]string{
			"delegate": {"user_2"},
		},
		ProposedChanges: map[string]any{
			"name":        "Acme",
			"description": nil,
		},
	}

	converted := NewMutationGalaPayload(payload)

	require.Equal(t, payload.MutationType, converted.MutationType)
	require.Equal(t, payload.Operation, converted.Operation)
	require.Equal(t, payload.EntityID, converted.EntityID)
	require.Equal(t, payload.ChangedFields, converted.ChangedFields)
	require.Equal(t, payload.ClearedFields, converted.ClearedFields)
	require.Equal(t, payload.ChangedEdges, converted.ChangedEdges)
	require.Equal(t, payload.AddedIDs, converted.AddedIDs)
	require.Equal(t, payload.RemovedIDs, converted.RemovedIDs)
	require.Equal(t, payload.ProposedChanges["name"], converted.ProposedChanges["name"])
	require.Contains(t, converted.ProposedChanges, "description")
	require.Nil(t, converted.ProposedChanges["description"])
}

// TestNewMutationGalaEnvelope verifies envelope creation from legacy mutation emit inputs.
func TestNewMutationGalaEnvelope(t *testing.T) {
	t.Parallel()

	contextManager, err := gala.NewContextManager(gala.NewTypedContextCodec[galaAdapterTestActor](gala.ContextKey("adapter_actor")))
	require.NoError(t, err)

	runtime, err := gala.NewRuntime(gala.RuntimeOptions{ContextManager: contextManager})
	require.NoError(t, err)

	topic := gala.Topic[MutationGalaPayload]{Name: gala.TopicName("mutation.organization")}
	err = (gala.Registration[MutationGalaPayload]{
		Topic: topic,
		Codec: gala.JSONCodec[MutationGalaPayload]{},
	}).Register(runtime.Registry())
	require.NoError(t, err)

	props := soiree.NewProperties()
	props.Set(soiree.PropertyEventID, "evt_123")
	props.Set("mutation_field", "name")
	props.Set("count", 7)

	payload := &events.MutationPayload{
		MutationType:  "organization",
		Operation:     "UPDATE",
		EntityID:      "org_123",
		ChangedFields: []string{"name"},
	}

	emitCtx := context.Background()
	emitCtx = gala.WithFlag(emitCtx, gala.ContextFlagWorkflowBypass)
	emitCtx = contextx.With(emitCtx, galaAdapterTestActor{ID: "actor_123"})

	envelope, err := NewMutationGalaEnvelope(emitCtx, runtime, topic, payload, props)
	require.NoError(t, err)

	require.Equal(t, gala.EventID("evt_123"), envelope.ID)
	require.Equal(t, topic.Name, envelope.Topic)
	require.Equal(t, "evt_123", envelope.Headers.IdempotencyKey)
	require.Equal(t, "name", envelope.Headers.Properties["mutation_field"])
	require.Equal(t, "7", envelope.Headers.Properties["count"])
	require.Equal(t, true, envelope.ContextSnapshot.Flags[gala.ContextFlagWorkflowBypass])
	require.Contains(t, envelope.ContextSnapshot.Values, gala.ContextKey("adapter_actor"))

	decodedAny, err := runtime.Registry().DecodePayload(context.Background(), topic.Name, envelope.Payload)
	require.NoError(t, err)

	decoded, ok := decodedAny.(MutationGalaPayload)
	require.True(t, ok)
	require.Equal(t, payload.EntityID, decoded.EntityID)
	require.Equal(t, payload.Operation, decoded.Operation)
}

// TestNewGalaHeadersFromMutationProperties verifies property normalization for gala headers.
func TestNewGalaHeadersFromMutationProperties(t *testing.T) {
	t.Parallel()

	props := soiree.NewProperties()
	props.Set(soiree.PropertyEventID, "evt_456")
	props.Set("active", true)
	props.Set("count", 5)
	props.Set("", "ignored")

	headers := NewGalaHeadersFromMutationProperties(props)

	require.Equal(t, "evt_456", headers.IdempotencyKey)
	require.Equal(t, "true", headers.Properties["active"])
	require.Equal(t, "5", headers.Properties["count"])
	_, exists := headers.Properties[""]
	require.False(t, exists)
}
