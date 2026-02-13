package gala

import (
	"context"

	"github.com/samber/do/v2"
)

// ListenerID identifies a registered listener.
type ListenerID string

// HandlerContext provides event context and dependency resolution scope for listeners.
type HandlerContext struct {
	// Context is the restored event context used for listener execution.
	Context context.Context
	// Envelope is the envelope being processed.
	Envelope Envelope
	// Injector provides typed dependency lookup via samber/do.
	Injector do.Injector
}

// Handler processes a typed event payload.
type Handler[T any] func(HandlerContext, T) error

// Definition defines one listener binding.
type Definition[T any] struct {
	// Topic is the topic handled by this listener.
	Topic Topic[T]
	// Name is the stable listener name.
	Name string
	// Handle is the callback invoked for this listener.
	Handle Handler[T]
}

// Register registers the listener definition against a registry.
func (d Definition[T]) Register(registry *Registry) (ListenerID, error) {
	return AttachListener(registry, d)
}
