package gala

import (
	"context"
	"encoding/json"

	"github.com/samber/do/v2"
)

// injectorCodec restores a dependency from the gala injector onto the handler
// context without serializing the value itself. On capture it stores a sentinel
// marker; on restore it resolves the value from the injector and applies the
// provided setter to place it on the context
type injectorCodec[T any] struct {
	id       ContextKey
	injector do.Injector
	setter   func(context.Context, T) context.Context
}

// newInjectorCodec creates a codec that resolves T from the injector on restore
// and applies setter to attach it to the context
func newInjectorCodec[T any](id ContextKey, injector do.Injector, setter func(context.Context, T) context.Context) injectorCodec[T] {
	return injectorCodec[T]{id: id, injector: injector, setter: setter}
}

// Key returns the stable snapshot identifier
func (c injectorCodec[T]) Key() ContextKey {
	return c.id
}

// Capture stores a sentinel marker so Restore is invoked on the handler side
func (c injectorCodec[T]) Capture(_ context.Context) (json.RawMessage, bool, error) {
	return json.RawMessage(`true`), true, nil
}

// Restore resolves T from the injector and attaches it to the context
func (c injectorCodec[T]) Restore(ctx context.Context, _ json.RawMessage) (context.Context, error) {
	val, err := do.Invoke[T](c.injector)
	if err != nil {
		return ctx, ErrContextSnapshotRestoreFailed
	}

	return c.setter(ctx, val), nil
}
