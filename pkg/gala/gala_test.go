package gala

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/samber/do/v2"
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
func newTestGala(t *testing.T, dispatcher Dispatcher) *Gala {
	t.Helper()

	return newTestGalaInMemory(t, dispatcher)
}

// TestRuntimeDispatchEnvelopeWithDependencyInjectionAndContextRehydration verifies
// that handlers can resolve dependencies from samber/do and receive rehydrated context.
func TestRuntimeDispatchEnvelopeWithDependencyInjectionAndContextRehydration(t *testing.T) {
	injector := do.New()
	do.ProvideValue(injector, &runtimeTestFormatter{Prefix: "fmt"})

	contextManager, err := NewContextManager(NewTypedContextCodec[runtimeTestActor]("runtime_test_actor"))
	if err != nil {
		t.Fatalf("failed to build context manager: %v", err)
	}

	runtime := newTestGala(t, nil)
	runtime.injector = injector
	runtime.contextManager = contextManager

	topic := Topic[runtimeTestPayload]{
		Name: TopicName("runtime.test.event"),
	}

	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var observed string

	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.listener",
		Handle: func(handlerContext HandlerContext, payload runtimeTestPayload) error {
			formatter := do.MustInvoke[*runtimeTestFormatter](handlerContext.Injector)

			actor, exists := contextx.From[runtimeTestActor](handlerContext.Context)
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

	emitContext := contextx.With(context.Background(), runtimeTestActor{ID: "actor-1"})
	emitContext = WithFlag(emitContext, ContextFlagWorkflowBypass)

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "hello"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	snapshot, err := runtime.ContextManager().Capture(emitContext)
	if err != nil {
		t.Fatalf("failed to capture context snapshot: %v", err)
	}

	if err := runtime.DispatchEnvelope(context.Background(), Envelope{
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

// TestRuntimeDispatchEnvelopeWithAuthenticatedUserContext verifies default runtime codecs
// rehydrate auth context values for listener execution.
func TestRuntimeDispatchEnvelopeWithAuthenticatedUserContext(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{
		Name: TopicName("runtime.test.auth"),
	}

	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var observed auth.AuthenticatedUser
	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.auth.listener",
		Handle: func(handlerContext HandlerContext, _ runtimeTestPayload) error {
			au, err := auth.GetAuthenticatedUserFromContext(handlerContext.Context)
			if err != nil {
				return err
			}

			observed = *au
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	emitContext := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:          "subject_123",
		SubjectName:        "Codex User",
		SubjectEmail:       "codex@example.com",
		OrganizationID:     "org_123",
		OrganizationName:   "Acme Corp",
		OrganizationIDs:    []string{"org_123", "org_234"},
		AuthenticationType: auth.JWTAuthentication,
		OrganizationRole:   auth.OwnerRole,
		IsSystemAdmin:      true,
	})

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "auth"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	snapshot, err := runtime.ContextManager().Capture(emitContext)
	if err != nil {
		t.Fatalf("failed to capture context snapshot: %v", err)
	}

	if err := runtime.DispatchEnvelope(context.Background(), Envelope{
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

	if !observed.IsSystemAdmin {
		t.Fatalf("expected system admin flag to be true")
	}
}

// TestAttachListenerRequiresTopicRegistration verifies listener registration
// fails when the topic contract has not been registered.
func TestAttachListenerRequiresTopicRegistration(t *testing.T) {
	registry := NewRegistry()

	_, err := AttachListener(registry, Definition[runtimeTestPayload]{
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	err := runtime.DispatchEnvelope(context.Background(), Envelope{
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	listenerCalls := 0
	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.durable.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			listenerCalls++

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "durable"}, Headers{})
	if receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "durable"}, Headers{})
	if !errors.Is(receipt.Err, ErrDispatcherRequired) {
		t.Fatalf("expected ErrDispatcherRequired, got %v", receipt.Err)
	}

	if receipt.Accepted {
		t.Fatalf("expected non-accepted receipt")
	}
}

// TestRuntimeEmitEnvelopeUsesPrebuiltEventID verifies pre-built envelopes preserve the caller event ID.
func TestRuntimeEmitEnvelopeUsesPrebuiltEventID(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.envelope")}
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "prebuilt"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	prebuiltID := EventID("evt_prebuilt_123")
	err = runtime.EmitEnvelope(context.Background(), Envelope{
		ID:      prebuiltID,
		Topic:   topic.Name,
		Payload: encodedPayload,
	})
	if err != nil {
		t.Fatalf("unexpected emit envelope error: %v", err)
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

type runtimeAccessorTestDispatcher struct{}

func (runtimeAccessorTestDispatcher) Dispatch(context.Context, Envelope) error {
	return nil
}

func TestRuntimeAccessorsReturnConfiguredDependencies(t *testing.T) {
	injector := do.New()
	contextManager, err := NewContextManager()
	if err != nil {
		t.Fatalf("failed to build context manager: %v", err)
	}

	dispatcher := runtimeAccessorTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	runtime.injector = injector
	runtime.contextManager = contextManager

	if runtime.Injector() != injector {
		t.Fatalf("unexpected injector accessor result")
	}

	if runtime.ContextManager() != contextManager {
		t.Fatalf("unexpected context manager accessor result")
	}
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

func TestRuntimeAccessorsRequireRuntimeInstance(t *testing.T) {
	var runtime *Gala

	assertPanics(t, func() { _ = runtime.Registry() })
	assertPanics(t, func() { _ = runtime.Injector() })
	assertPanics(t, func() { _ = runtime.ContextManager() })
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()

	fn()
}

func TestRuntimeDispatchEnvelopeWrapsListenerErrors(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.listener.error")}
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	listenerErr := errors.New("listener failed")
	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "failing.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			return listenerErr
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "test"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.DispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
	})
	if err == nil {
		t.Fatalf("expected error from failing listener")
	}

	var listenerError ListenerError
	if !errors.As(err, &listenerError) {
		t.Fatalf("expected ListenerError, got %T", err)
	}

	if listenerError.ListenerName != "failing.listener" {
		t.Fatalf("expected listener name 'failing.listener', got %q", listenerError.ListenerName)
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "panicking.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			panic("listener panic")
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "test"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.DispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
	})
	if err == nil {
		t.Fatalf("expected error from panicking listener")
	}

	var listenerError ListenerError
	if !errors.As(err, &listenerError) {
		t.Fatalf("expected ListenerError, got %T", err)
	}

	if listenerError.ListenerName != "panicking.listener" {
		t.Fatalf("expected listener name 'panicking.listener', got %q", listenerError.ListenerName)
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeOperationPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeOperationPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	createCalls := 0
	updateCalls := 0

	if _, err := AttachListener(runtime.Registry(), Definition[runtimeOperationPayload]{
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

	if _, err := AttachListener(runtime.Registry(), Definition[runtimeOperationPayload]{
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

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeOperationPayload{
		Operation: ent.OpUpdateOne.String(),
		Message:   "test",
	})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.DispatchEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
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

type failingDispatcher struct {
	err error
}

func (d failingDispatcher) Dispatch(context.Context, Envelope) error {
	return d.err
}

func TestRuntimeEmitReturnsDurableDispatchError(t *testing.T) {
	dispatcher := failingDispatcher{err: errors.New("durable failed")}
	runtime := newTestGala(t, dispatcher)

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.durable.error")}
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "durable.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "test"}, Headers{})
	if receipt.Err == nil {
		t.Fatalf("expected error from durable dispatch")
	}

	if !errors.Is(receipt.Err, ErrDispatchFailed) {
		t.Fatalf("expected ErrDispatchFailed, got %v", receipt.Err)
	}
}

func TestRegistryConcurrentRegistration(t *testing.T) {
	registry := NewRegistry()

	const numGoroutines = 100

	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			topic := Topic[runtimeTestPayload]{Name: TopicName(fmt.Sprintf("topic.%d", n))}
			_ = RegisterTopic(registry, Registration[runtimeTestPayload]{
				Topic: topic,
				Codec: JSONCodec[runtimeTestPayload]{},
			})
			_, _ = AttachListener(registry, Definition[runtimeTestPayload]{
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
			expected: "gala: listener panicked",
		},
		{
			name:     "non-panicked listener",
			err:      ListenerError{ListenerName: "test.listener", Panicked: false, Cause: errors.New("failed")},
			expected: "gala: listener execution failed",
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
	registry := NewRegistry()

	if registry.InterestedIn("", "create") {
		t.Fatalf("expected false for empty topic")
	}
}

func TestRegistryInterestedInNoListeners(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.no.listeners")}
	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if registry.InterestedIn(topic.Name, "create") {
		t.Fatalf("expected false when no listeners registered")
	}
}

func TestRegistryInterestedInEmptyOperation(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.empty.operation")}
	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(registry, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	if !registry.InterestedIn(topic.Name, "") {
		t.Fatalf("expected true for empty operation when listeners exist")
	}

	if !registry.InterestedIn(topic.Name, "   ") {
		t.Fatalf("expected true for whitespace-only operation when listeners exist")
	}
}

func TestRegistryInterestedInWithOperationFilter(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeOperationPayload]{Name: TopicName("test.operation.filter")}
	if err := RegisterTopic(registry, Registration[runtimeOperationPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeOperationPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(registry, Definition[runtimeOperationPayload]{
		Topic:      topic,
		Name:       "test.create.listener",
		Operations: []string{"create"},
		Handle:     func(HandlerContext, runtimeOperationPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	if !registry.InterestedIn(topic.Name, "create") {
		t.Fatalf("expected true for matching operation")
	}

	if registry.InterestedIn(topic.Name, "update") {
		t.Fatalf("expected false for non-matching operation")
	}
}

func TestRegistryInterestedInWithWildcardListener(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeOperationPayload]{Name: TopicName("test.wildcard.listener")}
	if err := RegisterTopic(registry, Registration[runtimeOperationPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeOperationPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(registry, Definition[runtimeOperationPayload]{
		Topic:  topic,
		Name:   "test.wildcard",
		Handle: func(HandlerContext, runtimeOperationPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	if !registry.InterestedIn(topic.Name, "create") {
		t.Fatalf("expected true for wildcard listener with any operation")
	}

	if !registry.InterestedIn(topic.Name, "update") {
		t.Fatalf("expected true for wildcard listener with any operation")
	}
}

func TestValidateTopicRegistrationErrors(t *testing.T) {
	if err := RegisterTopic(nil, Registration[runtimeTestPayload]{}); !errors.Is(err, ErrRegistryRequired) {
		t.Fatalf("expected ErrRegistryRequired, got %v", err)
	}

	registry := NewRegistry()

	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: Topic[runtimeTestPayload]{Name: ""},
		Codec: JSONCodec[runtimeTestPayload]{},
	}); !errors.Is(err, ErrTopicNameRequired) {
		t.Fatalf("expected ErrTopicNameRequired, got %v", err)
	}

	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: Topic[runtimeTestPayload]{Name: "test.codec.required"},
		Codec: nil,
	}); !errors.Is(err, ErrCodecRequired) {
		t.Fatalf("expected ErrCodecRequired, got %v", err)
	}
}

func TestValidateListenerDefinitionErrors(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[runtimeTestPayload]{Name: TopicName("test.listener.validation")}

	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(nil, Definition[runtimeTestPayload]{}); !errors.Is(err, ErrRegistryRequired) {
		t.Fatalf("expected ErrRegistryRequired, got %v", err)
	}

	if _, err := AttachListener(registry, Definition[runtimeTestPayload]{
		Topic:  Topic[runtimeTestPayload]{Name: ""},
		Name:   "test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); !errors.Is(err, ErrTopicNameRequired) {
		t.Fatalf("expected ErrTopicNameRequired, got %v", err)
	}

	if _, err := AttachListener(registry, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); !errors.Is(err, ErrListenerNameRequired) {
		t.Fatalf("expected ErrListenerNameRequired, got %v", err)
	}

	if _, err := AttachListener(registry, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "test.listener",
		Handle: nil,
	}); !errors.Is(err, ErrListenerHandlerRequired) {
		t.Fatalf("expected ErrListenerHandlerRequired, got %v", err)
	}
}

func TestTopicAlreadyRegistered(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[runtimeTestPayload]{Name: TopicName("test.duplicate")}

	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); !errors.Is(err, ErrTopicAlreadyRegistered) {
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

	if runtime.Registry() == nil {
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	releaseFirst := make(chan struct{})

	callCount := 0
	callMu := sync.Mutex{}
	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
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

	errs := make(chan error, 2)
	go func() {
		receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "one"}, Headers{})
		errs <- receipt.Err
	}()

	<-firstStarted

	go func() {
		receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "two"}, Headers{})
		errs <- receipt.Err
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
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

	receiptCh := make(chan EmitReceipt, 1)
	go func() {
		receiptCh <- runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "async"}, Headers{})
	}()

	select {
	case <-started:
	case <-time.After(1 * time.Second):
		t.Fatalf("expected listener to start")
	}

	select {
	case receipt := <-receiptCh:
		if receipt.Err != nil {
			t.Fatalf("unexpected emit error: %v", receipt.Err)
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

func TestBuildQueueConfigIncludesAdditionalQueues(t *testing.T) {
	queues := buildQueueConfig("events", 10, map[string]int{
		"integrations": 4,
		"":             3,
		"bad":          0,
	})

	defaultQueue, ok := queues["events"]
	if !ok {
		t.Fatalf("expected default events queue in config")
	}
	if defaultQueue.MaxWorkers != 10 {
		t.Fatalf("expected events max workers 10, got %d", defaultQueue.MaxWorkers)
	}

	integrationQueue, ok := queues["integrations"]
	if !ok {
		t.Fatalf("expected integrations queue in config")
	}
	if integrationQueue.MaxWorkers != 4 {
		t.Fatalf("expected integrations max workers 4, got %d", integrationQueue.MaxWorkers)
	}

	if _, exists := queues[""]; exists {
		t.Fatalf("did not expect empty queue name in config")
	}
	if _, exists := queues["bad"]; exists {
		t.Fatalf("did not expect non-positive worker queue in config")
	}
}

func TestGalaCloseWithoutJobClient(t *testing.T) {
	runtime := &Gala{}

	if err := runtime.Close(); err != nil {
		t.Fatalf("expected no error when closing gala without job client, got %v", err)
	}
}

func TestEmitEnvelopeRequiresDispatcher(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.emit.envelope.dispatcher")}
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "test"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	err = runtime.EmitEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
	})
	if !errors.Is(err, ErrDispatcherRequired) {
		t.Fatalf("expected ErrDispatcherRequired, got %v", err)
	}
}

func TestEmitEnvelopeRequiresRegisteredTopic(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)

	err := runtime.EmitEnvelope(context.Background(), Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("unregistered.topic"),
		Payload: []byte(`{"message":"test"}`),
	})
	if !errors.Is(err, ErrTopicNotRegistered) {
		t.Fatalf("expected ErrTopicNotRegistered, got %v", err)
	}
}

func TestEmitWithHeadersReturnsEncodeError(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.emit.encode.error")}
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, "wrong type", Headers{})
	if !errors.Is(receipt.Err, ErrPayloadTypeMismatch) {
		t.Fatalf("expected ErrPayloadTypeMismatch, got %v", receipt.Err)
	}
}

func TestEmitWithHeadersReturnsTopicNotFoundError(t *testing.T) {
	runtime := newTestGala(t, nil)

	receipt := runtime.EmitWithHeaders(context.Background(), TopicName("missing.topic"), runtimeTestPayload{Message: "test"}, Headers{})
	if !errors.Is(receipt.Err, ErrTopicNotRegistered) {
		t.Fatalf("expected ErrTopicNotRegistered, got %v", receipt.Err)
	}
}

func TestDispatchEnvelopeReturnsTopicNotFoundError(t *testing.T) {
	runtime := newTestGala(t, nil)

	err := runtime.DispatchEnvelope(context.Background(), Envelope{
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeOperationPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeOperationPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	createCalls := 0
	if _, err := AttachListener(runtime.Registry(), Definition[runtimeOperationPayload]{
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

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeOperationPayload{
		Operation: "delete",
		Message:   "test",
	})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	if err := runtime.DispatchEnvelope(context.Background(), Envelope{
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
	registry := NewRegistry()

	listeners := registry.registeredListeners(TopicName("unknown.topic"))
	if listeners != nil {
		t.Fatalf("expected nil for unknown topic, got %v", listeners)
	}
}

func TestRegisteredListenersReturnsCopy(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.listeners.copy")}
	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(registry, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "test.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	first := registry.registeredListeners(topic.Name)
	second := registry.registeredListeners(topic.Name)

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

func TestListenerInterestedInOperationEdgeCases(t *testing.T) {
	wildcardListener := registeredListener{
		name: "wildcard",
		ops:  nil,
	}

	if !listenerInterestedInOperation(wildcardListener, "create") {
		t.Fatalf("expected wildcard listener to match any operation")
	}

	if !listenerInterestedInOperation(wildcardListener, "") {
		t.Fatalf("expected wildcard listener to match empty operation")
	}

	filteredListener := registeredListener{
		name: "filtered",
		ops:  map[string]struct{}{"create": {}},
	}

	if listenerInterestedInOperation(filteredListener, "") {
		t.Fatalf("expected filtered listener not to match empty operation")
	}

	if listenerInterestedInOperation(filteredListener, "update") {
		t.Fatalf("expected filtered listener not to match non-matching operation")
	}

	if !listenerInterestedInOperation(filteredListener, "create") {
		t.Fatalf("expected filtered listener to match 'create' operation")
	}
}

func TestContextManagerRegisterErrors(t *testing.T) {
	manager, err := NewContextManager()
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	if err := manager.Register(nil); !errors.Is(err, ErrContextCodecRequired) {
		t.Fatalf("expected ErrContextCodecRequired, got %v", err)
	}

	emptyKeyCodec := NewTypedContextCodec[runtimeTestActor]("")
	if err := manager.Register(emptyKeyCodec); !errors.Is(err, ErrContextCodecKeyRequired) {
		t.Fatalf("expected ErrContextCodecKeyRequired, got %v", err)
	}

	validCodec := NewTypedContextCodec[runtimeTestActor]("test_actor")
	if err := manager.Register(validCodec); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	if err := manager.Register(validCodec); !errors.Is(err, ErrContextCodecAlreadyRegistered) {
		t.Fatalf("expected ErrContextCodecAlreadyRegistered, got %v", err)
	}
}

func TestNewContextManagerWithInitialCodecs(t *testing.T) {
	codec1 := NewTypedContextCodec[runtimeTestActor]("actor_1")
	codec2 := NewTypedContextCodec[runtimeTestPayload]("payload_1")

	manager, err := NewContextManager(codec1, codec2)
	if err != nil {
		t.Fatalf("failed to create context manager with initial codecs: %v", err)
	}

	if err := manager.Register(codec1); !errors.Is(err, ErrContextCodecAlreadyRegistered) {
		t.Fatalf("expected codec1 to be already registered")
	}
}

func TestNewContextManagerInitialCodecError(t *testing.T) {
	nilCodec := NewTypedContextCodec[runtimeTestActor]("")

	_, err := NewContextManager(nilCodec)
	if !errors.Is(err, ErrContextCodecKeyRequired) {
		t.Fatalf("expected ErrContextCodecKeyRequired, got %v", err)
	}
}

func TestTypedContextCodecCaptureNotPresent(t *testing.T) {
	codec := NewTypedContextCodec[runtimeTestActor]("test_actor")

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

func TestTypedContextCodecCaptureAndRestore(t *testing.T) {
	codec := NewTypedContextCodec[runtimeTestActor]("test_actor")

	ctx := contextx.With(context.Background(), runtimeTestActor{ID: "actor-123"})

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

	actor, ok := contextx.From[runtimeTestActor](restored)
	if !ok {
		t.Fatalf("expected actor in restored context")
	}

	if actor.ID != "actor-123" {
		t.Fatalf("expected actor ID 'actor-123', got %q", actor.ID)
	}
}

func TestTypedContextCodecRestoreInvalidJSON(t *testing.T) {
	codec := NewTypedContextCodec[runtimeTestActor]("test_actor")

	_, err := codec.Restore(context.Background(), []byte("{invalid"))
	if !errors.Is(err, ErrContextSnapshotRestoreFailed) {
		t.Fatalf("expected ErrContextSnapshotRestoreFailed, got %v", err)
	}
}

func TestContextManagerCaptureAndRestore(t *testing.T) {
	codec := NewTypedContextCodec[runtimeTestActor]("test_actor")
	manager, err := NewContextManager(codec)
	if err != nil {
		t.Fatalf("failed to create context manager: %v", err)
	}

	ctx := contextx.With(context.Background(), runtimeTestActor{ID: "actor-456"})
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

	actor, ok := contextx.From[runtimeTestActor](restored)
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
	manager, err := NewContextManager()
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
	manager, err := NewContextManager()
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
	manager, err := NewContextManager()
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

func TestRegisterListenersErrors(t *testing.T) {
	if _, err := RegisterListeners(nil, Definition[runtimeTestPayload]{}); !errors.Is(err, ErrRegistryRequired) {
		t.Fatalf("expected ErrRegistryRequired, got %v", err)
	}

	registry := NewRegistry()
	if _, err := RegisterListeners(registry, Definition[runtimeTestPayload]{
		Topic:  Topic[runtimeTestPayload]{Name: "test.topic"},
		Name:   "",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	}); !errors.Is(err, ErrListenerNameRequired) {
		t.Fatalf("expected ErrListenerNameRequired, got %v", err)
	}
}

func TestRegisterListenersMultipleDefinitions(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.multi.listeners")}

	ids, err := RegisterListeners(registry,
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

	listeners := registry.registeredListeners(topic.Name)
	if len(listeners) != 2 {
		t.Fatalf("expected 2 registered listeners, got %d", len(listeners))
	}
}

func TestRegistryEncodePayloadUnknownTopic(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.EncodePayload(TopicName("unknown.topic"), runtimeTestPayload{})
	if !errors.Is(err, ErrTopicNotRegistered) {
		t.Fatalf("expected ErrTopicNotRegistered, got %v", err)
	}
}

func TestRegistryEncodePayloadEmptyTopic(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.EncodePayload(TopicName(""), runtimeTestPayload{})
	if !errors.Is(err, ErrTopicNameRequired) {
		t.Fatalf("expected ErrTopicNameRequired, got %v", err)
	}
}

func TestRegistryDecodePayloadUnknownTopic(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.DecodePayload(TopicName("unknown.topic"), []byte(`{}`))
	if !errors.Is(err, ErrTopicNotRegistered) {
		t.Fatalf("expected ErrTopicNotRegistered, got %v", err)
	}
}

func TestRegistryDecodePayloadEmptyTopic(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.DecodePayload(TopicName(""), []byte(`{}`))
	if !errors.Is(err, ErrTopicNameRequired) {
		t.Fatalf("expected ErrTopicNameRequired, got %v", err)
	}
}

func TestRegistryDecodePayloadInvalidJSON(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.decode.invalid")}
	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	_, err := registry.DecodePayload(topic.Name, []byte(`{invalid`))
	if !errors.Is(err, ErrPayloadDecodeFailed) {
		t.Fatalf("expected ErrPayloadDecodeFailed, got %v", err)
	}
}

func TestRegistryEncodePayloadTypeMismatch(t *testing.T) {
	registry := NewRegistry()

	topic := Topic[runtimeTestPayload]{Name: TopicName("test.encode.mismatch")}
	if err := RegisterTopic(registry, Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	_, err := registry.EncodePayload(topic.Name, "wrong type")
	if !errors.Is(err, ErrPayloadTypeMismatch) {
		t.Fatalf("expected ErrPayloadTypeMismatch, got %v", err)
	}
}

func TestContextCodecKey(t *testing.T) {
	codec := NewContextCodec()

	if key := codec.Key(); key != "durable" {
		t.Fatalf("expected key 'durable', got %q", key)
	}
}

func TestContextCodecCaptureWithoutAuthContext(t *testing.T) {
	codec := NewContextCodec()

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

func TestContextCodecCaptureAndRestore(t *testing.T) {
	codec := NewContextCodec()

	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:       "subject_test",
		OrganizationID:  "org_test",
		OrganizationIDs: []string{"org_1", "org_2"},
		IsSystemAdmin:   true,
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

	user, err := auth.GetAuthenticatedUserFromContext(restored)
	if err != nil {
		t.Fatalf("failed to get user from restored context: %v", err)
	}

	if user.SubjectID != "subject_test" {
		t.Fatalf("expected subject ID 'subject_test', got %q", user.SubjectID)
	}

	if len(user.OrganizationIDs) != 2 {
		t.Fatalf("expected 2 organization IDs, got %d", len(user.OrganizationIDs))
	}

	if !user.IsSystemAdmin {
		t.Fatalf("expected IsSystemAdmin to be true")
	}
}

func TestContextCodecRestoreInvalidJSON(t *testing.T) {
	codec := NewContextCodec()

	_, err := codec.Restore(context.Background(), []byte("{invalid"))
	if !errors.Is(err, ErrContextSnapshotRestoreFailed) {
		t.Fatalf("expected ErrContextSnapshotRestoreFailed, got %v", err)
	}
}

func TestAuthContextSnapshotToAuthenticatedUser(t *testing.T) {
	snapshot := AuthSnapshot{
		SubjectID:          "sub_123",
		SubjectName:        "Test User",
		SubjectEmail:       "test@example.com",
		OrganizationID:     "org_456",
		OrganizationName:   "Test Org",
		OrganizationIDs:    []string{"org_456", "org_789"},
		AuthenticationType: string(auth.JWTAuthentication),
		OrganizationRole:   string(auth.OwnerRole),
		IsSystemAdmin:      true,
	}

	user := snapshot.ToAuthenticatedUser()

	if user.SubjectID != "sub_123" {
		t.Fatalf("expected subject ID 'sub_123', got %q", user.SubjectID)
	}

	if user.AuthenticationType != auth.JWTAuthentication {
		t.Fatalf("expected JWT authentication type")
	}

	if user.OrganizationRole != auth.OwnerRole {
		t.Fatalf("expected owner role")
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
