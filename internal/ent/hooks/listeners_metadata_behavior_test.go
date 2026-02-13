package hooks

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func TestPayloadTouchesFields(t *testing.T) {
	t.Run("uses changed fields metadata", func(t *testing.T) {
		payload := &events.MutationPayload{
			ChangedFields: []string{"billing_email"},
		}

		assert.True(t, payloadTouchesFields(payload, "billing_email"))
		assert.False(t, payloadTouchesFields(payload, "billing_phone"))
	})

	t.Run("uses cleared fields metadata", func(t *testing.T) {
		payload := &events.MutationPayload{
			ClearedFields: []string{"billing_phone"},
		}

		assert.True(t, payloadTouchesFields(payload, "billing_phone"))
	})

	t.Run("falls back to mutation fields when metadata is absent", func(t *testing.T) {
		payload := &events.MutationPayload{
			Mutation: &fieldAwareMutation{
				fields: map[string]ent.Value{
					"billing_address": "123 Main",
				},
			},
		}

		assert.True(t, payloadTouchesFields(payload, "billing_address"))
	})

	t.Run("falls back to mutation cleared fields when metadata is absent", func(t *testing.T) {
		payload := &events.MutationPayload{
			Mutation: &fieldAwareMutation{
				cleared: map[string]bool{
					"billing_email": true,
				},
			},
		}

		assert.True(t, payloadTouchesFields(payload, "billing_email"))
	})

	t.Run("returns false for nil payload", func(t *testing.T) {
		assert.False(t, payloadTouchesFields(nil, "billing_email"))
	})
}

func TestHandleOrganizationSettingsUpdateOneSkipsWhenBillingFieldsUntouched(t *testing.T) {
	payload := &events.MutationPayload{
		ChangedFields: []string{"display_name"},
	}

	err := handleOrganizationSettingsUpdateOne(&soiree.EventContext{}, payload)
	require.NoError(t, err)
}

func TestMutationEntityIDResolution(t *testing.T) {
	t.Run("prefers payload entity id", func(t *testing.T) {
		ctx := eventContextWithProperties(t, map[string]any{"ID": "ctx-id"})

		id, ok := mutationEntityID(ctx, &events.MutationPayload{EntityID: "payload-id"})

		require.True(t, ok)
		assert.Equal(t, "payload-id", id)
	})

	t.Run("uses context string property when payload id is missing", func(t *testing.T) {
		ctx := eventContextWithProperties(t, map[string]any{"ID": "ctx-id"})

		id, ok := mutationEntityID(ctx, &events.MutationPayload{})

		require.True(t, ok)
		assert.Equal(t, "ctx-id", id)
	})

	t.Run("uses context stringer property when payload id is missing", func(t *testing.T) {
		ctx := eventContextWithProperties(t, map[string]any{"ID": testStringer("ctx-stringer-id")})

		id, ok := mutationEntityID(ctx, &events.MutationPayload{})

		require.True(t, ok)
		assert.Equal(t, "ctx-stringer-id", id)
	})

	t.Run("returns false when entity id is unavailable", func(t *testing.T) {
		id, ok := mutationEntityID(&soiree.EventContext{}, &events.MutationPayload{})

		require.False(t, ok)
		assert.Empty(t, id)
	})
}

func TestMutationStringFieldValueResolution(t *testing.T) {
	field := trustcenterdoc.FieldTrustCenterID

	t.Run("prefers proposed changes over context properties", func(t *testing.T) {
		ctx := eventContextWithProperties(t, map[string]any{field: "ctx-trust-center"})
		payload := &events.MutationPayload{
			ProposedChanges: map[string]any{
				field: "payload-trust-center",
			},
		}

		value := mutationStringFieldValue(ctx, payload, field)
		assert.Equal(t, "payload-trust-center", value)
	})

	t.Run("treats explicit nil proposed value as cleared field", func(t *testing.T) {
		ctx := eventContextWithProperties(t, map[string]any{field: "ctx-trust-center"})
		payload := &events.MutationPayload{
			ProposedChanges: map[string]any{
				field: nil,
			},
		}

		value := mutationStringFieldValue(ctx, payload, field)
		assert.Empty(t, value)
	})

	t.Run("falls back to context property", func(t *testing.T) {
		ctx := eventContextWithProperties(t, map[string]any{field: "ctx-trust-center"})

		value := mutationStringFieldValue(ctx, &events.MutationPayload{}, field)
		assert.Equal(t, "ctx-trust-center", value)
	})

	t.Run("falls back to context property stringer", func(t *testing.T) {
		ctx := eventContextWithProperties(t, map[string]any{field: testStringer("ctx-stringer")})

		value := mutationStringFieldValue(ctx, &events.MutationPayload{}, field)
		assert.Equal(t, "ctx-stringer", value)
	})
}

