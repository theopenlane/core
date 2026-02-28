package eventqueue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/pkg/gala"
)

// galaAdapterTestActor is a fixture context value for snapshot capture validation
type galaAdapterTestActor struct {
	// ID identifies the fixture actor stored in context snapshots
	ID string `json:"id"`
}

// TestNewMutationGalaEnvelope verifies envelope creation from mutation emit inputs
func TestNewMutationGalaEnvelope(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
	})
	assert.NoError(t, err)

	t.Cleanup(func() { _ = runtime.Close() })

	assert.NoError(t, runtime.ContextManager().Register(
		gala.NewTypedContextCodec[galaAdapterTestActor]("adapter_actor"),
	))

	topic := gala.Topic[MutationGalaPayload]{Name: gala.TopicName("mutation.organization")}
	err = gala.RegisterTopic(runtime.Registry(), gala.Registration[MutationGalaPayload]{
		Topic: topic,
		Codec: gala.JSONCodec[MutationGalaPayload]{},
	})
	assert.NoError(t, err)

	payload := MutationGalaPayload{
		MutationType:  "organization",
		Operation:     "UPDATE",
		EntityID:      "org_123",
		ChangedFields: []string{"name"},
		ProposedChanges: map[string]any{
			"mutation_field": "name",
			"count":          7,
		},
	}

	emitCtx := gala.WithFlag(context.Background(), gala.ContextFlagWorkflowBypass)
	emitCtx = gala.WithFlag(emitCtx, gala.ContextFlagWorkflowAllowEventEmission)
	emitCtx = contextx.With(emitCtx, galaAdapterTestActor{ID: "actor_123"})
	emitCtx = auth.WithCaller(emitCtx, &auth.Caller{
		SubjectID:          "subject_123",
		OrganizationID:     "org_123",
		OrganizationRole:   auth.OwnerRole,
		AuthenticationType: auth.JWTAuthentication,
	})

	metadata := NewMutationGalaMetadata("evt_123", payload)
	envelope, err := NewMutationGalaEnvelope(emitCtx, runtime, topic, payload, metadata)
	assert.NoError(t, err)

	assert.Equal(t, gala.EventID("evt_123"), envelope.ID)
	assert.Equal(t, topic.Name, envelope.Topic)
	assert.Equal(t, "evt_123", envelope.Headers.IdempotencyKey)
	assert.Equal(t, "name", envelope.Headers.Properties["mutation_field"])
	assert.Equal(t, "7", envelope.Headers.Properties["count"])
	assert.Equal(t, payload.EntityID, envelope.Headers.Properties[MutationPropertyEntityID])
	assert.Equal(t, true, envelope.ContextSnapshot.Flags[gala.ContextFlagWorkflowBypass])
	assert.Equal(t, true, envelope.ContextSnapshot.Flags[gala.ContextFlagWorkflowAllowEventEmission])
	assert.Contains(t, envelope.ContextSnapshot.Values, gala.ContextKey("adapter_actor"))
	assert.Contains(t, envelope.ContextSnapshot.Values, gala.ContextKey("durable"))

	restoredContext, err := runtime.ContextManager().Restore(context.Background(), envelope.ContextSnapshot)
	assert.NoError(t, err)

	restoredCaller, restoredOk := auth.CallerFromContext(restoredContext)
	assert.True(t, restoredOk)
	assert.Equal(t, "subject_123", restoredCaller.SubjectID)
	assert.Equal(t, "org_123", restoredCaller.OrganizationID)

	decodedAny, err := runtime.Registry().DecodePayload(topic.Name, envelope.Payload)
	assert.NoError(t, err)

	decoded, ok := decodedAny.(MutationGalaPayload)
	assert.True(t, ok)
	assert.Equal(t, payload.EntityID, decoded.EntityID)
	assert.Equal(t, payload.Operation, decoded.Operation)
}

// TestNewGalaHeadersFromMutationMetadata verifies property normalization for gala headers
func TestNewGalaHeadersFromMutationMetadata(t *testing.T) {
	t.Parallel()

	headers := NewGalaHeadersFromMutationMetadata(MutationGalaMetadata{
		EventID: "evt_456",
		Properties: map[string]string{
			"active": "true",
			"count":  "5",
			"":       "ignored",
		},
	})

	assert.Equal(t, "evt_456", headers.IdempotencyKey)
	assert.Equal(t, "true", headers.Properties["active"])
	assert.Equal(t, "5", headers.Properties["count"])
	_, exists := headers.Properties[""]
	assert.False(t, exists)
}

// TestMutationGalaPayloadChangeSetRoundTrip verifies payload change-set projections preserve values and clone maps/slices
func TestMutationGalaPayloadChangeSetRoundTrip(t *testing.T) {
	t.Parallel()

	payload := MutationGalaPayload{
		ChangedFields: []string{"status"},
		ChangedEdges:  []string{"controls"},
		AddedIDs: map[string][]string{
			"controls": {"one"},
		},
		RemovedIDs: map[string][]string{
			"controls": {"two"},
		},
		ProposedChanges: map[string]any{
			"status": "approved",
		},
	}

	changeSet := payload.ChangeSet()
	changeSet.ChangedFields[0] = "mutated"
	changeSet.AddedIDs["controls"][0] = "mutated"
	changeSet.ProposedChanges["status"] = "mutated"

	assert.Equal(t, "status", payload.ChangedFields[0])
	assert.Equal(t, "one", payload.AddedIDs["controls"][0])
	assert.Equal(t, "approved", payload.ProposedChanges["status"])

	var roundTrip MutationGalaPayload
	roundTrip.SetChangeSet(payload.ChangeSet())
	assert.Equal(t, payload.ChangedFields, roundTrip.ChangedFields)
	assert.Equal(t, payload.ChangedEdges, roundTrip.ChangedEdges)
	assert.Equal(t, payload.AddedIDs, roundTrip.AddedIDs)
	assert.Equal(t, payload.RemovedIDs, roundTrip.RemovedIDs)
	assert.Equal(t, payload.ProposedChanges, roundTrip.ProposedChanges)
}
