package hooks

import (
	"context"
	"sync"
	"testing"

	"entgo.io/ent"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

func TestAnyRuntimeInterestedMatchesConcernTopics(t *testing.T) {
	workflowRuntime, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create workflow runtime: %v", err)
	}

	notificationRuntime, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create notification runtime: %v", err)
	}

	workflowTopic := gala.Topic[entityops.MutationPayload]{
		Name: entityops.MutationTopicName(entityops.MutationConcernWorkflow, entgen.TypeTask),
	}
	if _, err := gala.Register(workflowRuntime,
		gala.Definition[entityops.MutationPayload]{
			Topic:      workflowTopic,
			Name:       "workflow.listener",
			Operations: []string{ent.OpCreate.String()},
			Handle: func(gala.HandlerContext, entityops.MutationPayload) error {
				return nil
			},
		},
	); err != nil {
		t.Fatalf("failed to register workflow listener: %v", err)
	}

	runtimes := []*gala.Gala{workflowRuntime, notificationRuntime}

	if !entityops.InterestedInMutation(runtimes, entgen.TypeTask, ent.OpCreate.String()) {
		t.Fatal("expected interest on the workflow concern topic")
	}

	if entityops.InterestedInMutation(runtimes, entgen.TypeTask, ent.OpDelete.String()) {
		t.Fatal("expected no interest for an unregistered operation")
	}

	if entityops.InterestedInMutation(runtimes, entgen.TypeControl, ent.OpCreate.String()) {
		t.Fatal("expected no interest for an unregistered schema")
	}
}

func TestMutationConcernTopics(t *testing.T) {
	topics := entityops.MutationConcernTopics(entgen.TypeTask)

	if topics[0] != entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeTask) {
		t.Fatalf("unexpected direct topic %q", topics[0])
	}

	if topics[1] != entityops.MutationTopicName(entityops.MutationConcernWorkflow, entgen.TypeTask) {
		t.Fatalf("unexpected workflow topic %q", topics[1])
	}

	if topics[2] != entityops.MutationTopicName(entityops.MutationConcernNotification, entgen.TypeTask) {
		t.Fatalf("unexpected notification topic %q", topics[2])
	}
}

// fakeMutation is a minimal ent.Mutation and utils.GenericMutation for driving EmitGalaEventHook without a database
type fakeMutation struct {
	op  ent.Op
	typ string
	id  string
	ids []string
}

func (m *fakeMutation) Op() ent.Op                                          { return m.op }
func (m *fakeMutation) Type() string                                        { return m.typ }
func (m *fakeMutation) Fields() []string                                    { return nil }
func (m *fakeMutation) Field(string) (ent.Value, bool)                      { return nil, false }
func (m *fakeMutation) SetField(string, ent.Value) error                    { return nil }
func (m *fakeMutation) AddedFields() []string                               { return nil }
func (m *fakeMutation) AddedField(string) (ent.Value, bool)                 { return nil, false }
func (m *fakeMutation) AddField(string, ent.Value) error                    { return nil }
func (m *fakeMutation) ClearedFields() []string                             { return nil }
func (m *fakeMutation) FieldCleared(string) bool                            { return false }
func (m *fakeMutation) ClearField(string) error                             { return nil }
func (m *fakeMutation) ResetField(string) error                             { return nil }
func (m *fakeMutation) AddedEdges() []string                                { return nil }
func (m *fakeMutation) AddedIDs(string) []ent.Value                         { return nil }
func (m *fakeMutation) RemovedEdges() []string                              { return nil }
func (m *fakeMutation) RemovedIDs(string) []ent.Value                       { return nil }
func (m *fakeMutation) ClearedEdges() []string                              { return nil }
func (m *fakeMutation) EdgeCleared(string) bool                             { return false }
func (m *fakeMutation) ClearEdge(string) error                              { return nil }
func (m *fakeMutation) ResetEdge(string) error                              { return nil }
func (m *fakeMutation) OldField(context.Context, string) (ent.Value, error) { return nil, nil }
func (m *fakeMutation) ID() (string, bool)                                  { return m.id, m.id != "" }
func (m *fakeMutation) IDs(context.Context) ([]string, error) {
	if len(m.ids) > 0 {
		return m.ids, nil
	}

	return []string{m.id}, nil
}
func (m *fakeMutation) Client() *entgen.Client { return nil }

