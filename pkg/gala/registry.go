package gala

import (
	"strings"
	"sync"
)

// Registration ties a typed topic to its codec
type Registration[T any] struct {
	// Topic defines the typed topic contract
	Topic Topic[T]
	// Codec serializes and deserializes payloads for the topic
	Codec Codec[T]
}

// Registry stores topic codecs, policies, and listeners
type Registry struct {
	mu sync.RWMutex
	// topics stores topic metadata and codec wrappers by topic name
	topics map[TopicName]topicRegistration
	// listeners stores registered listeners by topic name
	listeners map[TopicName][]registeredListener
}

// topicRegistration stores non-generic topic metadata and codec wrappers
type topicRegistration struct {
	// encode is a wrapper around the topic codec's Encode method for non-generic payloads
	encode func(any) ([]byte, error)
	// decode is a wrapper around the topic codec's Decode method for non-generic payloads
	decode func([]byte) (any, error)
}

// registeredListener stores non-generic listener wrappers
type registeredListener struct {
	// id is the unique identifier for this listener
	id ListenerID
	// name is the human-friendly name for this listener #mitb
	name string
	// ops is the set of operations this listener is interested in, empty means topic-level interest
	ops map[string]struct{}
	// handle is a wrapper around the listener definition's Handle method for non-generic payloads
	handle func(HandlerContext, any) error
}

// NewRegistry creates an empty topic/listener registry
func NewRegistry() *Registry {
	return &Registry{
		topics:    map[TopicName]topicRegistration{},
		listeners: map[TopicName][]registeredListener{},
	}
}

// RegisterTopic registers one typed topic in the registry
func RegisterTopic[T any](registry *Registry, registration Registration[T]) error {
	if registry == nil {
		return ErrRegistryRequired
	}

	if err := validateTopicRegistration(registration); err != nil {
		return err
	}

	topic := registration.Topic.Name

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.topics[topic]; exists {
		return ErrTopicAlreadyRegistered
	}

	registry.topics[topic] = topicRegistration{
		encode: wrapTopicEncoder(registration),
		decode: wrapTopicDecoder(registration),
	}

	return nil
}

// AttachListener registers one typed listener in the registry
func AttachListener[T any](registry *Registry, definition Definition[T]) (ListenerID, error) {
	if registry == nil {
		return "", ErrRegistryRequired
	}

	if err := validateListenerDefinition(definition); err != nil {
		return "", err
	}

	topic := definition.Topic.Name

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.topics[topic]; !exists {
		return "", ErrListenerTopicNotRegistered
	}

	listenerID := ListenerID(NewEventID())

	listener := registeredListener{
		id:   listenerID,
		name: definition.Name,
		ops:  normalizeOperations(definition.Operations),
		handle: func(handlerCtx HandlerContext, payload any) error {
			typedPayload, ok := payload.(T)
			if !ok {
				return ErrPayloadTypeMismatch
			}

			return definition.Handle(handlerCtx, typedPayload)
		},
	}

	registry.listeners[topic] = append(registry.listeners[topic], listener)

	return listenerID, nil
}

// EncodePayload encodes a payload for a registered topic
func (r *Registry) EncodePayload(topic TopicName, payload any) ([]byte, error) {
	registration, err := r.topicRegistration(topic)
	if err != nil {
		return nil, err
	}

	encoded, err := registration.encode(payload)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

// DecodePayload decodes payload bytes for a registered topic
func (r *Registry) DecodePayload(topic TopicName, payload []byte) (any, error) {
	registration, err := r.topicRegistration(topic)
	if err != nil {
		return nil, err
	}

	return registration.decode(payload)
}

// registeredListeners returns a snapshot of listeners for one topic.
func (r *Registry) registeredListeners(topic TopicName) []registeredListener {
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

// InterestedIn reports whether any listener is registered for topic+operation.
// Empty operation means topic-level interest only.
func (r *Registry) InterestedIn(topic TopicName, operation string) bool {
	if topic == "" {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	listeners := r.listeners[topic]
	if len(listeners) == 0 {
		return false
	}

	operation = strings.TrimSpace(operation)
	if operation == "" {
		return true
	}

	for _, listener := range listeners {
		if listenerInterestedInOperation(listener, operation) {
			return true
		}
	}

	return false
}

// listenerInterestedInOperation reports whether a listener matches an operation filter.
// Callers must pass a trimmed operation string.
func listenerInterestedInOperation(listener registeredListener, operation string) bool {
	if len(listener.ops) == 0 {
		return true
	}

	if operation == "" {
		return false
	}

	_, ok := listener.ops[operation]

	return ok
}

// topicRegistration resolves one topic registration by name
func (r *Registry) topicRegistration(topic TopicName) (topicRegistration, error) {
	if topic == "" {
		return topicRegistration{}, ErrTopicNameRequired
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	registration, exists := r.topics[topic]
	if !exists {
		return topicRegistration{}, ErrTopicNotRegistered
	}

	return registration, nil
}

// validateTopicRegistration validates topic registration requirements
func validateTopicRegistration[T any](registration Registration[T]) error {
	if registration.Topic.Name == "" {
		return ErrTopicNameRequired
	}

	if registration.Codec == nil {
		return ErrCodecRequired
	}

	return nil
}

// validateListenerDefinition validates listener definition requirements
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

// normalizeOperations normalizes operation filters for one listener registration
func normalizeOperations(operations []string) map[string]struct{} {
	if len(operations) == 0 {
		return nil
	}

	normalized := map[string]struct{}{}
	for _, operation := range operations {
		operation = strings.TrimSpace(operation)
		if operation == "" {
			continue
		}

		normalized[operation] = struct{}{}
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

// wrapTopicEncoder creates a non-generic encoder wrapper for one topic
func wrapTopicEncoder[T any](registration Registration[T]) func(any) ([]byte, error) {
	return func(payload any) ([]byte, error) {
		typedPayload, ok := payload.(T)
		if !ok {
			return nil, ErrPayloadTypeMismatch
		}

		encoded, err := registration.Codec.Encode(typedPayload)
		if err != nil {
			return nil, ErrPayloadEncodeFailed
		}

		return encoded, nil
	}
}

// wrapTopicDecoder creates a non-generic decoder wrapper for one topic
func wrapTopicDecoder[T any](registration Registration[T]) func([]byte) (any, error) {
	return func(payload []byte) (any, error) {
		decoded, err := registration.Codec.Decode(payload)
		if err != nil {
			return nil, ErrPayloadDecodeFailed
		}

		return decoded, nil
	}
}
