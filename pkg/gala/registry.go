package gala

import (
	"context"
	"fmt"
	"sync"
)

// Registration ties a typed topic to its codec and policy.
type Registration[T any] struct {
	// Topic defines the typed topic contract.
	Topic Topic[T]
	// Codec serializes and deserializes payloads for the topic.
	Codec Codec[T]
	// Policy defines dispatch behavior for the topic.
	Policy TopicPolicy
}

// Register registers this topic with the registry.
func (r Registration[T]) Register(registry *Registry) error {
	return RegisterTopic(registry, r)
}

// Registry stores topic codecs, policies, and listeners.
type Registry struct {
	mu        sync.RWMutex
	topics    map[TopicName]topicRegistration
	listeners map[TopicName][]registeredListener
}

// topicRegistration stores non-generic topic metadata and codec wrappers.
type topicRegistration struct {
	policy        TopicPolicy
	schemaVersion int
	encode        func(context.Context, any) ([]byte, error)
	decode        func(context.Context, []byte) (any, error)
}

// registeredListener stores non-generic listener wrappers.
type registeredListener struct {
	id     ListenerID
	name   string
	handle func(HandlerContext, any) error
}

// NewRegistry creates an empty topic/listener registry.
func NewRegistry() *Registry {
	return &Registry{
		topics:    map[TopicName]topicRegistration{},
		listeners: map[TopicName][]registeredListener{},
	}
}

// RegisterTopic registers one typed topic in the registry.
func RegisterTopic[T any](registry *Registry, registration Registration[T]) error {
	if registry == nil {
		return ErrRuntimeRequired
	}

	if err := validateTopicRegistration(registration); err != nil {
		return err
	}

	topic := registration.Topic.Name

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.topics[topic]; exists {
		return fmt.Errorf("%w: %s", ErrTopicAlreadyRegistered, topic)
	}

	registry.topics[topic] = topicRegistration{
		policy:        registration.Policy,
		schemaVersion: registration.Topic.EffectiveSchemaVersion(),
		encode:        wrapTopicEncoder(registration),
		decode:        wrapTopicDecoder(registration),
	}

	return nil
}

// AttachListener registers one typed listener in the registry.
func AttachListener[T any](registry *Registry, definition Definition[T]) (ListenerID, error) {
	if registry == nil {
		return "", ErrRuntimeRequired
	}

	if err := validateListenerDefinition(definition); err != nil {
		return "", err
	}

	topic := definition.Topic.Name

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.topics[topic]; !exists {
		return "", fmt.Errorf("%w: %s", ErrListenerTopicNotRegistered, topic)
	}

	listenerID := ListenerID(NewEventID())
	listener := registeredListener{
		id:   listenerID,
		name: definition.Name,
		handle: func(handlerCtx HandlerContext, payload any) error {
			typedPayload, ok := payload.(T)
			if !ok {
				return fmt.Errorf("%w: listener=%s topic=%s", ErrPayloadTypeMismatch, definition.Name, topic)
			}

			return definition.Handle(handlerCtx, typedPayload)
		},
	}

	registry.listeners[topic] = append(registry.listeners[topic], listener)

	return listenerID, nil
}

// TopicPolicy returns policy metadata for a topic.
func (r *Registry) TopicPolicy(topic TopicName) (TopicPolicy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registration, exists := r.topics[topic]
	if !exists {
		return TopicPolicy{}, false
	}

	return registration.policy, true
}

// EncodePayload encodes a payload for a registered topic and returns schema version.
func (r *Registry) EncodePayload(ctx context.Context, topic TopicName, payload any) ([]byte, int, error) {
	registration, err := r.topicRegistration(topic)
	if err != nil {
		return nil, 0, err
	}

	encoded, err := registration.encode(ctx, payload)
	if err != nil {
		return nil, 0, err
	}

	return encoded, registration.schemaVersion, nil
}

// DecodePayload decodes payload bytes for a registered topic.
func (r *Registry) DecodePayload(ctx context.Context, topic TopicName, payload []byte) (any, error) {
	registration, err := r.topicRegistration(topic)
	if err != nil {
		return nil, err
	}

	return registration.decode(ctx, payload)
}

// Listeners returns a snapshot of listeners for one topic.
func (r *Registry) Listeners(topic TopicName) []registeredListener {
	r.mu.RLock()
	defer r.mu.RUnlock()

	listeners := r.listeners[topic]
	if len(listeners) == 0 {
		return nil
	}

	copied := make([]registeredListener, len(listeners))
	copy(copied, listeners)

	return copied
}

// topicRegistration resolves one topic registration by name.
func (r *Registry) topicRegistration(topic TopicName) (topicRegistration, error) {
	if topic == "" {
		return topicRegistration{}, ErrTopicNameRequired
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	registration, exists := r.topics[topic]
	if !exists {
		return topicRegistration{}, fmt.Errorf("%w: %s", ErrTopicNotRegistered, topic)
	}

	return registration, nil
}

// validateTopicRegistration validates topic registration requirements.
func validateTopicRegistration[T any](registration Registration[T]) error {
	if registration.Topic.Name == "" {
		return ErrTopicNameRequired
	}

	if registration.Codec == nil {
		return ErrCodecRequired
	}

	return nil
}

// validateListenerDefinition validates listener definition requirements.
func validateListenerDefinition[T any](definition Definition[T]) error {
	if definition.Topic.Name == "" {
		return ErrTopicNameRequired
	}

	if definition.Name == "" {
		return ErrListenerNameRequired
	}

	if definition.Handle == nil {
		return ErrListenerHandlerRequired
	}

	return nil
}

// wrapTopicEncoder creates a non-generic encoder wrapper for one topic.
func wrapTopicEncoder[T any](registration Registration[T]) func(context.Context, any) ([]byte, error) {
	return func(ctx context.Context, payload any) ([]byte, error) {
		typedPayload, ok := payload.(T)
		if !ok {
			return nil, fmt.Errorf("%w: topic=%s", ErrPayloadTypeMismatch, registration.Topic.Name)
		}

		encoded, err := registration.Codec.Encode(ctx, typedPayload)
		if err != nil {
			return nil, err
		}

		return encoded, nil
	}
}

// wrapTopicDecoder creates a non-generic decoder wrapper for one topic.
func wrapTopicDecoder[T any](registration Registration[T]) func(context.Context, []byte) (any, error) {
	return func(ctx context.Context, payload []byte) (any, error) {
		decoded, err := registration.Codec.Decode(ctx, payload)
		if err != nil {
			return nil, err
		}

		return decoded, nil
	}
}
