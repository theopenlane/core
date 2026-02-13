//go:build test

package hooks_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/riverqueue/river"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/events/soiree"
)

type listenerDispatchMode string

const (
	listenerDispatchModeInline listenerDispatchMode = "inline"
	listenerDispatchModeOutbox listenerDispatchMode = "outbox"
	testListenerAuthSubjectID  string               = "01ARZ3NDEKTSV4RRFFQ69G5FC9"
)

func TestTrustCenterMutationCacheRefreshParity(t *testing.T) {
	for _, mode := range []listenerDispatchMode{
		listenerDispatchModeInline,
		listenerDispatchModeOutbox,
	} {
		t.Run(string(mode), func(t *testing.T) {
			client := newOutboxTestClient(t, true)

			server, requestCount, freshValues := newCacheRefreshTestServer(t)
			configureTrustCenterCacheRefresh(t, server.URL)

			eventer := newParityEventer(t, client, mode)

			trustCenterID := "tc-parity-" + string(mode)
			insertTrustCenterRow(t, client, trustCenterID, "trust-center-"+string(mode))

			runMutationAndDispatchIfNeeded(t, client, eventer, mode, &fakeMutation{typ: entgen.TypeTrustCenter}, trustCenterID)

			require.Eventually(t, func() bool {
				return requestCount.Load() == 1
			}, 5*time.Second, 100*time.Millisecond)

			require.Equal(t, "1", firstFreshValue(t, freshValues))
		})
	}
}

func TestTrustCenterDocCreateCacheRefreshParity(t *testing.T) {
	for _, mode := range []listenerDispatchMode{
		listenerDispatchModeInline,
		listenerDispatchModeOutbox,
	} {
		t.Run(string(mode), func(t *testing.T) {
			client := newOutboxTestClient(t, true)

			server, requestCount, freshValues := newCacheRefreshTestServer(t)
			configureTrustCenterCacheRefresh(t, server.URL)

			eventer := newParityEventer(t, client, mode)

			trustCenterID := "tc-doc-parity-" + string(mode)
			insertTrustCenterRow(t, client, trustCenterID, "trust-center-doc-"+string(mode))

			mutation := &configuredMutation{
				typ: entgen.TypeTrustCenterDoc,
				op:  ent.OpCreate,
				fields: map[string]ent.Value{
					trustcenterdoc.FieldTrustCenterID: trustCenterID,
					trustcenterdoc.FieldVisibility:    enums.TrustCenterDocumentVisibilityProtected,
				},
			}

			runMutationAndDispatchIfNeeded(t, client, eventer, mode, mutation, "doc-parity-"+string(mode))

			require.Eventually(t, func() bool {
				return requestCount.Load() == 1
			}, 5*time.Second, 100*time.Millisecond)

			require.Equal(t, "1", firstFreshValue(t, freshValues))
		})
	}
}

func newParityEventer(t *testing.T, client *entgen.Client, mode listenerDispatchMode) *hooks.Eventer {
	t.Helper()

	opts := []hooks.EventerOpts{
		hooks.WithWorkflowListenersEnabled(false),
	}
	if mode == listenerDispatchModeOutbox {
		opts = append(opts,
			hooks.WithMutationOutboxEnabled(true),
			hooks.WithMutationOutboxFailOnEnqueueError(true),
		)
	}

	eventer := hooks.NewEventer(opts...)
	eventer.Initialize(client)

	require.NoError(t, hooks.RegisterListeners(eventer))

	return eventer
}

func runMutationAndDispatchIfNeeded(
	t *testing.T,
	client *entgen.Client,
	eventer *hooks.Eventer,
	mode listenerDispatchMode,
	mutation entgen.Mutation,
	eventID string,
) {
	t.Helper()

	baseCtx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:     testListenerAuthSubjectID,
		IsSystemAdmin: true,
	})

	mutator := hooks.EmitEventHook(eventer)(entgen.MutateFunc(func(_ context.Context, _ entgen.Mutation) (entgen.Value, error) {
		return &hooks.EventID{ID: eventID}, nil
	}))

	tx, err := client.Tx(baseCtx)
	require.NoError(t, err)

	txCtx := entgen.NewTxContext(baseCtx, tx)
	_, err = mutator.Mutate(txCtx, mutation)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	if mode != listenerDispatchModeOutbox {
		return
	}

	require.Eventually(t, func() bool {
		return countMutationDispatchJobs(t, client) == 1
	}, 5*time.Second, 100*time.Millisecond)

	args := latestMutationDispatchArgs(t, client)
	worker := eventqueue.NewMutationDispatchWorker(func() *soiree.EventBus {
		return eventer.Emitter
	})

	require.NoError(t, worker.Work(baseCtx, &river.Job[eventqueue.MutationDispatchArgs]{
		Args: args,
	}))
}

