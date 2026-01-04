package soiree

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// EventBus manages subscribing and unsubscribing listeners to topics and emitting events to subscribers
type EventBus struct {
	topics            sync.Map
	errorHandler      func(Event, error) error
	idGenerator       func() string
	panicHandler      PanicHandler
	pool              *PondPool
	closed            atomic.Value
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

	m.closed.Store(false)

	for _, opt := range opts {
		opt(m)
	}

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

// GetClient returns the client set on the event bus (alias for Client)
func (m *EventBus) GetClient() any {
	return m.client
}

// InterestedIn checks if the event bus has any listeners registered for the given topic
func (m *EventBus) InterestedIn(topicName string) bool {
	t, err := m.getTopic(topicName)
	if err != nil {
		return false
	}

	return t.hasListeners()
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

	if !isValidTopicName(topicName) {
		return "", ErrInvalidTopicName
	}

	t := m.ensureTopic(topicName)
	listenerID := m.idGenerator()
	t.addListener(listenerID, listener)

	return listenerID, nil
}

// Off unsubscribes a listener from a topic using the listener's unique ID
func (m *EventBus) Off(topicName string, listenerID string) error {
	t, err := m.getTopic(topicName)
	if err != nil {
		return err
	}

	return t.removeListener(listenerID)
}

// Emit asynchronously dispatches an event to all subscribers of the event's topic
func (m *EventBus) Emit(eventName string, payload any) <-chan error {
	errChan := make(chan error, m.errChanBufferSize)

	if m.closed.Load().(bool) {
		m.trySendErr(errChan, ErrEmitterClosed)
		close(errChan)
		return errChan
	}

	topicName, event, prepErr := m.prepareEmit(eventName, payload)
	if prepErr != nil {
		m.trySendErr(errChan, prepErr)
		close(errChan)

		return errChan
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

	if m.pool != nil {
		m.pool.Submit(func() {
			defer close(errChan)
			m.handleEvents(topicName, event, func(err error) {
				m.trySendErr(errChan, err)
			})
		})
	} else {
		go func() {
			defer close(errChan)
			m.handleEvents(topicName, event, func(err error) {
				m.trySendErr(errChan, err)
			})
		}()
	}

	return errChan
}

// EmitSync dispatches an event synchronously to all subscribers of the event's topic
func (m *EventBus) EmitSync(eventName string, payload any) []error {
	var errs []error

	if m.closed.Load().(bool) {
		return []error{ErrEmitterClosed}
	}

	topicName, event, prepErr := m.prepareEmit(eventName, payload)
	if prepErr != nil {
		return []error{prepErr}
	}

	if m.store != nil {
		if err := m.store.SaveEvent(event); err != nil {
			errs = append(errs, err)

			return errs
		}
	}

	if _, ok := m.store.(eventQueue); ok {
		return nil
	}

	m.handleEvents(topicName, event, func(err error) {
		errs = append(errs, err)
	})

	return errs
}

func (m *EventBus) handleEvents(topicName string, payload any, errorHandler func(error)) {
	defer func() {
		if r := recover(); r != nil && m.panicHandler != nil {
			m.panicHandler(r)
		}
	}()

	event, ok := payload.(Event)
	if !ok {
		event = NewBaseEvent(topicName, payload)
	} else if strings.TrimSpace(event.Topic()) != "" && event.Topic() != topicName {
		errorHandler(fmt.Errorf("%w: emit topic %q != event topic %q", ErrEventTopicMismatch, topicName, event.Topic()))

		return
	}

	if m.client != nil && event.Client() == nil {
		event.SetClient(m.client)
	}

	m.ensureEventID(event)

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

func (m *EventBus) getTopic(topicName string) (*topic, error) {
	t, ok := m.topics.Load(topicName)
	if !ok {
		return nil, fmt.Errorf("%w: unable to find topic '%s'", ErrTopicNotFound, topicName)
	}

	return t.(*topic), nil
}

func (m *EventBus) ensureTopic(topicName string) *topic {
	t, _ := m.topics.LoadOrStore(topicName, newTopic())
	return t.(*topic)
}

func (m *EventBus) triggerWithRetry(t *topic, event Event) []error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var errs []error

	for _, id := range t.listenerIDs {
		item, ok := t.listeners[id]
		if !ok {
			continue
		}

		if err := m.runListenerWithRetry(event, id, item); err != nil {
			errs = append(errs, err)
		}

		if event.IsAborted() {
			break
		}
	}

	return errs
}

func (m *EventBus) runListenerWithRetry(event Event, id string, item *listenerItem) error {
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
		lastErr = item.call(ctx)

		if m.store != nil {
			_ = m.store.SaveHandlerResult(ctx.event, id, lastErr)
		}

		if lastErr == nil {
			return nil
		}

		wait := backOff.NextBackOff()
		if wait == backoff.Stop {
			break
		}

		time.Sleep(wait)
	}

	return lastErr
}

// Close terminates the event pool and releases resources
func (m *EventBus) Close() error {
	if m.closed.Load().(bool) {
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

	if m.pool != nil {
		m.pool.Release()
	}

	deregister(m)

	return nil
}

func (m *EventBus) startQueueConsumer(q eventQueue) {
	ctx, cancel := context.WithCancel(context.Background())
	m.queueCancel = cancel

	consume := func() {
		m.consumeQueue(ctx, q)
	}

	if m.pool != nil {
		m.pool.Submit(consume)
	} else {
		go consume()
	}
}

func (m *EventBus) consumeQueue(ctx context.Context, q eventQueue) {
	for {
		if ctx.Err() != nil || m.closed.Load().(bool) {
			return
		}

		evt, err := q.DequeueEvent(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(100 * time.Millisecond) //nolint:mnd
			continue
		}

		if m.client != nil && evt.Client() == nil {
			evt.SetClient(m.client)
		}

		m.ensureEventID(evt)
		m.handleEvents(evt.Topic(), evt, func(error) {})
	}
}

func (m *EventBus) trySendErr(ch chan<- error, err error) {
	if err == nil || ch == nil {
		return
	}

	select {
	case ch <- err:
	default:
	}
}

func (m *EventBus) prepareEmit(eventName string, payload any) (string, Event, error) {
	topicName := strings.TrimSpace(eventName)

	if event, ok := payload.(Event); ok {
		eventTopic := strings.TrimSpace(event.Topic())

		switch {
		case topicName == "" && eventTopic == "":
			return "", nil, ErrInvalidTopicName
		case topicName != "" && eventTopic != "" && topicName != eventTopic:
			return "", nil, fmt.Errorf("%w: emit topic %q != event topic %q", ErrEventTopicMismatch, topicName, eventTopic)
		case eventTopic == "" && topicName != "":
			event = cloneEventWithTopic(event, topicName)
		case topicName == "" && eventTopic != "":
			topicName = eventTopic
		default:
			topicName = eventTopic
		}

		if topicName == "" || !isValidTopicName(topicName) {
			return "", nil, ErrInvalidTopicName
		}

		if m.client != nil && event.Client() == nil {
			event.SetClient(m.client)
		}

		m.ensureEventID(event)

		return topicName, event, nil
	}

	if topicName == "" || !isValidTopicName(topicName) {
		return "", nil, ErrInvalidTopicName
	}

	event := NewBaseEvent(topicName, payload)
	if m.client != nil {
		event.SetClient(m.client)
	}

	m.ensureEventID(event)

	return topicName, event, nil
}

func cloneEventWithTopic(event Event, topic string) Event {
	cloned := NewBaseEvent(topic, event.Payload())
	cloned.SetProperties(event.Properties())
	cloned.SetAborted(event.IsAborted())
	cloned.SetContext(event.Context())
	cloned.SetClient(event.Client())
	return cloned
}

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

type matchedTopic struct {
	pattern string
	t       *topic
}

type topicSpecificity struct {
	multiWildcards  int
	singleWildcards int
	segments        int
	length          int
}

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
