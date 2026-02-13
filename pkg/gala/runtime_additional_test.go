package gala

import (
	"context"
	"errors"
	"testing"
)

type codecTestUnsupportedPayload struct {
	Unsupported func()
}

type runtimeAccessorTestDurableDispatcher struct{}

func (runtimeAccessorTestDurableDispatcher) DispatchDurable(context.Context, Envelope, TopicPolicy) error {
	return nil
}

func TestRuntimeAccessorsReturnConfiguredDependencies(t *testing.T) {
	injector := NewInjector()
	contextManager, err := NewContextManager()
	if err != nil {
		t.Fatalf("failed to build context manager: %v", err)
	}

	dispatcher := runtimeAccessorTestDurableDispatcher{}
	runtime, err := NewRuntime(RuntimeOptions{
		Injector:          injector,
		ContextManager:    contextManager,
		DurableDispatcher: dispatcher,
	})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	if runtime.Injector() != injector {
		t.Fatalf("unexpected injector accessor result")
	}

	if runtime.ContextManager() != contextManager {
		t.Fatalf("unexpected context manager accessor result")
	}

	if runtime.DurableDispatcher() != dispatcher {
		t.Fatalf("unexpected durable dispatcher accessor result")
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

func TestRuntimeAccessorsAreNilSafe(t *testing.T) {
	var runtime *Runtime

	if runtime.Registry() != nil {
		t.Fatalf("expected nil registry for nil runtime")
	}

	if runtime.Injector() != nil {
		t.Fatalf("expected nil injector for nil runtime")
	}

	if runtime.ContextManager() != nil {
		t.Fatalf("expected nil context manager for nil runtime")
	}

	if runtime.DurableDispatcher() != nil {
		t.Fatalf("expected nil durable dispatcher for nil runtime")
	}
}
