//go:build test

package hooks_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/testutils"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/enttest"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/gala"
)

const testMutationEventID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"

func TestEmitEventHookOutboxCommitEnqueuesJob(t *testing.T) {
	client := newOutboxTestClient(t, true)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
	)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		return nil
	})

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.Equal(t, 0, countMutationDispatchJobs(t, client))

	require.NoError(t, tx.Commit())
	require.Eventually(t, func() bool {
		return countMutationDispatchJobs(t, client) == 1
	}, 3*time.Second, 100*time.Millisecond)
}

func TestEmitEventHookOutboxRollbackSkipsEnqueue(t *testing.T) {
	client := newOutboxTestClient(t, true)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
	)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		return nil
	})

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)

	require.NoError(t, tx.Rollback())
	require.Equal(t, 0, countMutationDispatchJobs(t, client))
}

func TestEmitEventHookOutboxEnqueueFailureStrictFallsBackInline(t *testing.T) {
	client := newOutboxTestClient(t, false)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
	)

	listenerCalled := make(chan struct{}, 1)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		select {
		case listenerCalled <- struct{}{}:
		default:
		}

		return nil
	})
	require.NoError(t, hooks.RegisterListeners(eventer))

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	select {
	case <-listenerCalled:
	case <-time.After(3 * time.Second):
		t.Fatal("expected strict mode inline fallback listener emission when outbox enqueue fails")
	}
}

func TestEmitEventHookOutboxEnqueueFailureFallsBackWhenNotStrict(t *testing.T) {
	client := newOutboxTestClient(t, false)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(false),
	)

	listenerCalled := make(chan struct{}, 1)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		select {
		case listenerCalled <- struct{}{}:
		default:
		}

		return nil
	})
	require.NoError(t, hooks.RegisterListeners(eventer))

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	select {
	case <-listenerCalled:
	case <-time.After(3 * time.Second):
		t.Fatal("expected fallback inline listener emission")
	}
}

func TestEmitEventHookOutboxTopicFilterSkipsEnqueueAndFallsBackInline(t *testing.T) {
	client := newOutboxTestClient(t, true)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
		hooks.WithMutationOutboxTopics([]string{entgen.TypeUser}),
	)

	listenerCalled := make(chan struct{}, 1)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		select {
		case listenerCalled <- struct{}{}:
		default:
		}

		return nil
	})
	require.NoError(t, hooks.RegisterListeners(eventer))

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 0, countMutationDispatchJobs(t, client))

	select {
	case <-listenerCalled:
	case <-time.After(3 * time.Second):
		t.Fatal("expected inline listener emission when topic is not outbox-enabled")
	}
}

func TestEmitEventHookOutboxTopicFilterAllowsEnqueue(t *testing.T) {
	client := newOutboxTestClient(t, true)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
		hooks.WithMutationOutboxTopics([]string{entgen.TypeOrganization}),
	)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		return nil
	})

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Eventually(t, func() bool {
		return countMutationDispatchJobs(t, client) == 1
	}, 3*time.Second, 100*time.Millisecond)
}

func TestEmitEventHookGalaTopicModeV2OnlySkipsLegacyOnSuccess(t *testing.T) {
	client := newOutboxTestClient(t, true)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	dispatcher := &recordingGalaDispatcher{}
	runtime, err := gala.NewRuntime(gala.RuntimeOptions{
		DurableDispatcher: dispatcher,
	})
	require.NoError(t, err)

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
		hooks.WithGalaRuntimeProvider(func() *gala.Runtime { return runtime }),
		hooks.WithGalaDualEmitEnabled(true),
		hooks.WithGalaTopicModes(map[string]workflows.GalaTopicMode{
			entgen.TypeOrganization: workflows.GalaTopicModeV2Only,
		}),
	)

	listenerCalled := make(chan struct{}, 1)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		select {
		case listenerCalled <- struct{}{}:
		default:
		}

		return nil
	})
	require.NoError(t, hooks.RegisterListeners(eventer))

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Eventually(t, func() bool {
		return dispatcher.Count() == 1
	}, 3*time.Second, 100*time.Millisecond)
	require.Equal(t, 0, countMutationDispatchJobs(t, client))

	select {
	case <-listenerCalled:
		t.Fatal("expected no inline legacy listener execution for v2_only mode when gala dispatch succeeds")
	case <-time.After(500 * time.Millisecond):
	}
}

