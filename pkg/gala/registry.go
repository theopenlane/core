package gala

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/riverqueue/river"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// registry stores topic codecs, policies, and listeners
type registry struct {
	mu sync.RWMutex
	// topics stores topic metadata and codec wrappers by topic name
	topics map[TopicName]topicRegistration
	// listeners stores registered listeners by topic name
	listeners map[TopicName][]registeredListener
	// detached retains listener topics for retryable durable cleanup
	detached map[ListenerID]TopicName
}

// topicRegistration stores non-generic topic metadata and payload wrappers
type topicRegistration struct {
	// encode JSON-serializes non-generic payloads for the topic
	encode func(any) ([]byte, error)
	// decode JSON-deserializes payload bytes into the topic's payload type
	decode func([]byte) (any, error)
	// uniqueKey derives the uniqueness key for non-generic payloads
	uniqueKey func(any) string
	// kind is the job kind emissions on the topic dispatch under
	kind string
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
	// handle wraps the definition's Handle for non-generic payloads
	handle func(HandlerContext, any, string) error
}

// newRegistry creates an empty topic/listener registry
func newRegistry() *registry {
	return &registry{
		topics:    map[TopicName]topicRegistration{},
		listeners: map[TopicName][]registeredListener{},
		detached:  map[ListenerID]TopicName{},
	}
}

// registerTopic registers one typed topic and its payload wrappers in the registry
func registerTopic[T any](registry *registry, topic Topic[T]) error {
	if registry == nil {
		return ErrRegistryRequired
	}

	if topic.Name == "" {
		return ErrTopicNameRequired
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.topics[topic.Name]; exists {
		return ErrTopicAlreadyRegistered
	}

	registry.topics[topic.Name] = topicRegistration{
		encode:    encodeTopicPayload[T],
		decode:    decodeTopicPayload[T],
		uniqueKey: wrapTopicUniqueKey(topic),
		kind:      topic.Kind,
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

	if definition.Schedule != nil && g.dispatchMode == DispatchModeInMemory {
		return "", nil
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
		id:             listenerID,
		name:           name + "#" + string(listenerID),
		definitionName: name,
		ops:            normalizeOperations(definition.Operations),
		handle:         wrapDefinitionHandle(g, definition),
	}

	g.registry.listeners[topic] = append(g.registry.listeners[topic], listener)

	return listenerID, nil
}

// wrapDefinitionHandle builds the non-generic dispatch wrapper for one definition
func wrapDefinitionHandle[T any](g *Gala, definition Definition[T]) func(HandlerContext, any, string) error {
	handler := definition.Handle
	if definition.Schedule != nil {
		handler = scheduleHandler(g, definition)
	}

	return func(handlerCtx HandlerContext, payload any, operation string) error {
		typedPayload, ok := payload.(T)
		if !ok {
			return ErrPayloadTypeMismatch
		}

		if definition.Gate != nil && !definition.Gate(handlerCtx.Context, typedPayload) {
			return ErrListenerGated
		}

		caller, _ := auth.CallerFromContext(handlerCtx.Context)

		if definition.Caller != nil {
			caller = definition.Caller(caller, typedPayload)
			handlerCtx.Context = auth.WithCaller(handlerCtx.Context, caller)
		}

		handlerCtx.Caller = caller

		handlerCtx.Context = logx.WithCallerIdentity(handlerCtx.Context)
		handlerCtx.Context = logx.WithFields(handlerCtx.Context, map[string]any{
			"event_id":  string(handlerCtx.Envelope.ID),
			"topic":     string(handlerCtx.Envelope.Topic),
			"operation": operation,
		})

		if definition.LogFields != nil {
			handlerCtx.Context = logx.WithFields(handlerCtx.Context, definition.LogFields(typedPayload))
		}

		for _, key := range definition.ContextKeys {
			handlerCtx.Context = key(handlerCtx.Context)
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

	return lo.Map(listeners, func(l registeredListener, _ int) string {
		return l.name
	})
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

// detachListeners removes listeners and retains their topics until cleanup succeeds
func (r *registry) detachListeners(ids []ListenerID) []TopicName {
	if len(ids) == 0 {
		return nil
	}

	targets := lo.Keyify(lo.Filter(ids, func(id ListenerID, _ int) bool {
		return id != ""
	}))
	if len(targets) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	affected := map[TopicName]struct{}{}
	for id := range targets {
		if topic, ok := r.detached[id]; ok {
			affected[topic] = struct{}{}
		}
	}

	for topic, listeners := range r.listeners {
		kept := make([]registeredListener, 0, len(listeners))

		for _, listener := range listeners {
			if _, remove := targets[listener.id]; !remove {
				kept = append(kept, listener)

				continue
			}

			r.detached[listener.id] = topic
			affected[topic] = struct{}{}
		}

		if len(kept) == 0 {
			delete(r.listeners, topic)

			continue
		}

		r.listeners[topic] = kept
	}

	return lo.Keys(affected)
}

func (r *registry) completeListenerRemoval(ids []ListenerID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range ids {
		delete(r.detached, id)
	}
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
		if listenerMatches(listener, operation) {
			return true
		}
	}

	return false
}

// listenerMatches reports whether a listener matches the trimmed operation
func listenerMatches(listener registeredListener, operation string) bool {
	if len(listener.ops) == 0 {
		return true
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
	case definition.Schedule != nil && definition.Schedule.State == nil:
		return ErrListenerScheduleStateRequired
	case definition.Schedule != nil && definition.Schedule.Wrap == nil:
		return ErrListenerScheduleWrapRequired
	case definition.Schedule == nil && definition.Handle == nil:
		return ErrListenerHandlerRequired
	}

	return nil
}

// normalizeOperations normalizes operation filters for one listener registration
func normalizeOperations(operations []string) map[string]struct{} {
	trimmed := lo.FilterMap(operations, func(operation string, _ int) (string, bool) {
		operation = strings.TrimSpace(operation)

		return operation, operation != ""
	})

	if len(trimmed) == 0 {
		return nil
	}

	return lo.Keyify(trimmed)
}

// encodeTopicPayload narrows the payload (any -> T) and JSON-serializes it for registry storage
func encodeTopicPayload[T any](payload any) ([]byte, error) {
	typedPayload, ok := payload.(T)
	if !ok {
		return nil, ErrPayloadTypeMismatch
	}

	encoded, err := json.Marshal(typedPayload)
	if err != nil {
		return nil, ErrPayloadEncodeFailed
	}

	return encoded, nil
}

// decodeTopicPayload JSON-deserializes payload bytes and widens the result (T -> any) for registry storage
func decodeTopicPayload[T any](payload []byte) (any, error) {
	if len(payload) == 0 {
		return nil, ErrEnvelopePayloadRequired
	}

	var decoded T
	if err := jsonx.RoundTrip(payload, &decoded); err != nil {
		return nil, ErrPayloadDecodeFailed
	}

	return decoded, nil
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
