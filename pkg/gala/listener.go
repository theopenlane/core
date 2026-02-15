package gala

import (
	"context"
	"errors"

	"github.com/samber/do/v2"
	"github.com/theopenlane/utils/ulids"
)

// ListenerID identifies a registered listener
type ListenerID string

// EventID is a stable identifier used for idempotency and traceability
type EventID string

// NewEventID creates a new event identifier
func NewEventID() EventID {
	return EventID(ulids.New().String())
}

// HandlerContext provides event context and dependency resolution scope for listeners
type HandlerContext struct {
	// Context is the restored event context used for listener execution
	Context context.Context
	// Envelope is the envelope being processed
	Envelope Envelope
	// Injector provides typed dependency lookup via samber/do
	Injector do.Injector
}

// Handler processes a typed event payload
type Handler[T any] func(HandlerContext, T) error

// TopicName is the stable string identifier for a topic
type TopicName string

// Topic defines a strongly typed topic contract
type Topic[T any] struct {
	// Name is the stable topic identifier
	Name TopicName
}

// Definition defines one listener binding
type Definition[T any] struct {
	// Topic is the topic handled by this listener
	Topic Topic[T]
	// Name is the stable listener name
	Name string
	// Operations optionally scopes listener interest to specific mutation operations
	// Empty means the listener accepts all operations for the topic
	Operations []string
	// Handle is the callback invoked for this listener
	Handle Handler[T]
}

// RegisterListeners registers listeners and ensures their topic contracts are configured
func RegisterListeners[T any](registry *Registry, definitions ...Definition[T]) ([]ListenerID, error) {
	ids := make([]ListenerID, 0, len(definitions))

	for _, definition := range definitions {
		err := RegisterTopic(registry, Registration[T]{
			Topic: definition.Topic,
			Codec: JSONCodec[T]{},
		})
		if err != nil && !errors.Is(err, ErrTopicAlreadyRegistered) {
			return nil, err
		}

		id, err := AttachListener(registry, definition)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
