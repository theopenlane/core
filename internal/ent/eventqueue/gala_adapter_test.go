package eventqueue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/pkg/gala"
)

// galaAdapterTestActor is a fixture context value for snapshot capture validation
type galaAdapterTestActor struct {
	// ID identifies the fixture actor stored in context snapshots
	ID string `json:"id"`
}

type galaAdapterRuntime struct {
	registry       *gala.Registry
	contextManager *gala.ContextManager
}

// Registry returns the fixture registry for envelope encoding and decoding
func (r galaAdapterRuntime) Registry() *gala.Registry {
	return r.registry
}

// ContextManager returns the fixture context manager for snapshot capture and restore
func (r galaAdapterRuntime) ContextManager() *gala.ContextManager {
	return r.contextManager
}

// TestNewMutationGalaEnvelope verifies envelope creation from legacy mutation emit inputs
func TestNewMutationGalaEnvelope(t *testing.T) {
	t.Parallel()

	contextManager, err := gala.NewContextManager(
		gala.NewContextCodec(),
		gala.NewTypedContextCodec[galaAdapterTestActor]("adapter_actor"),
	)
	require.NoError(t, err)

	runtime := galaAdapterRuntime{
		registry:       gala.NewRegistry(),
		contextManager: contextManager,
	}

	topic := gala.Topic[MutationGalaPayload]{Name: gala.TopicName("mutation.organization")}
	err = gala.RegisterTopic(runtime.Registry(), gala.Registration[MutationGalaPayload]{
		Topic: topic,
		Codec: gala.JSONCodec[MutationGalaPayload]{},
	})
	require.NoError(t, err)

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
	emitCtx = auth.WithAuthenticatedUser(emitCtx, &auth.AuthenticatedUser{
		SubjectID:          "subject_123",
		OrganizationID:     "org_123",
		OrganizationRole:   auth.OwnerRole,
		AuthenticationType: auth.JWTAuthentication,
	})

	metadata := NewMutationGalaMetadata("evt_123", payload)
	envelope, err := NewMutationGalaEnvelope(emitCtx, runtime, topic, payload, metadata)
	require.NoError(t, err)

	require.Equal(t, gala.EventID("evt_123"), envelope.ID)
	require.Equal(t, topic.Name, envelope.Topic)
	require.Equal(t, "evt_123", envelope.Headers.IdempotencyKey)
	require.Equal(t, "name", envelope.Headers.Properties["mutation_field"])
	require.Equal(t, "7", envelope.Headers.Properties["count"])
	require.Equal(t, payload.EntityID, envelope.Headers.Properties[MutationPropertyEntityID])
	require.Equal(t, true, envelope.ContextSnapshot.Flags[gala.ContextFlagWorkflowBypass])
	require.Equal(t, true, envelope.ContextSnapshot.Flags[gala.ContextFlagWorkflowAllowEventEmission])
	require.Contains(t, envelope.ContextSnapshot.Values, gala.ContextKey("adapter_actor"))
	require.Contains(t, envelope.ContextSnapshot.Values, gala.ContextKey("durable"))

	restoredContext, err := runtime.ContextManager().Restore(context.Background(), envelope.ContextSnapshot)
	require.NoError(t, err)

	restoredUser, err := auth.GetAuthenticatedUserFromContext(restoredContext)
	require.NoError(t, err)
	require.Equal(t, "subject_123", restoredUser.SubjectID)
	require.Equal(t, "org_123", restoredUser.OrganizationID)

	decodedAny, err := runtime.Registry().DecodePayload(topic.Name, envelope.Payload)
	require.NoError(t, err)

	decoded, ok := decodedAny.(MutationGalaPayload)
	require.True(t, ok)
	require.Equal(t, payload.EntityID, decoded.EntityID)
	require.Equal(t, payload.Operation, decoded.Operation)
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

	require.Equal(t, "evt_456", headers.IdempotencyKey)
	require.Equal(t, "true", headers.Properties["active"])
	require.Equal(t, "5", headers.Properties["count"])
	_, exists := headers.Properties[""]
	require.False(t, exists)
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

	require.Equal(t, "status", payload.ChangedFields[0])
	require.Equal(t, "one", payload.AddedIDs["controls"][0])
	require.Equal(t, "approved", payload.ProposedChanges["status"])

	var roundTrip MutationGalaPayload
	roundTrip.SetChangeSet(payload.ChangeSet())
	require.Equal(t, payload.ChangedFields, roundTrip.ChangedFields)
	require.Equal(t, payload.ChangedEdges, roundTrip.ChangedEdges)
	require.Equal(t, payload.AddedIDs, roundTrip.AddedIDs)
	require.Equal(t, payload.RemovedIDs, roundTrip.RemovedIDs)
	require.Equal(t, payload.ProposedChanges, roundTrip.ProposedChanges)
}
