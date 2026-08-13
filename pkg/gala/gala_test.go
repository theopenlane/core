package gala

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/samber/do/v2"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

// runtimeTestPayload is a fixture payload used in runtime tests.
type runtimeTestPayload struct {
	Message string `json:"message"`
}

// runtimeTestActor is a fixture context value used in runtime tests.
type runtimeTestActor struct {
	ID string `json:"id"`
}

// runtimeTestActorKey is the context key used for runtimeTestActor codec fixtures.
var runtimeTestActorKey = contextx.NewKey[runtimeTestActor]()

// runtimeTestPayloadKey is the context key used for runtimeTestPayload codec fixtures.
var runtimeTestPayloadKey = contextx.NewKey[runtimeTestPayload]()

// runtimeTestFormatter is a fixture dependency used in runtime tests.
type runtimeTestFormatter struct {
	Prefix string
}

// runtimeTestDispatcher captures durable dispatch calls in tests.
type runtimeTestDispatcher struct {
	calls     int
	envelopes []Envelope
	err       error
}

// Dispatch records durable dispatch invocations.
func (d *runtimeTestDispatcher) Dispatch(_ context.Context, envelope Envelope) error {
	d.calls++
	d.envelopes = append(d.envelopes, envelope)

	return d.err
}

// newTestGala creates a gala instance with a mock dispatcher for unit tests.
// For integration tests with real PostgreSQL/River, use NewTestGala from test_helpers_test.go.
func newTestGala(t *testing.T, d dispatcher) *Gala {
	t.Helper()

	return newTestGalaInMemory(t, d)
}

// testCaller returns a minimal caller for dispatch tests
func testCaller() *auth.Caller {
	return &auth.Caller{SubjectID: "test_subject"}
}

// testCallerSnapshot captures a context snapshot carrying a minimal caller
func testCallerSnapshot(t *testing.T, runtime *Gala) ContextSnapshot {
	t.Helper()

	snapshot, err := runtime.contextManager.Capture(auth.WithCaller(context.Background(), testCaller()))
	if err != nil {
		t.Fatalf("failed to capture caller snapshot: %v", err)
	}

	return snapshot
}

// TestRuntimeDispatchEnvelopeWithDependencyInjectionAndContextRehydration verifies
// that handlers can resolve dependencies from samber/do and receive rehydrated context.
func TestRuntimeDispatchEnvelopeWithDependencyInjectionAndContextRehydration(t *testing.T) {
	injector := do.New()
	do.ProvideValue(injector, &runtimeTestFormatter{Prefix: "fmt"})

	contextManager, err := newContextManager(
		NewKeyCodec("runtime_test_actor", runtimeTestActorKey),
		NewKeyCodec("caller", auth.CallerKey),
	)
	if err != nil {
		t.Fatalf("failed to build context manager: %v", err)
	}

	runtime := newTestGala(t, nil)
	runtime.injector = injector
	runtime.contextManager = contextManager

	topic := Topic[runtimeTestPayload]{
		Name: TopicName("runtime.test.event"),
	}

	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var observed string

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.listener",
		Handle: func(handlerContext HandlerContext, payload runtimeTestPayload) error {
			formatter := do.MustInvoke[*runtimeTestFormatter](handlerContext.Injector)

			actor, exists := runtimeTestActorKey.Get(handlerContext.Context)
			if !exists {
				return errors.New("missing rehydrated actor context")
			}

			if !HasFlag(handlerContext.Context, ContextFlagWorkflowBypass) {
				return errors.New("missing rehydrated workflow bypass flag")
			}

			observed = formatter.Prefix + ":" + payload.Message + ":" + actor.ID

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	emitContext := runtimeTestActorKey.Set(context.Background(), runtimeTestActor{ID: "actor-1"})
	emitContext = auth.WithCaller(emitContext, testCaller())
	emitContext = WithFlag(emitContext, ContextFlagWorkflowBypass)

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "hello"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	snapshot, err := runtime.contextManager.Capture(emitContext)
	if err != nil {
		t.Fatalf("failed to capture context snapshot: %v", err)
	}

	if err := runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: snapshot,
	}); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if observed != "fmt:hello:actor-1" {
		t.Fatalf("unexpected listener output: %s", observed)
	}
}

// TestRuntimeDispatchEnvelopeWithCallerContext verifies default runtime codecs
// rehydrate auth context values for listener execution.
func TestRuntimeDispatchEnvelopeWithCallerContext(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{
		Name: TopicName("runtime.test.auth"),
	}

	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var observed *auth.Caller
	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.auth.listener",
		Handle: func(handlerContext HandlerContext, _ runtimeTestPayload) error {
			caller, ok := auth.CallerFromContext(handlerContext.Context)
			if !ok || caller == nil {
				return auth.ErrNoAuthUser
			}

			observed = caller
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	emitContext := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:          "subject_123",
		SubjectName:        "Codex User",
		SubjectEmail:       "codex@example.com",
		OrganizationID:     "org_123",
		OrganizationName:   "Acme Corp",
		OrganizationIDs:    []string{"org_123", "org_234"},
		AuthenticationType: auth.JWTAuthentication,
		OrganizationRole:   auth.OwnerRole,
		ActiveSubscription: true,
		Capabilities:       auth.CapSystemAdmin,
		Impersonation: &auth.ImpersonationContext{
			Type:              auth.AdminImpersonation,
			ImpersonatorID:    "admin_123",
			ImpersonatorEmail: "admin@example.com",
			TargetUserID:      "subject_123",
			TargetUserEmail:   "codex@example.com",
			Reason:            "support",
		},
		OriginalSystemAdmin: &auth.Caller{
			SubjectID:    "admin_123",
			SubjectEmail: "admin@example.com",
			Capabilities: auth.CapSystemAdmin,
		},
	})

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "auth"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	snapshot, err := runtime.contextManager.Capture(emitContext)
	if err != nil {
		t.Fatalf("failed to capture context snapshot: %v", err)
	}

	if err := runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: snapshot,
	}); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if observed.SubjectID != "subject_123" {
		t.Fatalf("unexpected subject id %q", observed.SubjectID)
	}

	if observed.OrganizationID != "org_123" {
		t.Fatalf("unexpected organization id %q", observed.OrganizationID)
	}

	if observed.OrganizationRole != auth.OwnerRole {
		t.Fatalf("unexpected organization role %q", observed.OrganizationRole)
	}

	if !observed.Has(auth.CapSystemAdmin) {
		t.Fatalf("expected system admin flag to be true")
	}

	if !observed.ActiveSubscription {
		t.Fatalf("expected active subscription flag to round-trip")
	}

	if observed.Impersonation == nil || observed.Impersonation.ImpersonatorID != "admin_123" {
		t.Fatalf("expected impersonation context to round-trip")
	}

	if observed.OriginalSystemAdmin == nil || observed.OriginalSystemAdmin.SubjectID != "admin_123" {
		t.Fatalf("expected original system admin lineage to round-trip")
	}
}