func TestShouldInvalidateCacheForTrustCenterDoc(t *testing.T) {
	tests := []struct {
		name    string
		payload *events.MutationPayload
		want    bool
	}{
		{
			name: "delete always invalidates",
			payload: &events.MutationPayload{
				Operation: ent.OpDeleteOne.String(),
			},
			want: true,
		},
		{
			name: "create publicly visible invalidates",
			payload: &events.MutationPayload{
				Operation: ent.OpCreate.String(),
				ProposedChanges: map[string]any{
					trustcenterdoc.FieldVisibility: string(enums.TrustCenterDocumentVisibilityPubliclyVisible),
				},
			},
			want: true,
		},
		{
			name: "create protected invalidates",
			payload: &events.MutationPayload{
				Operation: ent.OpCreate.String(),
				ProposedChanges: map[string]any{
					trustcenterdoc.FieldVisibility: string(enums.TrustCenterDocumentVisibilityProtected),
				},
			},
			want: true,
		},
		{
			name: "create not visible does not invalidate",
			payload: &events.MutationPayload{
				Operation: ent.OpCreate.String(),
				ProposedChanges: map[string]any{
					trustcenterdoc.FieldVisibility: string(enums.TrustCenterDocumentVisibilityNotVisible),
				},
			},
			want: false,
		},
		{
			name: "update visibility invalidates",
			payload: &events.MutationPayload{
				Operation:     ent.OpUpdateOne.String(),
				ChangedFields: []string{trustcenterdoc.FieldVisibility},
			},
			want: true,
		},
		{
			name: "update unrelated fields does not invalidate",
			payload: &events.MutationPayload{
				Operation:     ent.OpUpdateOne.String(),
				ChangedFields: []string{"title"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldInvalidateCacheForTrustCenterDoc(tt.payload))
		})
	}
}

func TestShouldInvalidateCacheForSubprocessor(t *testing.T) {
	t.Run("delete invalidates", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation: ent.OpDelete.String(),
		}

		assert.True(t, shouldInvalidateCacheForSubprocessor(payload))
	})

	t.Run("create with relevant metadata invalidates", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation:     ent.OpCreate.String(),
			ChangedFields: []string{subprocessor.FieldName},
		}

		assert.True(t, shouldInvalidateCacheForSubprocessor(payload))
	})

	t.Run("update with irrelevant metadata does not invalidate", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation:     ent.OpUpdate.String(),
			ChangedFields: []string{"description"},
		}

		assert.False(t, shouldInvalidateCacheForSubprocessor(payload))
	})

	t.Run("falls back to mutation fields when metadata is absent", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation: ent.OpUpdateOne.String(),
			Mutation: &fieldAwareMutation{
				fields: map[string]ent.Value{
					subprocessor.FieldLogoRemoteURL: "https://example.com/logo.png",
				},
			},
		}

		assert.True(t, shouldInvalidateCacheForSubprocessor(payload))
	})
}

