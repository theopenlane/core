package eventqueue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
)

// galaAdapterTestActor is a fixture context value for snapshot capture validation.
type galaAdapterTestActor struct {
	ID string `json:"id"`
}

type galaAdapterRuntime struct {
	registry       *gala.Registry
	contextManager *gala.ContextManager
}

func (r galaAdapterRuntime) Registry() *gala.Registry {
	return r.registry
}

func (r galaAdapterRuntime) ContextManager() *gala.ContextManager {
	return r.contextManager
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

	contextManager, err := gala.NewContextManager(
		gala.NewAuthContextCodec(),
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

	payload := &events.MutationPayload{
		MutationType:  "organization",
		Operation:     "UPDATE",
		EntityID:      "org_123",
		ChangedFields: []string{"name"},
		ProposedChanges: map[string]any{
			"mutation_field": "name",
			"count":          7,
		},
	}

	emitCtx := workflows.WithContext(context.Background())
	emitCtx = workflows.WithAllowWorkflowEventEmission(emitCtx)
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

// TestProjectGalaFlagsFromWorkflowContext verifies known workflow context markers
// are projected to Gala flags during envelope context capture.
func TestProjectGalaFlagsFromWorkflowContext(t *testing.T) {
	t.Parallel()

	projected := projectGalaFlagsFromWorkflowContext(
		workflows.WithAllowWorkflowEventEmission(workflows.WithContext(context.Background())),
	)

	require.True(t, gala.HasFlag(projected, gala.ContextFlagWorkflowBypass))
	require.True(t, gala.HasFlag(projected, gala.ContextFlagWorkflowAllowEventEmission))
}

// TestNewGalaHeadersFromMutationMetadata verifies property normalization for gala headers.
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
