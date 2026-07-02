package gala

import (
	"context"
	"encoding/json"

	"github.com/samber/do/v2"
)

// InjectorCodec restores a dependency from the gala injector onto the handler
// context without serializing the value itself. On capture it stores a sentinel
// marker; on restore it resolves the value from the injector and applies the
// provided setter to place it on the context
type InjectorCodec[T any] struct {
	id       ContextKey
	injector do.Injector
	setter   func(context.Context, T) context.Context
}

// NewInjectorCodec creates a codec that resolves T from the injector on restore
// and applies setter to attach it to the context
func NewInjectorCodec[T any](id ContextKey, injector do.Injector, setter func(context.Context, T) context.Context) InjectorCodec[T] {
	return InjectorCodec[T]{id: id, injector: injector, setter: setter}
}

// Key returns the stable snapshot identifier
func (c InjectorCodec[T]) Key() ContextKey {
	return c.id
}

// Capture stores a sentinel marker so Restore is invoked on the handler side
func (c InjectorCodec[T]) Capture(_ context.Context) (json.RawMessage, bool, error) {
	return json.RawMessage(`true`), true, nil
}

// Restore resolves T from the injector and attaches it to the context
func (c InjectorCodec[T]) Restore(ctx context.Context, _ json.RawMessage) (context.Context, error) {
	val, err := do.Invoke[T](c.injector)
	if err != nil {
		return ctx, ErrContextSnapshotRestoreFailed
	}

	return c.setter(ctx, val), nil
}
