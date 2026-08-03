package gala

import (
	"strings"
	"sync"
)

// Registry stores topic codecs, policies, and listeners
type registry struct {
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

// newRegistry creates an empty topic/listener registry
func newRegistry() *registry {
	return &registry{
		topics:    map[TopicName]topicRegistration{},
		listeners: map[TopicName][]registeredListener{},
	}
}

// registerTopic registers one typed topic and its codec in the registry
func registerTopic[T any](registry *registry, topic Topic[T], codec Codec[T]) error {
	if registry == nil {
		return ErrRegistryRequired
	}

	if topic.Name == "" {
		return ErrTopicNameRequired
	}

	if codec == nil {
		return ErrCodecRequired
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.topics[topic.Name]; exists {
		return ErrTopicAlreadyRegistered
	}

	registry.topics[topic.Name] = topicRegistration{
		encode: wrapTopicEncoder(codec),
		decode: wrapTopicDecoder(codec),
	}

	return nil
}

// attachListener registers one typed listener in the registry
func attachListener[T any](registry *registry, definition Definition[T]) (ListenerID, error) {
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

// listenerNamesForTopic returns the registered listener names for a topic
func (r *registry) listenerNamesForTopic(topic TopicName) []string {
	listeners := r.registeredListeners(topic)
	if len(listeners) == 0 {
		return nil
	}

	names := make([]string, len(listeners))
	for i, l := range listeners {
		names[i] = l.name
	}

	return names
}

// registeredListeners returns a snapshot of listeners for one topic.
func (r *registry) registeredListeners(topic TopicName) []registeredListener {
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

// InterestedIn reports whether any listener is registered for topic+operation
// Empty operation means topic-level interest only
func (r *registry) InterestedIn(topic TopicName, operation string) bool {
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

// listenerInterestedInOperation reports whether a listener matches an operation filter
// Callers must pass a trimmed operation string
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
func (r *registry) topicRegistration(topic TopicName) (topicRegistration, error) {
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

// wrapTopicEncoder coverts a type-specific codec into a common shape for registry storage (any -> T)
func wrapTopicEncoder[T any](codec Codec[T]) func(any) ([]byte, error) {
	return func(payload any) ([]byte, error) {
		typedPayload, ok := payload.(T)
		if !ok {
			return nil, ErrPayloadTypeMismatch
		}

		encoded, err := codec.Encode(typedPayload)
		if err != nil {
			return nil, ErrPayloadEncodeFailed
		}

		return encoded, nil
	}
}

// wrapTopicDecoder wides the return type on the way out from the encoder (T -> any) for registry storage
func wrapTopicDecoder[T any](codec Codec[T]) func([]byte) (any, error) {
	return func(payload []byte) (any, error) {
		decoded, err := codec.Decode(payload)
		if err != nil {
			return nil, ErrPayloadDecodeFailed
		}

		return decoded, nil
	}
}