func TestShouldInvalidateCacheForStandard(t *testing.T) {
	t.Run("delete invalidates", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation: ent.OpDeleteOne.String(),
		}

		assert.True(t, shouldInvalidateCacheForStandard(payload))
	})

	t.Run("update with relevant metadata invalidates", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation:     ent.OpUpdate.String(),
			ClearedFields: []string{standard.FieldLogoFileID},
		}

		assert.True(t, shouldInvalidateCacheForStandard(payload))
	})

	t.Run("update with irrelevant metadata does not invalidate", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation:     ent.OpUpdate.String(),
			ChangedFields: []string{"description"},
		}

		assert.False(t, shouldInvalidateCacheForStandard(payload))
	})

	t.Run("falls back to mutation cleared fields when metadata is absent", func(t *testing.T) {
		payload := &events.MutationPayload{
			Operation: ent.OpUpdateOne.String(),
			Mutation: &fieldAwareMutation{
				cleared: map[string]bool{
					standard.FieldLogoFileID: true,
				},
			},
		}

		assert.True(t, shouldInvalidateCacheForStandard(payload))
	})
}

type testStringer string

func (s testStringer) String() string {
	return string(s)
}

func eventContextWithProperties(t *testing.T, properties map[string]any) *soiree.EventContext {
	t.Helper()

	const topic = "hooks.listeners.metadata.test"
	bus := soiree.New()
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	ctxCh := make(chan *soiree.EventContext, 1)
	_, err := bus.On(topic, func(ctx *soiree.EventContext) error {
		ctxCh <- ctx

		return nil
	})
	require.NoError(t, err)

	event := soiree.NewBaseEvent(topic, "payload")
	props := soiree.NewProperties()

	for key, value := range properties {
		props.Set(key, value)
	}

	event.SetProperties(props)

	for emitErr := range bus.Emit(topic, event) {
		require.NoError(t, emitErr)
	}

	select {
	case ctx := <-ctxCh:
		return ctx
	case <-time.After(time.Second):
		t.Fatal("expected listener context")
		return nil
	}
}

type fieldAwareMutation struct {
	fields  map[string]ent.Value
	cleared map[string]bool
}

func (m *fieldAwareMutation) Op() ent.Op {
	return ent.OpUpdateOne
}

func (m *fieldAwareMutation) Type() string {
	return "TestMutation"
}

func (m *fieldAwareMutation) Fields() []string {
	if len(m.fields) == 0 {
		return nil
	}

	return lo.Keys(m.fields)
}

func (m *fieldAwareMutation) Field(name string) (ent.Value, bool) {
	if m.fields == nil {
		return nil, false
	}

	value, ok := m.fields[name]
	return value, ok
}

func (m *fieldAwareMutation) SetField(string, ent.Value) error {
	return nil
}

func (m *fieldAwareMutation) AddedFields() []string {
	return nil
}

func (m *fieldAwareMutation) AddedField(string) (ent.Value, bool) {
	return nil, false
}

func (m *fieldAwareMutation) AddField(string, ent.Value) error {
	return nil
}

func (m *fieldAwareMutation) ClearedFields() []string {
	if len(m.cleared) == 0 {
		return nil
	}

	return lo.Keys(lo.PickBy(m.cleared, func(_ string, cleared bool) bool { return cleared }))
}

func (m *fieldAwareMutation) FieldCleared(name string) bool {
	if m.cleared == nil {
		return false
	}

	return m.cleared[name]
}

func (m *fieldAwareMutation) ClearField(string) error {
	return nil
}

func (m *fieldAwareMutation) ResetField(string) error {
	return nil
}

func (m *fieldAwareMutation) AddedEdges() []string {
	return nil
}

func (m *fieldAwareMutation) AddedIDs(string) []ent.Value {
	return nil
}

func (m *fieldAwareMutation) RemovedEdges() []string {
	return nil
}

func (m *fieldAwareMutation) RemovedIDs(string) []ent.Value {
	return nil
}

func (m *fieldAwareMutation) ClearedEdges() []string {
	return nil
}

func (m *fieldAwareMutation) EdgeCleared(string) bool {
	return false
}

func (m *fieldAwareMutation) ClearEdge(string) error {
	return nil
}

func (m *fieldAwareMutation) ResetEdge(string) error {
	return nil
}

func (m *fieldAwareMutation) OldField(context.Context, string) (ent.Value, error) {
	return nil, nil
}