// TestAttachListenerRequiresTopicRegistration verifies listener registration
// fails when the topic contract has not been registered.
func TestAttachListenerRequiresTopicRegistration(t *testing.T) {
	runtime := newTestGala(t, nil)

	_, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: Topic[runtimeTestPayload]{Name: TopicName("missing.topic")},
		Name:  "runtime.test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			return nil
		},
	})
	if !errors.Is(err, ErrListenerTopicNotRegistered) {
		t.Fatalf("expected ErrListenerTopicNotRegistered, got %v", err)
	}
}

// TestRuntimeDispatchEnvelopeReturnsDecodeError verifies malformed payload data
// returns a decode error before listener execution.
func TestRuntimeDispatchEnvelopeReturnsDecodeError(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.decode")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	err := runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: json.RawMessage("{bad"),
	})
	if !errors.Is(err, ErrPayloadDecodeFailed) {
		t.Fatalf("expected ErrPayloadDecodeFailed, got %v", err)
	}
}

// TestRuntimeEmitUsesDispatcher verifies emits enqueue through the durable dispatcher
// without invoking listeners on the emit call path.
func TestRuntimeEmitUsesDispatcher(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.durable")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	listenerCalls := 0
	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.durable.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			listenerCalls++

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	id, err := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "durable"}, Headers{Kind: JobKindSystem})
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	if id == "" {
		t.Fatalf("expected non-empty event id from emit")
	}

	if dispatcher.calls != 1 {
		t.Fatalf("expected 1 durable dispatch call, got %d", dispatcher.calls)
	}

	if listenerCalls != 0 {
		t.Fatalf("expected 0 listener calls, got %d", listenerCalls)
	}
}

// TestRuntimeEmitRequiresDispatcher verifies emit fails when runtime has no
// durable dispatcher configured.
func TestRuntimeEmitRequiresDispatcher(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.durable.required")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "durable"}, Headers{Kind: JobKindSystem}); !errors.Is(err, ErrDispatcherRequired) {
		t.Fatalf("expected ErrDispatcherRequired, got %v", err)
	}
}

// TestRuntimeEmitWithEventIDUsesPrebuiltEventID verifies WithEventID and WithRawPayload
// preserve the caller event identity and pre-encoded payload bytes.
func TestRuntimeEmitWithEventIDUsesPrebuiltEventID(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.envelope")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "prebuilt"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	prebuiltID := EventID("evt_prebuilt_123")
	id, err := runtime.EmitWithHeaders(context.Background(), topic.Name, nil, Headers{Kind: JobKindSystem}, WithEventID(prebuiltID), WithRawPayload(encodedPayload))
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	if id != prebuiltID {
		t.Fatalf("expected emit to return prebuilt event id %q, got %q", prebuiltID, id)
	}

	if dispatcher.calls != 1 {
		t.Fatalf("expected one durable dispatch, got %d", dispatcher.calls)
	}

	if len(dispatcher.envelopes) != 1 {
		t.Fatalf("expected one recorded envelope, got %d", len(dispatcher.envelopes))
	}

	if dispatcher.envelopes[0].ID != prebuiltID {
		t.Fatalf("expected preserved prebuilt event id %q, got %q", prebuiltID, dispatcher.envelopes[0].ID)
	}

	if string(dispatcher.envelopes[0].Payload) != string(encodedPayload) {
		t.Fatalf("expected raw payload to be preserved, got %s", string(dispatcher.envelopes[0].Payload))
	}
}

func TestNewGalaRequiresConnectionURI(t *testing.T) {
	_, err := NewGala(context.Background(), Config{})
	if !errors.Is(err, ErrRiverConnectionURIRequired) {
		t.Fatalf("expected ErrRiverConnectionURIRequired, got %v", err)
	}
}

func TestGalaWorkersRequireJobClient(t *testing.T) {
	runtime := &Gala{}

	if err := runtime.StartWorkers(context.Background()); !errors.Is(err, ErrRiverJobClientRequired) {
		t.Fatalf("expected ErrRiverJobClientRequired on start, got %v", err)
	}

	if err := runtime.StopWorkers(context.Background()); !errors.Is(err, ErrRiverJobClientRequired) {
		t.Fatalf("expected ErrRiverJobClientRequired on stop, got %v", err)
	}
}

type codecTestUnsupportedPayload struct {
	Unsupported func()
}

type runtimeOperationPayload struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
}

func TestJSONCodecErrorPaths(t *testing.T) {
	codec := JSONCodec[codecTestUnsupportedPayload]{}

	_, err := codec.Encode(codecTestUnsupportedPayload{
		Unsupported: func() {},
	})
	if !errors.Is(err, ErrPayloadEncodeFailed) {
		t.Fatalf("expected ErrPayloadEncodeFailed, got %v", err)
	}

	decodeCodec := JSONCodec[runtimeTestPayload]{}
	_, err = decodeCodec.Decode(nil)
	if !errors.Is(err, ErrEnvelopePayloadRequired) {
		t.Fatalf("expected ErrEnvelopePayloadRequired, got %v", err)
	}
}

