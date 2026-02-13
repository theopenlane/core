package gala

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

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

// runtimeTestDurableDispatcher captures durable dispatch calls in tests.
type runtimeTestDurableDispatcher struct {
	calls     int
	envelopes []Envelope
	policies  []TopicPolicy
	err       error
}

// DispatchDurable records durable dispatch invocations.
func (d *runtimeTestDurableDispatcher) DispatchDurable(_ context.Context, envelope Envelope, policy TopicPolicy) error {
	d.calls++
	d.envelopes = append(d.envelopes, envelope)
	d.policies = append(d.policies, policy)

	return d.err
}

// TestRuntimeEmitDispatchesWithDependencyInjectionAndContextRehydration verifies
// that handlers can resolve dependencies from samber/do and receive rehydrated context.
func TestRuntimeEmitDispatchesWithDependencyInjectionAndContextRehydration(t *testing.T) {
	injector := NewInjector()
	ProvideValue(injector, &runtimeTestFormatter{Prefix: "fmt"})

	contextManager, err := NewContextManager(NewTypedContextCodec[runtimeTestActor](ContextKey("runtime_test_actor")))
	if err != nil {
		t.Fatalf("failed to build context manager: %v", err)
	}

	runtime, err := NewRuntime(RuntimeOptions{
		Injector:       injector,
		ContextManager: contextManager,
	})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	topic := Topic[runtimeTestPayload]{
		Name:          TopicName("runtime.test.event"),
		SchemaVersion: 2,
	}

	if err := (Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var observed string

	if _, err := (Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.listener",
		Handle: func(handlerContext HandlerContext, payload runtimeTestPayload) error {
			formatter := MustResolveFromContext[*runtimeTestFormatter](handlerContext)

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
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	emitContext := contextx.With(context.Background(), runtimeTestActor{ID: "actor-1"})
	emitContext = WithFlag(emitContext, ContextFlagWorkflowBypass)

	receipt := EmitTyped(emitContext, runtime, topic, runtimeTestPayload{Message: "hello"})
	if receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if !receipt.Accepted {
		t.Fatalf("expected accepted receipt")
	}

	if receipt.EventID == "" {
		t.Fatalf("expected event id")
	}

	if observed != "fmt:hello:actor-1" {
		t.Fatalf("unexpected listener output: %s", observed)
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
	runtime, err := NewRuntime(RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.decode")}
	if err := (Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	err = runtime.DispatchEnvelope(context.Background(), Envelope{
		ID:            NewEventID(),
		Topic:         topic.Name,
		SchemaVersion: topic.EffectiveSchemaVersion(),
		Payload:       json.RawMessage("{bad"),
	})
	if !errors.Is(err, ErrPayloadDecodeFailed) {
		t.Fatalf("expected ErrPayloadDecodeFailed, got %v", err)
	}
}

// TestRuntimeEmitDurableModeUsesDurableDispatcher verifies durable mode enqueues
// without invoking inline listeners on the emit call path.
func TestRuntimeEmitDurableModeUsesDurableDispatcher(t *testing.T) {
	dispatcher := &runtimeTestDurableDispatcher{}
	runtime, err := NewRuntime(RuntimeOptions{DurableDispatcher: dispatcher})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.durable")}
	if err := (Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
		Policy: TopicPolicy{
			EmitMode: EmitModeDurable,
		},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	inlineCalls := 0
	if _, err := (Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.durable.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			inlineCalls++

			return nil
		},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	receipt := EmitTyped(context.Background(), runtime, topic, runtimeTestPayload{Message: "durable"})
	if receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if dispatcher.calls != 1 {
		t.Fatalf("expected 1 durable dispatch call, got %d", dispatcher.calls)
	}

	if inlineCalls != 0 {
		t.Fatalf("expected 0 inline calls, got %d", inlineCalls)
	}
}

// TestRuntimeEmitDualModeDispatchesDurableAndInline verifies dual mode dispatches
// to both durable and inline listener paths.
func TestRuntimeEmitDualModeDispatchesDurableAndInline(t *testing.T) {
	dispatcher := &runtimeTestDurableDispatcher{}
	runtime, err := NewRuntime(RuntimeOptions{DurableDispatcher: dispatcher})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.dual")}
	if err := (Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
		Policy: TopicPolicy{
			EmitMode: EmitModeDual,
		},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	inlineCalls := 0
	if _, err := (Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "runtime.test.dual.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			inlineCalls++

			return nil
		},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	receipt := EmitTyped(context.Background(), runtime, topic, runtimeTestPayload{Message: "dual"})
	if receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if dispatcher.calls != 1 {
		t.Fatalf("expected 1 durable dispatch call, got %d", dispatcher.calls)
	}

	if inlineCalls != 1 {
		t.Fatalf("expected 1 inline call, got %d", inlineCalls)
	}
}

// TestRuntimeEmitDurableModeRequiresDurableDispatcher verifies durable mode fails
// when runtime has no durable dispatcher configured.
func TestRuntimeEmitDurableModeRequiresDurableDispatcher(t *testing.T) {
	runtime, err := NewRuntime(RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.durable.required")}
	if err := (Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
		Policy: TopicPolicy{
			EmitMode: EmitModeDurable,
		},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	receipt := EmitTyped(context.Background(), runtime, topic, runtimeTestPayload{Message: "durable"})
	if !errors.Is(receipt.Err, ErrDurableDispatcherRequired) {
		t.Fatalf("expected ErrDurableDispatcherRequired, got %v", receipt.Err)
	}

	if receipt.Accepted {
		t.Fatalf("expected non-accepted receipt")
	}
}

// TestRuntimeEmitEnvelopeUsesPrebuiltEventID verifies pre-built envelopes preserve the caller event ID.
func TestRuntimeEmitEnvelopeUsesPrebuiltEventID(t *testing.T) {
	dispatcher := &runtimeTestDurableDispatcher{}
	runtime, err := NewRuntime(RuntimeOptions{DurableDispatcher: dispatcher})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	topic := Topic[runtimeTestPayload]{Name: TopicName("runtime.test.envelope")}
	if err := (Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
		Policy: TopicPolicy{
			EmitMode: EmitModeDurable,
		},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	encodedPayload, _, err := runtime.Registry().EncodePayload(context.Background(), topic.Name, runtimeTestPayload{Message: "prebuilt"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	prebuiltID := EventID("evt_prebuilt_123")
	err = runtime.EmitEnvelope(context.Background(), Envelope{
		ID:            prebuiltID,
		Topic:         topic.Name,
		SchemaVersion: topic.EffectiveSchemaVersion(),
		Payload:       encodedPayload,
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