// recordedMutation pairs a delivered payload with its envelope's soft-delete marker
type recordedMutation struct {
	payload    entityops.MutationPayload
	softDelete bool
}

// mutationRecorder collects the mutation payloads delivered to the catch-all test listeners
type mutationRecorder struct {
	mu     sync.Mutex
	events []recordedMutation
}

func (r *mutationRecorder) record(event recordedMutation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
}

func (r *mutationRecorder) recorded() []recordedMutation {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]recordedMutation(nil), r.events...)
}

// newRecordingRuntime returns an in-memory runtime recording every direct-topic emission
// through one regular and one soft-delete listener
func newRecordingRuntime(t *testing.T, schemaType string) (*gala.Gala, *mutationRecorder) {
	t.Helper()

	runtime, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	recorder := &mutationRecorder{}

	topic := entityops.MutationTopic(entityops.MutationConcernDirect, schemaType)
	// pass-through caller hook; the emitting test contexts carry no restored caller
	caller := func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
		return restored
	}
	handle := func(_ gala.HandlerContext, payload entityops.MutationPayload) error {
		recorder.record(recordedMutation{payload: payload, softDelete: payload.Operation == entityops.OpSoftDelete})

		return nil
	}

	if _, err := gala.Register(runtime,
		gala.Definition[entityops.MutationPayload]{
			Topic:      topic,
			Name:       "test.recorder",
			Operations: entityops.RegularMutationOps,
			Caller:     caller,
			Handle:     handle,
		},
		gala.Definition[entityops.MutationPayload]{
			Topic:      topic,
			Name:       "test.recorder.soft_delete",
			Operations: []string{entityops.OpSoftDelete},
			Caller:     caller,
			Handle:     handle,
		},
	); err != nil {
		t.Fatalf("failed to register recorder listeners: %v", err)
	}

	return runtime, recorder
}

// mutateSoftDelete re-enters the hook chain the way the soft-delete mixin does: rewrite the
// delete to an update, mark the entx soft-delete context, and mutate again
func mutateSoftDelete(t *testing.T, hook ent.Hook, mutation *fakeMutation, markSkip bool) {
	t.Helper()

	inner := hook(ent.MutateFunc(func(context.Context, ent.Mutation) (ent.Value, error) {
		return 1, nil
	}))

	mixin := ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		if markSkip {
			entityops.VetoEmission(ctx)
		}

		m.(*fakeMutation).op = ent.OpUpdate

		return inner.Mutate(entx.IsSoftDelete(ctx, m.Type()), m)
	})

	if _, err := hook(mixin).Mutate(context.Background(), mutation); err != nil {
		t.Fatalf("mutate failed: %v", err)
	}
}

func TestEmitGalaEventHookDirectSoftDeleteEmitsOnce(t *testing.T) {
	runtime, recorder := newRecordingRuntime(t, entgen.TypeTask)

	mutateSoftDelete(t, EmitGalaEventHook(runtime), &fakeMutation{op: ent.OpDeleteOne, typ: entgen.TypeTask, id: "task-1"}, false)
	runtime.WaitIdle()

	events := recorder.recorded()
	if len(events) != 1 {
		t.Fatalf("expected exactly one emission, got %d", len(events))
	}

	if events[0].payload.Operation != entityops.OpSoftDelete {
		t.Fatalf("expected %q operation, got %q", entityops.OpSoftDelete, events[0].payload.Operation)
	}

	if !events[0].softDelete {
		t.Fatal("expected the envelope to carry the soft-delete operation")
	}

	if events[0].payload.EntityID != "task-1" {
		t.Fatalf("expected entity id %q, got %q", "task-1", events[0].payload.EntityID)
	}
}

func TestEmitGalaEventHookOuterSkipSuppressesSoftDeleteEmission(t *testing.T) {
	runtime, recorder := newRecordingRuntime(t, entgen.TypeTask)

	mutateSoftDelete(t, EmitGalaEventHook(runtime), &fakeMutation{op: ent.OpDeleteOne, typ: entgen.TypeTask, id: "task-1"}, true)
	runtime.WaitIdle()

	if events := recorder.recorded(); len(events) != 0 {
		t.Fatalf("expected no emissions after outer skip, got %d", len(events))
	}
}

