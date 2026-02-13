package gala

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/samber/do/v2"
)

// RuntimeOptions configures a gala runtime.
type RuntimeOptions struct {
	// Registry stores topic/listener metadata.
	Registry *Registry
	// Injector provides typed dependencies for listeners.
	Injector do.Injector
	// DurableDispatcher dispatches envelopes through durable infrastructure.
	DurableDispatcher DurableDispatcher
	// ContextManager manages capture/restore of context snapshots.
	ContextManager *ContextManager
}

// Runtime dispatches envelopes using registered topics and listeners.
type Runtime struct {
	registry          *Registry
	injector          do.Injector
	durableDispatcher DurableDispatcher
	contextManager    *ContextManager
}

// NewRuntime creates a gala runtime.
func NewRuntime(options RuntimeOptions) (*Runtime, error) {
	registry := options.Registry
	if registry == nil {
		registry = NewRegistry()
	}

	injector := options.Injector
	if injector == nil {
		injector = NewInjector()
	}

	contextManager := options.ContextManager
	if contextManager == nil {
		manager, err := NewContextManager()
		if err != nil {
			return nil, err
		}

		contextManager = manager
	}

	return &Runtime{
		registry:          registry,
		injector:          injector,
		durableDispatcher: options.DurableDispatcher,
		contextManager:    contextManager,
	}, nil
}

// Registry returns the runtime registry.
func (r *Runtime) Registry() *Registry {
	return r.registry
}

// Injector returns the runtime dependency injector.
func (r *Runtime) Injector() do.Injector {
	return r.injector
}

// ContextManager returns the runtime context manager.
func (r *Runtime) ContextManager() *ContextManager {
	return r.contextManager
}

// DurableDispatcher returns the runtime durable dispatcher.
func (r *Runtime) DurableDispatcher() DurableDispatcher {
	return r.durableDispatcher
}

// Emit emits a payload with default headers.
func (r *Runtime) Emit(ctx context.Context, topic TopicName, payload any) EmitReceipt {
	return r.EmitWithHeaders(ctx, topic, payload, Headers{})
}

// EmitWithHeaders emits a payload with explicit headers.
func (r *Runtime) EmitWithHeaders(ctx context.Context, topic TopicName, payload any, headers Headers) EmitReceipt {
	if r == nil {
		return EmitReceipt{Err: ErrRuntimeRequired}
	}

	policy, _ := r.registry.TopicPolicy(topic)

	encodedPayload, schemaVersion, err := r.registry.EncodePayload(ctx, topic, payload)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	snapshot, err := r.contextManager.Capture(ctx)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	envelope := Envelope{
		ID:              NewEventID(),
		Topic:           topic,
		SchemaVersion:   schemaVersion,
		OccurredAt:      time.Now().UTC(),
		Headers:         headers,
		Payload:         encodedPayload,
		ContextSnapshot: snapshot,
	}

	if err := r.dispatchByPolicy(ctx, envelope, policy); err != nil {
		return EmitReceipt{EventID: envelope.ID, Err: err}
	}

	return EmitReceipt{EventID: envelope.ID, Accepted: true}
}

// EmitEnvelope dispatches a pre-built envelope using the topic policy registered for its topic.
func (r *Runtime) EmitEnvelope(ctx context.Context, envelope Envelope) error {
	if r == nil {
		return ErrRuntimeRequired
	}

	policy, _ := r.registry.TopicPolicy(envelope.Topic)

	return r.dispatchByPolicy(ctx, envelope, policy)
}

// dispatchByPolicy dispatches an envelope based on emit mode semantics.
func (r *Runtime) dispatchByPolicy(ctx context.Context, envelope Envelope, policy TopicPolicy) error {
	mode := policy.EffectiveEmitMode()

	switch mode {
	case EmitModeInline:
		return r.DispatchEnvelope(ctx, envelope)
	case EmitModeDurable:
		return r.dispatchDurable(ctx, envelope, policy)
	case EmitModeDual:
		durableErr := r.dispatchDurable(ctx, envelope, policy)
		inlineErr := r.DispatchEnvelope(ctx, envelope)
		if durableErr != nil || inlineErr != nil {
			return errors.Join(durableErr, inlineErr)
		}

		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedEmitMode, mode)
	}
}

// dispatchDurable dispatches an envelope through the configured durable dispatcher.
func (r *Runtime) dispatchDurable(ctx context.Context, envelope Envelope, policy TopicPolicy) error {
	if r.durableDispatcher == nil {
		return ErrDurableDispatcherRequired
	}

	if err := r.durableDispatcher.DispatchDurable(ctx, envelope, policy); err != nil {
		return fmt.Errorf("%w: %w", ErrDurableDispatchFailed, err)
	}

	return nil
}

// DispatchEnvelope dispatches one envelope to all listeners on the topic.
func (r *Runtime) DispatchEnvelope(ctx context.Context, envelope Envelope) error {
	if r == nil {
		return ErrRuntimeRequired
	}

	if envelope.Topic == "" {
		return ErrEnvelopeTopicRequired
	}

	if len(envelope.Payload) == 0 {
		return ErrEnvelopePayloadRequired
	}

	decodedPayload, err := r.registry.DecodePayload(ctx, envelope.Topic, envelope.Payload)
	if err != nil {
		return err
	}

	restoredContext, err := r.contextManager.Restore(ctx, envelope.ContextSnapshot)
	if err != nil {
		return err
	}

	handlerContext := HandlerContext{
		Context:  restoredContext,
		Envelope: envelope,
		Injector: r.injector,
	}

	listeners := r.registry.Listeners(envelope.Topic)
	for _, listener := range listeners {
		if err := listener.handle(handlerContext, decodedPayload); err != nil {
			return fmt.Errorf("%w: listener=%s topic=%s: %w", ErrListenerExecutionFailed, listener.name, envelope.Topic, err)
		}
	}

	return nil
}

// EmitTyped emits a payload using a typed topic helper.
func EmitTyped[T any](ctx context.Context, runtime *Runtime, topic Topic[T], payload T) EmitReceipt {
	if runtime == nil {
		return EmitReceipt{Err: ErrRuntimeRequired}
	}

	return runtime.Emit(ctx, topic.Name, payload)
}
