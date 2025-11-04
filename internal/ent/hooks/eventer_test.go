package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func TestAddMutationListenerRegistersBindings(t *testing.T) {
	eventer := NewEventer()

	pre := func(*soiree.EventContext) error { return nil }
	post := func(*soiree.EventContext) error { return nil }

	eventer.AddMutationListener(
		"Entity",
		func(*soiree.EventContext, *MutationPayload) error { return nil },
		soiree.WithPreHooks(pre),
		soiree.WithPostHooks(post),
	)

	if len(eventer.listeners) != 1 {
		t.Fatalf("expected stored mutation listener, got %v", eventer.listeners)
	}
	if listener := eventer.listeners[0]; listener.entity != "Entity" {
		t.Fatalf("unexpected entity stored: %s", listener.entity)
	}

	if _, ok := eventer.entities["Entity"]; !ok {
		t.Fatal("expected entity to be tracked")
	}
}

func TestEmitEventOnHonoursRegisteredEntities(t *testing.T) {
	eventer := NewEventer()
	eventer.AddMutationListener("Entity", func(*soiree.EventContext, *MutationPayload) error { return nil })

	predicate := eventer.emitEventOn()

	ctx := context.Background()

	if !predicate(ctx, &fakeMutation{typ: "Entity"}) {
		t.Fatal("expected predicate to allow emission for registered entity")
	}

	if predicate(ctx, &fakeMutation{typ: "Other"}) {
		t.Fatal("expected predicate to block emission for unregistered entity")
	}
}

type fakeMutation struct {
	typ    string
	op     ent.Op
	fields map[string]any
}

func (m *fakeMutation) Type() string { return m.typ }

func (m *fakeMutation) Op() ent.Op { return m.op }

func (m *fakeMutation) Fields() []string {
	result := make([]string, 0, len(m.fields))
	for key := range m.fields {
		result = append(result, key)
	}

	return result
}

func (m *fakeMutation) Field(name string) (ent.Value, bool) {
	val, ok := m.fields[name]
	if !ok {
		return nil, false
	}

	return val, true
}

// The remaining methods satisfy the entgen.Mutation interface but are unused in these tests.

func (m *fakeMutation) ID() (ent.Value, bool)                               { return nil, false }
func (m *fakeMutation) OldValue(context.Context, string) (ent.Value, error) { return nil, nil }
func (m *fakeMutation) ClearedFields() []string                             { return nil }
func (m *fakeMutation) AddedEdges() []string                                { return nil }
func (m *fakeMutation) RemovedEdges() []string                              { return nil }
func (m *fakeMutation) ClearedEdges() []string                              { return nil }
func (m *fakeMutation) EdgeCleared(string) bool                             { return false }
func (m *fakeMutation) ClearedFieldsExist() bool                            { return false }
func (m *fakeMutation) ResetField(string) error                             { return nil }
func (m *fakeMutation) ResetEdge(string) error                              { return nil }
func (m *fakeMutation) AddedIDs(string) []ent.Value                         { return nil }
func (m *fakeMutation) RemovedIDs(string) []ent.Value                       { return nil }
func (m *fakeMutation) Cleared(string) bool                                 { return false }
func (m *fakeMutation) SetField(string, ent.Value) error                    { return nil }
func (m *fakeMutation) AddedEIDs(string) []ent.Value                        { return nil }
func (m *fakeMutation) RemovedEIDs(string) []ent.Value                      { return nil }
func (m *fakeMutation) OldIDs(context.Context, string) ([]ent.Value, error) {
	return nil, nil
}
func (m *fakeMutation) AddField(string, ent.Value) error { return nil }
func (m *fakeMutation) AddEdge(string, ent.Value) error  { return nil }
func (m *fakeMutation) ClearField(string) error          { return nil }
func (m *fakeMutation) ClearEdge(string) error           { return nil }
func (m *fakeMutation) FieldsLen() int                   { return len(m.fields) }
func (m *fakeMutation) AddedFields() []string            { return nil }
func (m *fakeMutation) AddedField(string) (ent.Value, bool) {
	return nil, false
}
func (m *fakeMutation) FieldCleared(string) bool                            { return false }
func (m *fakeMutation) OldField(context.Context, string) (ent.Value, error) { return nil, nil }