func TestEmitGalaEventHookHardDeleteEmits(t *testing.T) {
	runtime, recorder := newRecordingRuntime(t, entgen.TypeTask)

	hooked := EmitGalaEventHook(runtime)(ent.MutateFunc(func(context.Context, ent.Mutation) (ent.Value, error) {
		return 1, nil
	}))

	if _, err := hooked.Mutate(context.Background(), &fakeMutation{op: ent.OpDeleteOne, typ: entgen.TypeTask, id: "task-1"}); err != nil {
		t.Fatalf("mutate failed: %v", err)
	}

	runtime.WaitIdle()

	events := recorder.recorded()
	if len(events) != 1 {
		t.Fatalf("expected one emission for a hard delete, got %d", len(events))
	}

	if events[0].payload.Operation != ent.OpDeleteOne.String() {
		t.Fatalf("expected %q operation, got %q", ent.OpDeleteOne.String(), events[0].payload.Operation)
	}

	if events[0].payload.EntityID != "task-1" {
		t.Fatalf("expected entity id %q, got %q", "task-1", events[0].payload.EntityID)
	}

	if events[0].softDelete {
		t.Fatal("expected no soft-delete marker on a hard delete")
	}
}

func TestEmitGalaEventHookVetoedCascadeDeleteEmitsNothing(t *testing.T) {
	runtime, recorder := newRecordingRuntime(t, entgen.TypeTask)

	hooked := EmitGalaEventHook(runtime)(ent.MutateFunc(func(context.Context, ent.Mutation) (ent.Value, error) {
		return 1, nil
	}))

	ctx := entityops.WithEmissionVetoed(entx.SkipSoftDelete(context.Background()))
	if _, err := hooked.Mutate(ctx, &fakeMutation{op: ent.OpDeleteOne, typ: entgen.TypeTask, id: "task-1"}); err != nil {
		t.Fatalf("mutate failed: %v", err)
	}

	runtime.WaitIdle()

	if events := recorder.recorded(); len(events) != 0 {
		t.Fatalf("expected no emissions for a vetoed cascade delete, got %d", len(events))
	}
}

func TestEmitGalaEventHookBulkUpdateEmitsPerRow(t *testing.T) {
	runtime, recorder := newRecordingRuntime(t, entgen.TypeTask)

	hooked := EmitGalaEventHook(runtime)(ent.MutateFunc(func(context.Context, ent.Mutation) (ent.Value, error) {
		return 2, nil
	}))

	if _, err := hooked.Mutate(context.Background(), &fakeMutation{op: ent.OpUpdate, typ: entgen.TypeTask, ids: []string{"task-1", "task-2"}}); err != nil {
		t.Fatalf("mutate failed: %v", err)
	}

	runtime.WaitIdle()

	events := recorder.recorded()
	if len(events) != 2 {
		t.Fatalf("expected one emission per mutated row, got %d", len(events))
	}

	seen := map[string]bool{}
	for _, event := range events {
		if event.payload.Operation != ent.OpUpdate.String() {
			t.Fatalf("expected %q operation, got %q", ent.OpUpdate.String(), event.payload.Operation)
		}

		seen[event.payload.EntityID] = true
	}

	if !seen["task-1"] || !seen["task-2"] {
		t.Fatalf("expected envelopes for both rows, got %v", seen)
	}
}

func TestEmitGalaEventHookUpdateNeverClassifiesSoftDelete(t *testing.T) {
	runtime, recorder := newRecordingRuntime(t, entgen.TypeTask)

	hooked := EmitGalaEventHook(runtime)(ent.MutateFunc(func(context.Context, ent.Mutation) (ent.Value, error) {
		return &struct {
			ID string `json:"id"`
		}{ID: "task-1"}, nil
	}))

	contexts := map[string]context.Context{
		"plain":   context.Background(),
		"cascade": entx.SkipSoftDelete(context.Background()),
	}

	for name, ctx := range contexts {
		if _, err := hooked.Mutate(ctx, &fakeMutation{op: ent.OpUpdateOne, typ: entgen.TypeTask, id: "task-1"}); err != nil {
			t.Fatalf("%s mutate failed: %v", name, err)
		}
	}

	runtime.WaitIdle()

	events := recorder.recorded()
	if len(events) != len(contexts) {
		t.Fatalf("expected %d emissions, got %d", len(contexts), len(events))
	}

	for _, event := range events {
		if event.payload.Operation != ent.OpUpdateOne.String() {
			t.Fatalf("expected %q operation, got %q", ent.OpUpdateOne.String(), event.payload.Operation)
		}

		if event.softDelete {
			t.Fatal("expected no soft-delete marker on an update emission")
		}
	}
}