func TestRuntimeDispatchEnvelopeWrapsListenerErrors(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.listener.error")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	listenerErr := errors.New("listener failed")
	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "failing.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			return listenerErr
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "test"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: testCallerSnapshot(t, runtime),
	})
	if err == nil {
		t.Fatalf("expected error from failing listener")
	}

	var listenerError ListenerError
	if !errors.As(err, &listenerError) {
		t.Fatalf("expected ListenerError, got %T", err)
	}

	if !strings.HasPrefix(listenerError.ListenerName, "failing.listener#") {
		t.Fatalf("expected listener name prefixed 'failing.listener#', got %q", listenerError.ListenerName)
	}

	if listenerError.Panicked {
		t.Fatalf("expected Panicked=false for non-panicking listener")
	}

	if !errors.Is(listenerError.Cause, listenerErr) {
		t.Fatalf("expected cause to be original error, got %v", listenerError.Cause)
	}
}

func TestRuntimeDispatchEnvelopeRecoversPanic(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.listener.panic")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "panicking.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			panic("listener panic")
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "test"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: testCallerSnapshot(t, runtime),
	})
	if err == nil {
		t.Fatalf("expected error from panicking listener")
	}

	var listenerError ListenerError
	if !errors.As(err, &listenerError) {
		t.Fatalf("expected ListenerError, got %T", err)
	}

	if !strings.HasPrefix(listenerError.ListenerName, "panicking.listener#") {
		t.Fatalf("expected listener name prefixed 'panicking.listener#', got %q", listenerError.ListenerName)
	}

	if !listenerError.Panicked {
		t.Fatalf("expected Panicked=true for panicking listener")
	}

	if !errors.Is(listenerError.Cause, ErrListenerPanicked) {
		t.Fatalf("expected cause to be ErrListenerPanicked, got %v", listenerError.Cause)
	}
}

func TestRuntimeDispatchEnvelopeFiltersListenersByOperation(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeOperationPayload]{Name: TopicName("runtime.test.listener.operation")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeOperationPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	createCalls := 0
	updateCalls := 0

	if _, err := attachListener(runtime, Definition[runtimeOperationPayload]{
		Topic:      topic,
		Name:       "create.listener",
		Operations: []string{ent.OpCreate.String()},
		Handle: func(HandlerContext, runtimeOperationPayload) error {
			createCalls++
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register create listener: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeOperationPayload]{
		Topic:      topic,
		Name:       "update.listener",
		Operations: []string{ent.OpUpdate.String(), ent.OpUpdateOne.String()},
		Handle: func(HandlerContext, runtimeOperationPayload) error {
			updateCalls++
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register update listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeOperationPayload{
		Operation: ent.OpUpdateOne.String(),
		Message:   "test",
	})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: testCallerSnapshot(t, runtime),
	})
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if createCalls != 0 {
		t.Fatalf("expected create listener not to run, got %d", createCalls)
	}

	if updateCalls != 1 {
		t.Fatalf("expected update listener to run once, got %d", updateCalls)
	}
}

func TestRuntimeEmitReturnsDurableDispatchError(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{err: errors.New("durable failed")}
	runtime := newTestGala(t, dispatcher)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.durable.error")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "durable.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "test"}, Headers{Kind: JobKindSystem}); !errors.Is(err, ErrDispatchFailed) {
		t.Fatalf("expected ErrDispatchFailed, got %v", err)
	}
}

func TestRegistryConcurrentRegistration(t *testing.T) {
	runtime := newTestGala(t, nil)

	const numGoroutines = 100

	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			topic := Topic[runtimeTestPayload]{Name: TopicName(fmt.Sprintf("topic.%d", n))}
			_ = registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{})
			_, _ = attachListener(runtime, Definition[runtimeTestPayload]{
				Topic:  topic,
				Name:   fmt.Sprintf("listener.%d", n),
				Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
			})
		}(i)
	}

	wg.Wait()
}

