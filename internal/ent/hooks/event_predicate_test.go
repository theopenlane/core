package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/require"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func TestEmitEventOnHonoursEmitterRegistrations(t *testing.T) {
	emitter := soiree.NewEventPool()
	t.Cleanup(func() {
		require.NoError(t, emitter.Close())
	})

	eventer := &Eventer{Emitter: emitter}

	_, err := emitter.On(entgen.TypeOrganization, func(_ *soiree.EventContext) error { return nil })
	require.NoError(t, err)

	predicate := eventer.emitEventOn()
	ctx := context.Background()
	mutation := &fakeMutation{typ: entgen.TypeOrganization}

	require.True(t, predicate(ctx, mutation))
}

type fakeMutation struct {
	typ string
}

func (m *fakeMutation) Op() ent.Op {
	return ent.OpUpdate
}

func (m *fakeMutation) Type() string {
	return m.typ
}

func (m *fakeMutation) Fields() []string {
	return nil
}

func (m *fakeMutation) Field(string) (ent.Value, bool) {
	return nil, false
}

func (m *fakeMutation) SetField(string, ent.Value) error {
	return nil
}

func (m *fakeMutation) AddedFields() []string {
	return nil
}

func (m *fakeMutation) AddedField(string) (ent.Value, bool) {
	return nil, false
}

func (m *fakeMutation) AddField(string, ent.Value) error {
	return nil
}

func (m *fakeMutation) ClearedFields() []string {
	return nil
}

func (m *fakeMutation) FieldCleared(string) bool {
	return false
}

func (m *fakeMutation) ClearField(string) error {
	return nil
}

func (m *fakeMutation) ResetField(string) error {
	return nil
}

func (m *fakeMutation) AddedEdges() []string {
	return nil
}

func (m *fakeMutation) AddedIDs(string) []ent.Value {
	return nil
}

func (m *fakeMutation) RemovedEdges() []string {
	return nil
}

func (m *fakeMutation) RemovedIDs(string) []ent.Value {
	return nil
}

func (m *fakeMutation) ClearedEdges() []string {
	return nil
}

func (m *fakeMutation) EdgeCleared(string) bool {
	return false
}

func (m *fakeMutation) ClearEdge(string) error {
	return nil
}

func (m *fakeMutation) ResetEdge(string) error {
	return nil
}

func (m *fakeMutation) OldField(context.Context, string) (ent.Value, error) {
	return nil, nil
}
