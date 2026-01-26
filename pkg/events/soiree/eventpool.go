package soiree

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/rs/zerolog/log"
)

// Emitter defines the minimal interface for emitting events
type Emitter interface {
	Emit(topic string, payload any) <-chan error
}

// EventBus manages subscribing and unsubscribing listeners to topics and emitting events to subscribers
type EventBus struct {
	topics            sync.Map
	errorHandler      func(Event, error) error
	idGenerator       func() string
	panicHandler      PanicHandler
	pool              *Pool
	poolOpts          []PoolOption
	closed            atomic.Bool
	errChanBufferSize int
	client            any
	store             eventStore
	queueCancel       context.CancelFunc
	maxRetries        int
	backOffFactory    func() backoff.BackOff
}

// New initializes a new EventBus with optional configuration options
func New(opts ...Option) *EventBus {
	m := &EventBus{
		topics:            sync.Map{},
		errorHandler:      defaultErrorHandler,
		idGenerator:       defaultIDGenerator,
		panicHandler:      defaultPanicHandler,
		errChanBufferSize: 10, //nolint:mnd
		maxRetries:        1,
		backOffFactory: func() backoff.BackOff {
			return backoff.NewExponentialBackOff()
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	m.pool = NewPool(m.poolOpts...)

	register(m)

	if q, ok := m.store.(eventQueue); ok {
		m.startQueueConsumer(q)
	}

	return m
}

// Client returns the client set on the event bus
func (m *EventBus) Client() any {
	return m.client
}

// InterestedIn checks if the event bus has any listeners registered for the given topic
func (m *EventBus) InterestedIn(topicName string) bool {
	topicName = normalizeTopicName(topicName)
	if err := validateTopicName(topicName); err != nil {
		return false
	}

	if t, ok := m.topics.Load(topicName); ok {
		if t.(*topic).hasListeners() {
			return true
		}
	}

	interested := false
	m.topics.Range(func(key, value any) bool {
		pattern := key.(string)
		if pattern == topicName {
			return true
		}
		if matchTopicPattern(pattern, topicName) && value.(*topic).hasListeners() {
			interested = true
			return false
		}

		return true
	})

	return interested
}

// RegisterListeners registers multiple listener bindings and returns their IDs
func (m *EventBus) RegisterListeners(bindings ...ListenerBinding) ([]string, error) {
	ids := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		id, err := binding.Register(m)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// On subscribes a listener to a topic with the given name and returns a unique listener ID
func (m *EventBus) On(topicName string, listener Listener) (string, error) {
	if listener == nil {
		return "", ErrNilListener
	}

	topicName = normalizeTopicName(topicName)
	if err := validateTopicName(topicName); err != nil {
		return "", err
	}

	t := m.ensureTopic(topicName)
	listenerID := m.idGenerator()
	t.addListener(listenerID, listener)

	return listenerID, nil
}

// Off unsubscribes a listener from a topic using the listener's unique ID
func (m *EventBus) Off(topicName string, listenerID string) error {
	topicName = normalizeTopicName(topicName)
	if err := validateTopicName(topicName); err != nil {
		return err
	}

	t, err := m.getTopic(topicName)
	if err != nil {
		return err
	}

	return t.removeListener(listenerID)
}

// Emit asynchronously dispatches an event to all subscribers of the event's topic
func (m *EventBus) Emit(eventName string, payload any) <-chan error {
	return m.EmitWithContext(context.Background(), eventName, payload)
}

// EmitWithContext asynchronously dispatches an event with the given context for timeout/cancellation control
func (m *EventBus) EmitWithContext(ctx context.Context, eventName string, payload any) <-chan error {
	errChan := make(chan error, m.errChanBufferSize)

	if m.closed.Load() {
		m.trySendErr(errChan, ErrEmitterClosed)
		close(errChan)
		return errChan
	}

	topicName, event, prepErr := m.prepareEvent(eventName, payload)
	if prepErr != nil {
		m.trySendErr(errChan, prepErr)
		close(errChan)

		return errChan
	}

	if ctx != context.Background() {
		event.SetContext(ctx)
	}

	if m.store != nil {
		if err := m.store.SaveEvent(event); err != nil {
			m.trySendErr(errChan, err)
		}
	}

	if _, ok := m.store.(eventQueue); ok {
		close(errChan)

		return errChan
	}

	m.pool.Submit(func() {
		defer close(errChan)
		m.handleEvents(topicName, event, func(err error) {
			m.trySendErr(errChan, err)
		})
	})

	return errChan
}

// handleEvents dispatches an event to all matching topic listeners and invokes the error handler for any failures
func (m *EventBus) handleEvents(topicName string, event Event, errorHandler func(error)) {
	defer func() {
		if r := recover(); r != nil && m.panicHandler != nil {
			m.panicHandler(r)
		}
	}()

	if event == nil {
		errorHandler(ErrNilPayload)
		return
	}

	matches := make([]matchedTopic, 0)
	m.topics.Range(func(key, value any) bool {
		topicPattern := key.(string)
		if matchTopicPattern(topicPattern, topicName) {
			matches = append(matches, matchedTopic{pattern: topicPattern, t: value.(*topic)})
		}

		return true
	})

	sort.Slice(matches, func(i, j int) bool {
		return compareTopicSpecificity(matches[i].pattern, matches[j].pattern)
	})

	for _, match := range matches {
		topicErrors := m.triggerWithRetry(match.t, event)
		m.handleTopicErrors(event, topicErrors, errorHandler)

		if event.IsAborted() {
			break
		}
	}
}

// handleTopicErrors processes errors from topic listeners through the bus error handler
func (m *EventBus) handleTopicErrors(event Event, topicErrors []error, errorHandler func(error)) {
	for _, err := range topicErrors {
		if m.errorHandler != nil {
			err = m.errorHandler(event, err)
		}

		if err != nil {
			errorHandler(err)
		}
	}
}

// getTopic retrieves a topic by name or returns an error if not found
func (m *EventBus) getTopic(topicName string) (*topic, error) {
	t, ok := m.topics.Load(topicName)
	if !ok {
		return nil, fmt.Errorf("%w: unable to find topic '%s'", ErrTopicNotFound, topicName)
	}

	return t.(*topic), nil
}

// ensureTopic retrieves or creates a topic by name
func (m *EventBus) ensureTopic(topicName string) *topic {
	t, _ := m.topics.LoadOrStore(topicName, newTopic())
	return t.(*topic)
}

// triggerWithRetry invokes all listeners on a topic, collecting any errors
func (m *EventBus) triggerWithRetry(t *topic, event Event) []error {
	type listenerSnapshot struct {
		id       string
		listener Listener
	}

	t.mu.RLock()
	snapshots := make([]listenerSnapshot, 0, len(t.listenerIDs))
	for _, id := range t.listenerIDs {
		listener, ok := t.listeners[id]
		if !ok {
			continue
		}
		snapshots = append(snapshots, listenerSnapshot{id: id, listener: listener})
	}
	t.mu.RUnlock()

	var errs []error

	for _, snap := range snapshots {
		if err := m.runListenerWithRetry(event, snap.id, snap.listener); err != nil {
			errs = append(errs, err)
		}

		if event.IsAborted() {
			break
		}
	}

	return errs
}

// runListenerWithRetry executes a single listener with retry logic and deduplication
func (m *EventBus) runListenerWithRetry(event Event, id string, listener Listener) error {
	retries := m.maxRetries
	if retries <= 0 {
		retries = 1
	}

	eventID := EventID(event)
	if eventID != "" && m.store != nil {
		if deduper, ok := m.store.(handlerResultDeduper); ok {
			alreadySucceeded, err := deduper.HandlerSucceeded(event.Context(), eventID, id)
			if err != nil {
				return err
			}
			if alreadySucceeded {
				return nil
			}
		}
	}

	backOff := m.backOffFactory()
	var lastErr error

	for i := 0; i < retries; i++ {
		ctx := newEventContext(event)
		lastErr = listener(ctx)

		if m.store != nil {
			if saveErr := m.store.SaveHandlerResult(ctx.event, id, lastErr); saveErr != nil {
				log.Warn().Err(saveErr).Str("handler_id", id).Msg("failed to save handler result")
			}
		}

		if lastErr == nil {
			return nil
		}

		if i == retries-1 {
			break
		}

		wait := backOff.NextBackOff()
		if wait == backoff.Stop {
			break
		}

		time.Sleep(wait)
	}

	return lastErr
}

// WaitForIdle blocks until all submitted event handlers have completed
func (m *EventBus) WaitForIdle() {
	m.pool.WaitForIdle()
}

// Close terminates the event pool and releases resources
func (m *EventBus) Close() error {
	if m.closed.Load() {
		return ErrEmitterAlreadyClosed
	}

	m.closed.Store(true)

	if m.queueCancel != nil {
		m.queueCancel()
		m.queueCancel = nil
	}

	m.topics.Range(func(key, _ any) bool {
		m.topics.Delete(key)
		return true
	})

	m.pool.Release()

	deregister(m)

	return nil
}

// startQueueConsumer spawns a background goroutine to consume events from the queue
func (m *EventBus) startQueueConsumer(q eventQueue) {
	ctx, cancel := context.WithCancel(context.Background())
	m.queueCancel = cancel

	consume := func() {
		m.consumeQueue(ctx, q)
	}

	m.pool.Submit(consume)
}

// consumeQueue continuously dequeues and processes events until the context is cancelled
func (m *EventBus) consumeQueue(ctx context.Context, q eventQueue) {
	bo := m.backOffFactory()

	for {
		if ctx.Err() != nil || m.closed.Load() {
			return
		}

		evt, err := q.DequeueEvent(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			wait := bo.NextBackOff()
			if wait == backoff.Stop {
				bo.Reset()
				wait = bo.NextBackOff()
			}

			time.Sleep(wait)

			continue
		}

		bo.Reset()

		topicName, event, err := m.prepareEvent(evt.Topic(), evt)
		if err != nil {
			continue
		}

		m.handleEvents(topicName, event, func(error) {})
	}
}

// trySendErr attempts to send an error to the channel without blocking
func (m *EventBus) trySendErr(ch chan<- error, err error) {
	if err == nil || ch == nil {
		return
	}

	select {
	case ch <- err:
	default:
	}
}

// prepareEvent validates and normalizes the event name and payload before emission
func (m *EventBus) prepareEvent(eventName string, payload any) (string, Event, error) {
	topicName := normalizeTopicName(eventName)

	if event, ok := payload.(Event); ok {
		if event == nil {
			return "", nil, ErrNilPayload
		}

		rawTopic := event.Topic()
		eventTopic := normalizeTopicName(rawTopic)

		switch {
		case topicName == "" && eventTopic == "":
			return "", nil, ErrInvalidTopicName
		case topicName != "" && eventTopic != "" && topicName != eventTopic:
			return "", nil, fmt.Errorf("%w: emit topic %q != event topic %q", ErrEventTopicMismatch, topicName, rawTopic)
		case topicName == "":
			topicName = eventTopic
		}

		if err := validateTopicName(topicName); err != nil {
			return "", nil, err
		}

		if eventTopic == "" || rawTopic != topicName {
			event = cloneEventWithTopic(event, topicName)
		}

		if m.client != nil && event.Client() == nil {
			event.SetClient(m.client)
		}

		m.ensureEventID(event)

		return topicName, event, nil
	}

	if err := validateTopicName(topicName); err != nil {
		return "", nil, err
	}

	event := NewBaseEvent(topicName, payload)
	if m.client != nil {
		event.SetClient(m.client)
	}

	m.ensureEventID(event)

	return topicName, event, nil
}

// cloneEventWithTopic creates a copy of the event with a new topic name
func cloneEventWithTopic(event Event, topic string) Event {
	cloned := NewBaseEvent(topic, event.Payload())
	cloned.SetProperties(cloneProperties(event.Properties()))
	cloned.SetAborted(event.IsAborted())
	cloned.SetContext(event.Context())
	cloned.SetClient(event.Client())
	return cloned
}

// cloneProperties creates a shallow copy of the given Properties map
func cloneProperties(props Properties) Properties {
	if props == nil {
		return NewProperties()
	}

	cloned := NewProperties()
	maps.Copy(cloned, props)

	return cloned
}

// ensureEventID assigns a unique event ID if one is not already present
func (m *EventBus) ensureEventID(event Event) {
	if m == nil || event == nil {
		return
	}

	props := event.Properties()
	if props == nil {
		props = NewProperties()
	}

	if existing, ok := props[PropertyEventID].(string); ok && strings.TrimSpace(existing) != "" {
		event.SetProperties(props)
		return
	}

	props[PropertyEventID] = m.idGenerator()
	event.SetProperties(props)
}

// matchedTopic pairs a topic pattern with its corresponding topic instance
type matchedTopic struct {
	pattern string
	t       *topic
}

// topicSpecificity captures the wildcard and segment counts used for pattern ordering
type topicSpecificity struct {
	multiWildcards  int
	singleWildcards int
	segments        int
	length          int
}

// topicSpecificityKey computes the specificity metrics for a topic pattern
func topicSpecificityKey(pattern string) topicSpecificity {
	parts := strings.Split(pattern, ".")
	key := topicSpecificity{
		segments: len(parts),
		length:   len(pattern),
	}

	for _, part := range parts {
		switch part {
		case multiWildcard:
			key.multiWildcards++
		case singleWildcard:
			key.singleWildcards++
		}
	}

	return key
}

// compareTopicSpecificity returns true if pattern a should be ordered before pattern b
func compareTopicSpecificity(a, b string) bool {
	keyA := topicSpecificityKey(a)
	keyB := topicSpecificityKey(b)

	if keyA.multiWildcards != keyB.multiWildcards {
		return keyA.multiWildcards < keyB.multiWildcards
	}
	if keyA.singleWildcards != keyB.singleWildcards {
		return keyA.singleWildcards < keyB.singleWildcards
	}
	if keyA.segments != keyB.segments {
		return keyA.segments > keyB.segments
	}
	if keyA.length != keyB.length {
		return keyA.length > keyB.length
	}

	return a < b
}