func TestListenerErrorErrorMethod(t *testing.T) {
	tests := []struct {
		name     string
		err      ListenerError
		expected string
	}{
		{
			name:     "panicked listener",
			err:      ListenerError{ListenerName: "test.listener", Panicked: true, Cause: ErrListenerPanicked},
			expected: `gala: listener "test.listener" panicked: gala: listener panicked`,
		},
		{
			name:     "non-panicked listener",
			err:      ListenerError{ListenerName: "test.listener", Panicked: false, Cause: errors.New("failed")},
			expected: `gala: listener "test.listener" execution failed: failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestListenerErrorUnwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := ListenerError{ListenerName: "test.listener", Cause: cause}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Fatalf("expected unwrap to return cause, got %v", unwrapped)
	}

	if !errors.Is(err, cause) {
		t.Fatalf("expected errors.Is to match cause")
	}
}

func TestListenerErrorUnwrapNilCause(t *testing.T) {
	err := ListenerError{ListenerName: "test.listener", Cause: nil}

	if unwrapped := err.Unwrap(); unwrapped != nil {
		t.Fatalf("expected unwrap to return nil, got %v", unwrapped)
	}
}

func TestRegistryInterestedInEmptyTopic(t *testing.T) {
	registry := newRegistry()

	if registry.InterestedIn("", "create") {
		t.Fatalf("expected false for empty topic")
	}
}

func TestRegistryInterestedInNoListeners(t *testing.T) {
	registry := newRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.no.listeners")}
	if err := registerTopic(registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if registry.InterestedIn(topic.Name, "create") {
		t.Fatalf("expected false when no listeners registered")
	}
}

func TestRegistryInterestedInEmptyOperation(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.empty.operation")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	if !runtime.registry.InterestedIn(topic.Name, "") {
		t.Fatalf("expected true for empty operation when listeners exist")
	}

	if !runtime.registry.InterestedIn(topic.Name, "   ") {
		t.Fatalf("expected true for whitespace-only operation when listeners exist")
	}
}

func TestRegistryInterestedInWithOperationFilter(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeOperationPayload]{Name: TopicName("test.operation.filter")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeOperationPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeOperationPayload]{
		Topic:      topic,
		Name:       "test.create.listener",
		Operations: []string{"create"},
		Handle:     func(HandlerContext, runtimeOperationPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	if !runtime.registry.InterestedIn(topic.Name, "create") {
		t.Fatalf("expected true for matching operation")
	}

	if runtime.registry.InterestedIn(topic.Name, "update") {
		t.Fatalf("expected false for non-matching operation")
	}
}

func TestRegistryInterestedInWithWildcardListener(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeOperationPayload]{Name: TopicName("test.wildcard.listener")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeOperationPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeOperationPayload]{
		Topic:  topic,
		Name:   "test.wildcard",
		Handle: func(HandlerContext, runtimeOperationPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	if !runtime.registry.InterestedIn(topic.Name, "create") {
		t.Fatalf("expected true for wildcard listener with any operation")
	}

	if !runtime.registry.InterestedIn(topic.Name, "update") {
		t.Fatalf("expected true for wildcard listener with any operation")
	}
}

func TestValidateTopicRegistrationErrors(t *testing.T) {
	if err := registerTopic(nil, Topic[runtimeTestPayload]{}, JSONCodec[runtimeTestPayload]{}); !errors.Is(err, ErrRegistryRequired) {
		t.Fatalf("expected ErrRegistryRequired, got %v", err)
	}

	registry := newRegistry()

	if err := registerTopic(registry, Topic[runtimeTestPayload]{Name: ""}, JSONCodec[runtimeTestPayload]{}); !errors.Is(err, ErrTopicNameRequired) {
		t.Fatalf("expected ErrTopicNameRequired, got %v", err)
	}

	if err := registerTopic(registry, Topic[runtimeTestPayload]{Name: "test.codec.required"}, nil); !errors.Is(err, ErrCodecRequired) {
		t.Fatalf("expected ErrCodecRequired, got %v", err)
	}
}

func TestValidateListenerDefinitionErrors(t *testing.T) {
	runtime := newTestGala(t, nil)
	topic := Topic[runtimeTestPayload]{Name: TopicName("test.listener.validation")}

	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(nil, Definition[runtimeTestPayload]{}); !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic:  Topic[runtimeTestPayload]{Name: ""},
		Name:   "test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); !errors.Is(err, ErrTopicNameRequired) {
		t.Fatalf("expected ErrTopicNameRequired, got %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); err != nil {
		t.Fatalf("expected empty listener name to default to the topic name, got %v", err)
	}

	listeners := runtime.registry.registeredListeners(topic.Name)
	if len(listeners) != 1 || !strings.HasPrefix(listeners[0].name, string(topic.Name)+"#") {
		t.Fatalf("expected defaulted listener name prefixed %q, got %+v", string(topic.Name)+"#", listeners)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "test.listener",
		Handle: nil,
	}); !errors.Is(err, ErrListenerHandlerRequired) {
		t.Fatalf("expected ErrListenerHandlerRequired, got %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
		Schedule: &ScheduleSpec[runtimeTestPayload]{
			Handle: func(context.Context, runtimeTestPayload) (int, error) { return 0, nil },
		},
	}); !errors.Is(err, ErrListenerHandlerConflict) {
		t.Fatalf("expected ErrListenerHandlerConflict, got %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic:    topic,
		Name:     "test.listener",
		Schedule: &ScheduleSpec[runtimeTestPayload]{},
	}); !errors.Is(err, ErrListenerHandlerRequired) {
		t.Fatalf("expected ErrListenerHandlerRequired for schedule without handle, got %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "test.listener",
		Schedule: &ScheduleSpec[runtimeTestPayload]{
			Handle: func(context.Context, runtimeTestPayload) (int, error) { return 0, nil },
		},
	}); !errors.Is(err, ErrListenerScheduleStateRequired) {
		t.Fatalf("expected ErrListenerScheduleStateRequired for schedule without state extractor, got %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "test.listener",
		Schedule: &ScheduleSpec[runtimeTestPayload]{
			Handle: func(context.Context, runtimeTestPayload) (int, error) { return 0, nil },
			State:  func(runtimeTestPayload) ScheduleState { return ScheduleState{} },
		},
	}); !errors.Is(err, ErrListenerScheduleWrapRequired) {
		t.Fatalf("expected ErrListenerScheduleWrapRequired for schedule without wrap builder, got %v", err)
	}
}

// TestAttachListenerSkipsScheduleInMemory verifies scheduled definitions skip on an
// in-memory runtime: no error, no listener attached
func TestAttachListenerSkipsScheduleInMemory(t *testing.T) {
	runtime, err := NewGala(context.Background(), Config{
		DispatchMode: DispatchModeInMemory,
	})
	if err != nil {
		t.Fatalf("unexpected in-memory gala initialization error: %v", err)
	}

	t.Cleanup(func() {
		_ = runtime.Close()
	})

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.inmemory.schedule")}

	ids, err := Register(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.inmemory.schedule.listener",
		Schedule: &ScheduleSpec[runtimeTestPayload]{
			Handle: func(context.Context, runtimeTestPayload) (int, error) { return 0, nil },
			State:  func(runtimeTestPayload) ScheduleState { return ScheduleState{} },
			Wrap:   func(p runtimeTestPayload, _ ScheduleState) runtimeTestPayload { return p },
		},
	})
	if err != nil {
		t.Fatalf("expected scheduled definition to skip without error, got %v", err)
	}

	if len(ids) != 0 {
		t.Fatalf("expected no listener ids for a skipped scheduled definition, got %v", ids)
	}

	if runtime.InterestedIn(topic.Name, "") {
		t.Fatal("expected no listener attached for a skipped scheduled definition")
	}
}

// TestDispatchEnvelopeWithoutCallerUsesZeroValue verifies a caller-less dispatch installs the zero-value caller
func TestDispatchEnvelopeWithoutCallerUsesZeroValue(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.caller.zerovalue")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var seen *auth.Caller

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.caller.zerovalue.listener",
		Handle: func(ctx HandlerContext, _ runtimeTestPayload) error {
			seen = ctx.Caller

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "no-caller"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
	})
	if err != nil {
		t.Fatalf("expected caller-less dispatch to succeed, got %v", err)
	}

	if seen == nil || seen.Capabilities != 0 || seen.OrganizationID != "" {
		t.Fatalf("expected zero-value caller, got %+v", seen)
	}
}

// TestDispatchEnvelopeCallerHookSuppliesCaller verifies the Caller hook receives a zero-value
// stand-in when no caller was restored and its result reaches the handler
func TestDispatchEnvelopeCallerHookSuppliesCaller(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.caller.hook")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var observed *auth.Caller

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.caller.hook.listener",
		Caller: func(restored *auth.Caller, _ runtimeTestPayload) *auth.Caller {
			if restored.SubjectID != "" {
				t.Fatalf("expected zero-value restored caller, got %q", restored.SubjectID)
			}

			return &auth.Caller{SubjectID: "hook_subject"}
		},
		Handle: func(handlerContext HandlerContext, _ runtimeTestPayload) error {
			observed = handlerContext.Caller

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "hook"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	if err := runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
	}); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if observed == nil || observed.SubjectID != "hook_subject" {
		t.Fatalf("expected hook-supplied caller on handler context, got %+v", observed)
	}
}

// TestDispatchEnvelopeGatedListenerSkipsSilently verifies a false Gate skips the handler without error
func TestDispatchEnvelopeGatedListenerSkipsSilently(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.gate.skip")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	handled := 0
	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.gate.skip.listener",
		Gate:  func(context.Context, runtimeTestPayload) bool { return false },
		Handle: func(HandlerContext, runtimeTestPayload) error {
			handled++

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "gated"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	if err := runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: testCallerSnapshot(t, runtime),
	}); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if handled != 0 {
		t.Fatalf("expected gated handler not to run, got %d calls", handled)
	}
}

func TestTopicAlreadyRegistered(t *testing.T) {
	registry := newRegistry()
	topic := Topic[runtimeTestPayload]{Name: TopicName("test.duplicate")}

	if err := registerTopic(registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	if err := registerTopic(registry, topic, JSONCodec[runtimeTestPayload]{}); !errors.Is(err, ErrTopicAlreadyRegistered) {
		t.Fatalf("expected ErrTopicAlreadyRegistered, got %v", err)
	}
}

func TestPayloadOperationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		payload  any
		expected string
	}{
		{
			name:     "nil payload",
			payload:  nil,
			expected: "",
		},
		{
			name:     "string payload",
			payload:  "test",
			expected: "",
		},
		{
			name:     "int payload",
			payload:  42,
			expected: "",
		},
		{
			name:     "nil pointer",
			payload:  (*runtimeOperationPayload)(nil),
			expected: "",
		},
		{
			name:     "struct without operation field",
			payload:  runtimeTestPayload{Message: "hello"},
			expected: "",
		},
		{
			name:     "struct with operation field",
			payload:  runtimeOperationPayload{Operation: "create", Message: "hello"},
			expected: "create",
		},
		{
			name:     "pointer to struct with operation field",
			payload:  &runtimeOperationPayload{Operation: "update", Message: "hello"},
			expected: "update",
		},
		{
			name:     "struct with whitespace operation",
			payload:  runtimeOperationPayload{Operation: "  create  ", Message: "hello"},
			expected: "create",
		},
		{
			name:     "struct with empty operation",
			payload:  runtimeOperationPayload{Operation: "", Message: "hello"},
			expected: "",
		},
		{
			name:     "map payload",
			payload:  map[string]string{"operation": "create"},
			expected: "",
		},
		{
			name:     "slice payload",
			payload:  []string{"create"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := payloadOperation(tt.payload)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestConfigValidateDefaults(t *testing.T) {
	config := Config{
		ConnectionURI: "postgres://localhost/test",
		QueueName:     "",
		WorkerCount:   0,
	}

	if err := config.validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if config.QueueName != DefaultQueueName {
		t.Fatalf("expected queue name to default to %q, got %q", DefaultQueueName, config.QueueName)
	}

	if config.WorkerCount != 1 {
		t.Fatalf("expected worker count to default to 1, got %d", config.WorkerCount)
	}
}

func TestConfigValidatePreservesExplicitValues(t *testing.T) {
	config := Config{
		ConnectionURI: "postgres://localhost/test",
		QueueName:     "custom_queue",
		WorkerCount:   10,
	}

	if err := config.validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if config.QueueName != "custom_queue" {
		t.Fatalf("expected queue name to be preserved as %q, got %q", "custom_queue", config.QueueName)
	}

	if config.WorkerCount != 10 {
		t.Fatalf("expected worker count to be preserved as 10, got %d", config.WorkerCount)
	}
}

func TestConfigValidateInMemoryAllowsMissingConnectionURI(t *testing.T) {
	config := Config{
		DispatchMode: DispatchModeInMemory,
	}

	if err := config.validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestConfigValidateRejectsUnknownDispatchMode(t *testing.T) {
	config := Config{
		DispatchMode: DispatchMode("unknown"),
	}

	if err := config.validate(); !errors.Is(err, ErrDispatchModeInvalid) {
		t.Fatalf("expected ErrDispatchModeInvalid, got %v", err)
	}
}

func TestNewGalaInMemoryModeDoesNotRequireRiverConnection(t *testing.T) {
	runtime, err := NewGala(context.Background(), Config{
		DispatchMode: DispatchModeInMemory,
	})
	if err != nil {
		t.Fatalf("unexpected in-memory gala initialization error: %v", err)
	}

	t.Cleanup(func() {
		_ = runtime.Close()
	})

	if runtime.registry == nil {
		t.Fatalf("expected in-memory gala registry to be initialized")
	}

	if runtime.inMemoryPool == nil {
		t.Fatalf("expected in-memory gala pool to be initialized")
	}

	if err := runtime.StartWorkers(context.Background()); err != nil {
		t.Fatalf("expected StartWorkers to be a no-op in in-memory mode, got %v", err)
	}

	if err := runtime.StopWorkers(context.Background()); err != nil {
		t.Fatalf("expected StopWorkers to be a no-op in in-memory mode, got %v", err)
	}
}

func TestInMemoryDispatchUsesPoolWorkerLimit(t *testing.T) {
	runtime, err := NewGala(context.Background(), Config{
		DispatchMode: DispatchModeInMemory,
		WorkerCount:  1,
	})
	if err != nil {
		t.Fatalf("unexpected in-memory gala initialization error: %v", err)
	}

	t.Cleanup(func() {
		_ = runtime.Close()
	})

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.inmemory.pool")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	releaseFirst := make(chan struct{})

	callCount := 0
	callMu := sync.Mutex{}
	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.inmemory.pool.listener",
		Handle: func(_ HandlerContext, _ runtimeTestPayload) error {
			callMu.Lock()
			callCount++
			current := callCount
			callMu.Unlock()

			if current == 1 {
				close(firstStarted)
				<-releaseFirst
				return nil
			}

			if current == 2 {
				close(secondStarted)
			}

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	emitContext := auth.WithCaller(context.Background(), testCaller())

	errs := make(chan error, 2)
	go func() {
		_, emitErr := runtime.EmitWithHeaders(emitContext, topic.Name, runtimeTestPayload{Message: "one"}, Headers{Kind: JobKindSystem})
		errs <- emitErr
	}()

	<-firstStarted

	go func() {
		_, emitErr := runtime.EmitWithHeaders(emitContext, topic.Name, runtimeTestPayload{Message: "two"}, Headers{Kind: JobKindSystem})
		errs <- emitErr
	}()

	select {
	case <-secondStarted:
		t.Fatalf("expected second listener execution to wait for in-memory pool worker availability")
	case <-time.After(100 * time.Millisecond): //nolint:mnd
	}

	close(releaseFirst)

	for i := 0; i < 2; i++ {
		if emitErr := <-errs; emitErr != nil {
			t.Fatalf("unexpected emit error: %v", emitErr)
		}
	}

	select {
	case <-secondStarted:
	case <-time.After(1 * time.Second):
		t.Fatalf("expected second listener execution after first completed")
	}
}

func TestInMemoryEmitReturnsBeforeListenerCompletes(t *testing.T) {
	runtime, err := NewGala(context.Background(), Config{
		DispatchMode: DispatchModeInMemory,
		WorkerCount:  1,
	})
	if err != nil {
		t.Fatalf("unexpected in-memory gala initialization error: %v", err)
	}

	t.Cleanup(func() {
		_ = runtime.Close()
	})

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.inmemory.async")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.inmemory.async.listener",
		Handle: func(_ HandlerContext, _ runtimeTestPayload) error {
			close(started)
			<-release
			close(done)
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	emitErrCh := make(chan error, 1)
	go func() {
		_, emitErr := runtime.EmitWithHeaders(auth.WithCaller(context.Background(), testCaller()), topic.Name, runtimeTestPayload{Message: "async"}, Headers{Kind: JobKindSystem})
		emitErrCh <- emitErr
	}()

	select {
	case <-started:
	case <-time.After(1 * time.Second):
		t.Fatalf("expected listener to start")
	}

	select {
	case emitErr := <-emitErrCh:
		if emitErr != nil {
			t.Fatalf("unexpected emit error: %v", emitErr)
		}
	case <-time.After(200 * time.Millisecond): //nolint:mnd
		t.Fatalf("expected emit to return before listener completion")
	}

	select {
	case <-done:
		t.Fatalf("expected listener to remain blocked until release")
	default:
	}

	close(release)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatalf("expected listener to complete after release")
	}
}

func TestGalaCloseWithoutJobClient(t *testing.T) {
	runtime := &Gala{}

	if err := runtime.Close(); err != nil {
		t.Fatalf("expected no error when closing gala without job client, got %v", err)
	}
}

func TestEmitReturnsEncodeError(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.emit.encode.error")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, "wrong type", Headers{Kind: JobKindSystem}); !errors.Is(err, ErrPayloadTypeMismatch) {
		t.Fatalf("expected ErrPayloadTypeMismatch, got %v", err)
	}
}

func TestEmitReturnsTopicNotFoundError(t *testing.T) {
	runtime := newTestGala(t, nil)

	if _, err := runtime.EmitWithHeaders(context.Background(), TopicName("missing.topic"), runtimeTestPayload{Message: "test"}, Headers{Kind: JobKindSystem}); !errors.Is(err, ErrTopicNotRegistered) {
		t.Fatalf("expected ErrTopicNotRegistered, got %v", err)
	}
}

func TestDispatchEnvelopeReturnsTopicNotFoundError(t *testing.T) {
	runtime := newTestGala(t, nil)

	err := runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("missing.topic"),
		Payload: []byte(`{"message":"test"}`),
	})
	if !errors.Is(err, ErrTopicNotRegistered) {
		t.Fatalf("expected ErrTopicNotRegistered, got %v", err)
	}
}

func TestDispatchEnvelopeSkipsListenersWithMismatchedOperation(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeOperationPayload]{Name: TopicName("test.operation.skip")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeOperationPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	createCalls := 0
	if _, err := attachListener(runtime, Definition[runtimeOperationPayload]{
		Topic:      topic,
		Name:       "create.only.listener",
		Operations: []string{"create"},
		Handle: func(HandlerContext, runtimeOperationPayload) error {
			createCalls++
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeOperationPayload{
		Operation: "delete",
		Message:   "test",
	})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	if err := runtime.dispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
	}); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if createCalls != 0 {
		t.Fatalf("expected create listener not to be called for delete operation, got %d calls", createCalls)
	}
}

func TestRegisteredListenersReturnsEmptyForUnknownTopic(t *testing.T) {
	registry := newRegistry()

	listeners := registry.registeredListeners(TopicName("unknown.topic"))
	if listeners != nil {
		t.Fatalf("expected nil for unknown topic, got %v", listeners)
	}
}

func TestRegisteredListenersReturnsCopy(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.listeners.copy")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	first := runtime.registry.registeredListeners(topic.Name)
	second := runtime.registry.registeredListeners(topic.Name)

	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("expected 1 listener in each copy")
	}

	if &first[0] == &second[0] {
		t.Fatalf("expected different slice backing arrays")
	}
}

func TestNormalizeOperationsEdgeCases(t *testing.T) {
	result := normalizeOperations(nil)
	if result != nil {
		t.Fatalf("expected nil for nil input, got %v", result)
	}

	result = normalizeOperations([]string{})
	if result != nil {
		t.Fatalf("expected nil for empty input, got %v", result)
	}

	result = normalizeOperations([]string{"", "  ", "\t"})
	if result != nil {
		t.Fatalf("expected nil for whitespace-only input, got %v", result)
	}

	result = normalizeOperations([]string{"create", "  update  ", "delete"})
	if len(result) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(result))
	}

	if _, ok := result["update"]; !ok {
		t.Fatalf("expected trimmed 'update' operation")
	}
}

func TestListenerMatchesEdgeCases(t *testing.T) {
	wildcardListener := registeredListener{
		name: "wildcard",
		ops:  nil,
	}

	if !listenerMatches(wildcardListener, "create") {
		t.Fatalf("expected wildcard listener to match any operation")
	}

	if !listenerMatches(wildcardListener, "") {
		t.Fatalf("expected wildcard listener to match empty operation")
	}

	filteredListener := registeredListener{
		name: "filtered",
		ops:  map[string]struct{}{"create": {}},
	}

	if listenerMatches(filteredListener, "") {
		t.Fatalf("expected filtered listener not to match empty operation")
	}

	if listenerMatches(filteredListener, "update") {
		t.Fatalf("expected filtered listener not to match non-matching operation")
	}

	if !listenerMatches(filteredListener, "create") {
		t.Fatalf("expected filtered listener to match 'create' operation")
	}
}

func TestContextManagerRegisterErrors(t *testing.T) {
	manager, err := newContextManager()
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	if err := manager.Register(nil); !errors.Is(err, ErrContextCodecRequired) {
		t.Fatalf("expected ErrContextCodecRequired, got %v", err)
	}

	emptyKeyCodec := NewKeyCodec("", runtimeTestActorKey)
	if err := manager.Register(emptyKeyCodec); !errors.Is(err, ErrContextCodecKeyRequired) {
		t.Fatalf("expected ErrContextCodecKeyRequired, got %v", err)
	}

	validCodec := NewKeyCodec("test_actor", runtimeTestActorKey)
	if err := manager.Register(validCodec); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	if err := manager.Register(validCodec); !errors.Is(err, ErrContextCodecAlreadyRegistered) {
		t.Fatalf("expected ErrContextCodecAlreadyRegistered, got %v", err)
	}
}

func TestNewContextManagerWithInitialCodecs(t *testing.T) {
	codec1 := NewKeyCodec("actor_1", runtimeTestActorKey)
	codec2 := NewKeyCodec("payload_1", runtimeTestPayloadKey)

	manager, err := newContextManager(codec1, codec2)
	if err != nil {
		t.Fatalf("failed to create context manager with initial codecs: %v", err)
	}

	if err := manager.Register(codec1); !errors.Is(err, ErrContextCodecAlreadyRegistered) {
		t.Fatalf("expected codec1 to be already registered")
	}
}

func TestNewContextManagerInitialCodecError(t *testing.T) {
	nilCodec := NewKeyCodec("", runtimeTestActorKey)

	_, err := newContextManager(nilCodec)
	if !errors.Is(err, ErrContextCodecKeyRequired) {
		t.Fatalf("expected ErrContextCodecKeyRequired, got %v", err)
	}
}

func TestContextManagerCaptureAndRestore(t *testing.T) {
	codec := NewKeyCodec("test_actor", runtimeTestActorKey)
	manager, err := newContextManager(codec)
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	ctx := runtimeTestActorKey.Set(context.Background(), runtimeTestActor{ID: "actor-456"})
	ctx = WithFlag(ctx, ContextFlagWorkflowBypass)

	snapshot, err := manager.Capture(ctx)
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}

	if len(snapshot.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(snapshot.Values))
	}

	if len(snapshot.Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(snapshot.Flags))
	}

	restored, err := manager.Restore(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	actor, ok := runtimeTestActorKey.Get(restored)
	if !ok {
		t.Fatalf("expected actor in restored context")
	}

	if actor.ID != "actor-456" {
		t.Fatalf("expected actor ID 'actor-456', got %q", actor.ID)
	}

	if !HasFlag(restored, ContextFlagWorkflowBypass) {
		t.Fatalf("expected workflow bypass flag in restored context")
	}
}

func TestContextManagerRestoreSkipsUnknownKeys(t *testing.T) {
	manager, err := newContextManager()
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	snapshot := ContextSnapshot{
		Values: map[ContextKey]json.RawMessage{
			"unknown_key": []byte(`{"id": "test"}`),
		},
	}

	restored, err := manager.Restore(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	if restored == nil {
		t.Fatalf("expected non-nil context")
	}
}

func TestContextManagerRestoreFalseFlags(t *testing.T) {
	manager, err := newContextManager()
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	snapshot := ContextSnapshot{
		Flags: map[ContextFlag]bool{
			ContextFlagWorkflowBypass: false,
		},
	}

	restored, err := manager.Restore(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	if HasFlag(restored, ContextFlagWorkflowBypass) {
		t.Fatalf("expected workflow bypass flag to be false")
	}
}

func TestContextManagerCaptureEmptyContext(t *testing.T) {
	manager, err := newContextManager()
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	snapshot, err := manager.Capture(context.Background())
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}

	if snapshot.Values != nil {
		t.Fatalf("expected nil values for empty context")
	}

	if snapshot.Flags != nil {
		t.Fatalf("expected nil flags for empty context")
	}
}

func TestRegisterErrors(t *testing.T) {
	if _, err := Register(nil, Definition[runtimeTestPayload]{}); !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}

	runtime := newTestGala(t, nil)
	if _, err := Register(runtime, Definition[runtimeTestPayload]{
		Topic:  Topic[runtimeTestPayload]{Name: "test.topic"},
		Name:   "",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); err != nil {
		t.Fatalf("expected empty listener name to default to the topic name, got %v", err)
	}

	listeners := runtime.registry.registeredListeners(TopicName("test.topic"))
	if len(listeners) != 1 || !strings.HasPrefix(listeners[0].name, "test.topic#") {
		t.Fatalf("expected defaulted listener name prefixed %q, got %+v", "test.topic#", listeners)
	}
}

func TestRegisterMultipleDefinitions(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.multi.listeners")}

	ids, err := Register(runtime,
		Definition[runtimeTestPayload]{
			Topic:  topic,
			Name:   "listener.one",
			Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
		},
		Definition[runtimeTestPayload]{
			Topic:  topic,
			Name:   "listener.two",
			Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
		},
	)
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 listener IDs, got %d", len(ids))
	}

	listeners := runtime.registry.registeredListeners(topic.Name)
	if len(listeners) != 2 {
		t.Fatalf("expected 2 registered listeners, got %d", len(listeners))
	}
}

func TestRegistryTopicRegistrationUnknownTopic(t *testing.T) {
	registry := newRegistry()

	_, err := registry.topicRegistration(TopicName("unknown.topic"))
	if !errors.Is(err, ErrTopicNotRegistered) {
		t.Fatalf("expected ErrTopicNotRegistered, got %v", err)
	}
}

func TestRegistryTopicRegistrationEmptyTopic(t *testing.T) {
	registry := newRegistry()

	_, err := registry.topicRegistration(TopicName(""))
	if !errors.Is(err, ErrTopicNameRequired) {
		t.Fatalf("expected ErrTopicNameRequired, got %v", err)
	}
}

func TestRegistryDecodePayloadInvalidJSON(t *testing.T) {
	registry := newRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.decode.invalid")}
	if err := registerTopic(registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	registration, err := registry.topicRegistration(topic.Name)
	if err != nil {
		t.Fatalf("failed to resolve topic registration: %v", err)
	}

	_, err = registration.decode([]byte(`{invalid`))
	if !errors.Is(err, ErrPayloadDecodeFailed) {
		t.Fatalf("expected ErrPayloadDecodeFailed, got %v", err)
	}
}

func TestRegistryEncodePayloadTypeMismatch(t *testing.T) {
	registry := newRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.encode.mismatch")}
	if err := registerTopic(registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	registration, err := registry.topicRegistration(topic.Name)
	if err != nil {
		t.Fatalf("failed to resolve topic registration: %v", err)
	}

	_, err = registration.encode("wrong type")
	if !errors.Is(err, ErrPayloadTypeMismatch) {
		t.Fatalf("expected ErrPayloadTypeMismatch, got %v", err)
	}
}

func TestLogFieldsCodecKey(t *testing.T) {
	codec := logFieldsCodec{}

	if key := codec.Key(); key != "log_fields" {
		t.Fatalf("expected key 'log_fields', got %q", key)
	}
}

func TestLogFieldsCodecCaptureEmpty(t *testing.T) {
	codec := logFieldsCodec{}

	raw, present, err := codec.Capture(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if present {
		t.Fatalf("expected not present for empty context")
	}

	if raw != nil {
		t.Fatalf("expected nil raw message")
	}
}

func TestLogFieldsCodecCaptureAndRestore(t *testing.T) {
	codec := logFieldsCodec{}

	ctx := logx.WithFields(context.Background(), map[string]any{
		"integration_id": "int_abc",
		"definition_id":  "def_xyz",
		"operation":      "sync_users",
	})

	raw, present, err := codec.Capture(ctx)
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}

	if !present {
		t.Fatalf("expected value present")
	}

	restored, err := codec.Restore(context.Background(), raw)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	fields := logx.FieldsFromContext(restored)
	if fields["integration_id"] != "int_abc" {
		t.Fatalf("expected integration_id 'int_abc', got %v", fields["integration_id"])
	}

	if fields["definition_id"] != "def_xyz" {
		t.Fatalf("expected definition_id 'def_xyz', got %v", fields["definition_id"])
	}

	if fields["operation"] != "sync_users" {
		t.Fatalf("expected operation 'sync_users', got %v", fields["operation"])
	}
}

func TestLogFieldsCodecRestoreInvalidJSON(t *testing.T) {
	codec := logFieldsCodec{}

	_, err := codec.Restore(context.Background(), []byte("{invalid"))
	if !errors.Is(err, ErrContextSnapshotRestoreFailed) {
		t.Fatalf("expected ErrContextSnapshotRestoreFailed, got %v", err)
	}
}

func TestLogFieldsCodecRestoreEmptyMap(t *testing.T) {
	codec := logFieldsCodec{}

	restored, err := codec.Restore(context.Background(), []byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := logx.FieldsFromContext(restored)
	if len(fields) != 0 {
		t.Fatalf("expected no fields, got %d", len(fields))
	}
}

func TestLogFieldsCodecRoundTripViaContextManager(t *testing.T) {
	manager, err := newContextManager(
		NewKeyCodec("caller", auth.CallerKey),
		logFieldsCodec{},
	)
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:      "subject_rt",
		OrganizationID: "org_rt",
	})
	ctx = logx.WithFields(ctx, map[string]any{
		"integration_id": "int_rt",
		"run_id":         "run_rt",
	})

	snapshot, err := manager.Capture(ctx)
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}

	if len(snapshot.Values) != 2 {
		t.Fatalf("expected 2 snapshot values (caller + log_fields), got %d", len(snapshot.Values))
	}

	restored, err := manager.Restore(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	caller, ok := auth.CallerFromContext(restored)
	if !ok || caller == nil {
		t.Fatalf("expected caller in restored context")
	}

	if caller.SubjectID != "subject_rt" {
		t.Fatalf("expected subject ID 'subject_rt', got %q", caller.SubjectID)
	}

	fields := logx.FieldsFromContext(restored)
	if fields["integration_id"] != "int_rt" {
		t.Fatalf("expected integration_id 'int_rt', got %v", fields["integration_id"])
	}

	if fields["run_id"] != "run_rt" {
		t.Fatalf("expected run_id 'run_rt', got %v", fields["run_id"])
	}
}

func TestWithFlagAndHasFlag(t *testing.T) {
	ctx := context.Background()

	if HasFlag(ctx, ContextFlagWorkflowBypass) {
		t.Fatalf("expected flag not set on empty context")
	}

	ctx = WithFlag(ctx, ContextFlagWorkflowBypass)

	if !HasFlag(ctx, ContextFlagWorkflowBypass) {
		t.Fatalf("expected flag to be set")
	}

	if HasFlag(ctx, ContextFlagWorkflowAllowEventEmission) {
		t.Fatalf("expected other flag not set")
	}

	ctx = WithFlag(ctx, ContextFlagWorkflowAllowEventEmission)

	if !HasFlag(ctx, ContextFlagWorkflowBypass) {
		t.Fatalf("expected first flag still set")
	}

	if !HasFlag(ctx, ContextFlagWorkflowAllowEventEmission) {
		t.Fatalf("expected second flag to be set")
	}
}

func TestNewEventIDGeneratesUniqueIDs(t *testing.T) {
	ids := make(map[EventID]bool)

	for range 100 {
		id := NewEventID()
		if ids[id] {
			t.Fatalf("duplicate event ID generated: %s", id)
		}

		ids[id] = true
	}
}
