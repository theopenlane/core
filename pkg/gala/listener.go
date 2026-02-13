package gala

import (
	"context"
	"errors"

	"github.com/samber/do/v2"
)

// ListenerID identifies a registered listener
type ListenerID string

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

// Definition defines one listener binding
type Definition[T any] struct {
	// Topic is the topic handled by this listener
	Topic Topic[T]
	// Name is the stable listener name
	Name string
	// Handle is the callback invoked for this listener
	Handle Handler[T]
}

// Register registers the listener definition against a registry
func (d Definition[T]) Register(registry *Registry) (ListenerID, error) {
	return AttachListener(registry, d)
}

// ListenerRegistration captures topic and listener registration in one operation
type ListenerRegistration[T any] struct {
	// Topic is the topic handled by this listener
	Topic Topic[T]
	// Codec serializes and deserializes payloads for the topic. Defaults to JSONCodec[T] when nil
	Codec Codec[T]
	// Policy defines dispatch behavior for the topic
	Policy TopicPolicy
	// Name is the stable listener name
	Name string
	// Handle is the callback invoked for this listener
	Handle Handler[T]
}

// RegisterListener registers a topic (idempotent) and listener in a single call
func RegisterListener[T any](registry *Registry, registration ListenerRegistration[T]) (ListenerID, error) {
	codec := registration.Codec
	if codec == nil {
		codec = JSONCodec[T]{}
	}

	err := RegisterTopic(registry, Registration[T]{
		Topic:  registration.Topic,
		Codec:  codec,
		Policy: registration.Policy,
	})
	if err != nil && !errors.Is(err, ErrTopicAlreadyRegistered) {
		return "", err
	}

	return AttachListener(registry, Definition[T]{
		Topic:  registration.Topic,
		Name:   registration.Name,
		Handle: registration.Handle,
	})
}

// RegisterListeners registers multiple listeners and their topics in order
func RegisterListeners[T any](registry *Registry, registrations ...ListenerRegistration[T]) ([]ListenerID, error) {
	ids := make([]ListenerID, 0, len(registrations))
	for _, registration := range registrations {
		id, err := RegisterListener(registry, registration)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// RegisterDurableListeners registers listeners as durable topics with one shared queue class
func RegisterDurableListeners[T any](registry *Registry, queueClass QueueClass, definitions ...Definition[T]) ([]ListenerID, error) {
	registrations := make([]ListenerRegistration[T], 0, len(definitions))
	for _, definition := range definitions {
		registrations = append(registrations, ListenerRegistration[T]{
			Topic:  definition.Topic,
			Name:   definition.Name,
			Handle: definition.Handle,
			Policy: TopicPolicy{
				EmitMode:   EmitModeDurable,
				QueueClass: queueClass,
			},
		})
	}

	return RegisterListeners(registry, registrations...)
}