func TestEmitEventHookGalaTopicModeDualEmitDispatchesBothPaths(t *testing.T) {
	client := newOutboxTestClient(t, true)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	dispatcher := &recordingGalaDispatcher{}
	runtime, err := gala.NewRuntime(gala.RuntimeOptions{
		DurableDispatcher: dispatcher,
	})
	require.NoError(t, err)

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
		hooks.WithGalaRuntimeProvider(func() *gala.Runtime { return runtime }),
		hooks.WithGalaDualEmitEnabled(true),
		hooks.WithGalaTopicModes(map[string]workflows.GalaTopicMode{
			entgen.TypeOrganization: workflows.GalaTopicModeDualEmit,
		}),
	)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		return nil
	})

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Eventually(t, func() bool {
		return dispatcher.Count() == 1
	}, 3*time.Second, 100*time.Millisecond)
	require.Eventually(t, func() bool {
		return countMutationDispatchJobs(t, client) == 1
	}, 3*time.Second, 100*time.Millisecond)
}

func TestEmitEventHookGalaTopicModeV2OnlyFallsBackLegacyWhenGalaUnavailable(t *testing.T) {
	client := newOutboxTestClient(t, true)

	bus := soiree.New(soiree.Client(client))
	t.Cleanup(func() {
		require.NoError(t, bus.Close())
	})

	eventer := hooks.NewEventer(
		hooks.WithEventerEmitter(bus),
		hooks.WithWorkflowListenersEnabled(false),
		hooks.WithMutationOutboxEnabled(true),
		hooks.WithMutationOutboxFailOnEnqueueError(true),
		hooks.WithGalaRuntimeProvider(func() *gala.Runtime { return nil }),
		hooks.WithGalaDualEmitEnabled(true),
		hooks.WithGalaTopicModes(map[string]workflows.GalaTopicMode{
			entgen.TypeOrganization: workflows.GalaTopicModeV2Only,
		}),
	)
	eventer.AddMutationListener(entgen.TypeOrganization, func(_ *soiree.EventContext, _ *events.MutationPayload) error {
		return nil
	})

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: testMutationEventID}, nil
	}))

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)

	ctx := entgen.NewTxContext(context.Background(), tx)
	_, err = mutator.Mutate(ctx, &fakeMutation{typ: entgen.TypeOrganization})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Eventually(t, func() bool {
		return countMutationDispatchJobs(t, client) == 1
	}, 3*time.Second, 100*time.Millisecond)
}

func countMutationDispatchJobs(t *testing.T, client *entgen.Client) int {
	t.Helper()

	require.NotNil(t, client.Job)

	var count int
	err := client.Job.GetPool().QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM river_job WHERE kind = $1",
		eventqueue.MutationDispatchJobKind,
	).Scan(&count)
	require.NoError(t, err)

	return count
}

func newOutboxTestClient(t *testing.T, withJobClient bool) *entgen.Client {
	t.Helper()

	tf := entdb.NewTestFixture()

	db, err := sql.Open("postgres", tf.URI)
	require.NoError(t, err)

	opts := []entgen.Option{}
	if withJobClient {
		opts = append(opts,
			entgen.Job(context.Background(),
				riverqueue.WithConnectionURI(tf.URI),
				riverqueue.WithRunMigrations(true),
			),
		)
	}

	client := enttest.Open(t, dialect.Postgres, tf.URI,
		enttest.WithMigrateOptions(entdb.EnablePostgresOption(db)),
		enttest.WithOptions(opts...),
	)
	client.WithJobClient()

	t.Cleanup(func() {
		if client.Job != nil {
			require.NoError(t, client.Job.Close())
		}

		require.NoError(t, client.Close())
		require.NoError(t, db.Close())
		testutils.TeardownFixture(tf)
	})

	return client
}

type recordingGalaDispatcher struct {
	mu    sync.Mutex
	count int
}

func (d *recordingGalaDispatcher) DispatchDurable(_ context.Context, _ gala.Envelope, _ gala.TopicPolicy) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.count++

	return nil
}

func (d *recordingGalaDispatcher) Count() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.count
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