func configureTrustCenterCacheRefresh(t *testing.T, serverURL string) {
	t.Helper()

	parsedURL, err := url.Parse(serverURL)
	require.NoError(t, err)

	hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
		DefaultTrustCenterDomain: parsedURL.Host,
		CacheRefreshScheme:       "http",
	})

	t.Cleanup(func() {
		hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{})
	})
}

func newCacheRefreshTestServer(t *testing.T) (*httptest.Server, *atomic.Int32, chan string) {
	t.Helper()

	var requestCount atomic.Int32
	freshValues := make(chan string, 4)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		select {
		case freshValues <- r.URL.Query().Get("fresh"):
		default:
		}

		w.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(server.Close)

	return server, &requestCount, freshValues
}

func firstFreshValue(t *testing.T, freshValues chan string) string {
	t.Helper()

	select {
	case value := <-freshValues:
		return value
	case <-time.After(time.Second):
		t.Fatal("expected cache refresh request with fresh query parameter")
		return ""
	}
}

func insertTrustCenterRow(t *testing.T, client *entgen.Client, trustCenterID, slug string) {
	t.Helper()

	_, err := client.Job.GetPool().Exec(
		context.Background(),
		"INSERT INTO trust_centers (id, slug) VALUES ($1, $2)",
		trustCenterID,
		slug,
	)
	require.NoError(t, err)
}

func latestMutationDispatchArgs(t *testing.T, client *entgen.Client) eventqueue.MutationDispatchArgs {
	t.Helper()

	var raw string
	err := client.Job.GetPool().QueryRow(
		context.Background(),
		"SELECT args::text FROM river_job WHERE kind = $1 ORDER BY id DESC LIMIT 1",
		eventqueue.MutationDispatchJobKind,
	).Scan(&raw)
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	var args eventqueue.MutationDispatchArgs
	require.NoError(t, json.Unmarshal([]byte(raw), &args))

	return args
}

type configuredMutation struct {
	typ     string
	op      ent.Op
	fields  map[string]ent.Value
	cleared map[string]bool
}

func (m *configuredMutation) Op() ent.Op {
	if m == nil || m.op == 0 {
		return ent.OpUpdate
	}

	return m.op
}

func (m *configuredMutation) Type() string {
	if m == nil {
		return ""
	}

	return m.typ
}

func (m *configuredMutation) Fields() []string {
	if m == nil || len(m.fields) == 0 {
		return nil
	}

	return lo.Keys(m.fields)
}

func (m *configuredMutation) Field(name string) (ent.Value, bool) {
	if m == nil || m.fields == nil {
		return nil, false
	}

	value, ok := m.fields[name]
	return value, ok
}

func (m *configuredMutation) SetField(string, ent.Value) error {
	return nil
}

func (m *configuredMutation) AddedFields() []string {
	return nil
}

func (m *configuredMutation) AddedField(string) (ent.Value, bool) {
	return nil, false
}

func (m *configuredMutation) AddField(string, ent.Value) error {
	return nil
}

func (m *configuredMutation) ClearedFields() []string {
	if m == nil || len(m.cleared) == 0 {
		return nil
	}

	return lo.Keys(lo.PickBy(m.cleared, func(_ string, cleared bool) bool { return cleared }))
}

func (m *configuredMutation) FieldCleared(name string) bool {
	if m == nil || m.cleared == nil {
		return false
	}

	return m.cleared[name]
}

func (m *configuredMutation) ClearField(string) error {
	return nil
}

func (m *configuredMutation) ResetField(string) error {
	return nil
}

func (m *configuredMutation) AddedEdges() []string {
	return nil
}

func (m *configuredMutation) AddedIDs(string) []ent.Value {
	return nil
}

func (m *configuredMutation) RemovedEdges() []string {
	return nil
}

func (m *configuredMutation) RemovedIDs(string) []ent.Value {
	return nil
}

func (m *configuredMutation) ClearedEdges() []string {
	return nil
}

func (m *configuredMutation) EdgeCleared(string) bool {
	return false
}

func (m *configuredMutation) ClearEdge(string) error {
	return nil
}

func (m *configuredMutation) ResetEdge(string) error {
	return nil
}

func (m *configuredMutation) OldField(context.Context, string) (ent.Value, error) {
	return nil, nil
}
