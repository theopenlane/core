package gala

import (
	"strings"
	"sync"

	"github.com/riverqueue/river"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/pkg/logx"
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
	// uniqueKey derives the uniqueness key for non-generic payloads, nil when none declared
	uniqueKey func(any) string
}

// registeredListener stores non-generic listener wrappers
type registeredListener struct {
	// id is the unique identifier for this listener
	id ListenerID
	// name is the human-friendly name for this listener #mitb
	name string
	// definitionName is the unsuffixed listener name used as the metrics label
	definitionName string
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
		encode:    wrapTopicEncoder(codec),
		decode:    wrapTopicDecoder(codec),
		uniqueKey: wrapTopicUniqueKey(topic),
	}

	return nil
}

// attachListener registers one typed listener on the gala runtime
func attachListener[T any](g *Gala, definition Definition[T]) (ListenerID, error) {
	if g == nil {
		return "", ErrGalaRequired
	}

	if err := validateListenerDefinition(definition); err != nil {
		return "", err
	}

	name := definition.Name
	if name == "" {
		name = string(definition.Topic.Name)
	}

	topic := definition.Topic.Name

	g.registry.mu.Lock()
	defer g.registry.mu.Unlock()

	if _, exists := g.registry.topics[topic]; !exists {
		return "", ErrListenerTopicNotRegistered
	}

	listenerID := ListenerID(NewEventID())

	listener := registeredListener{
		id: listenerID,
		// the id suffix keeps observability output attributable when multiple listeners
		// share a name on one topic
		name:           name + "#" + string(listenerID),
		definitionName: name,
		ops:            normalizeOperations(definition.Operations),
		handle:         wrapDefinitionHandle(g, definition),
	}

	g.registry.listeners[topic] = append(g.registry.listeners[topic], listener)

	return listenerID, nil
}

// wrapDefinitionHandle builds the non-generic dispatch wrapper for one definition, applying
// the dispatch pipeline in order: Gate, caller resolution, automatic log enrichment, Enrich,
// Elevate, then the handler, classifying handler errors through Cancel
func wrapDefinitionHandle[T any](g *Gala, definition Definition[T]) func(HandlerContext, any) error {
	handler := definition.Handle
	if definition.Schedule != nil {
		handler = scheduleHandler(g, definition)
	}

	return func(handlerCtx HandlerContext, payload any) error {
		typedPayload, ok := payload.(T)
		if !ok {
			return ErrPayloadTypeMismatch
		}

		if definition.Gate != nil && !definition.Gate(typedPayload) {
			return nil
		}

		caller, found := auth.CallerFromContext(handlerCtx.Context)
		if !found || caller == nil {
			caller = &auth.Caller{}
		}

		if definition.Caller != nil {
			caller = definition.Caller(caller, typedPayload)
		}

		handlerCtx.Context = auth.WithCaller(handlerCtx.Context, caller)
		handlerCtx.Caller = caller

		handlerCtx.Context = logx.WithFields(handlerCtx.Context, map[string]any{
			"event_id":  string(handlerCtx.Envelope.ID),
			"topic":     string(handlerCtx.Envelope.Topic),
			"operation": payloadOperation(typedPayload),
		})

		if definition.Enrich != nil {
			handlerCtx.Context = definition.Enrich(handlerCtx.Context, typedPayload)
		}

		if definition.Elevate != nil {
			handlerCtx.Context = definition.Elevate(handlerCtx.Context, typedPayload)
		}

		err := handler(handlerCtx, typedPayload)
		// scheduled definitions classify errors through Cancel inside the schedule loop
		if err != nil && definition.Schedule == nil && definition.Cancel != nil && definition.Cancel(handlerCtx.Context, typedPayload, err) {
			return river.JobCancel(err)
		}

		return err
	}
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

	switch {
	case definition.Schedule != nil && definition.Handle != nil:
		return ErrListenerHandlerConflict
	case definition.Schedule != nil && definition.Schedule.Handle == nil:
		return ErrListenerHandlerRequired
	case definition.Schedule == nil && definition.Handle == nil:
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

// wrapTopicUniqueKey narrows the payload type (any -> T) into the topic's UniqueKey derivation
func wrapTopicUniqueKey[T any](topic Topic[T]) func(any) string {
	if topic.UniqueKey == nil {
		return nil
	}

	return func(payload any) string {
		typedPayload, ok := payload.(T)
		if !ok {
			return ""
		}

		return topic.UniqueKey(typedPayload)
	}
}
