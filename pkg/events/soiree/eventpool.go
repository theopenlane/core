package soiree

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// EventPool struct is controlling subscribing and unsubscribing listeners to topics, and emitting events to all subscribers
type EventPool struct {
	// topics with concurrent access support
	topics sync.Map
	// errorHandler will handle errors that occur during event handling
	errorHandler func(Event, error) error
	// idGenerator generates unique IDs for listeners
	idGenerator func() string
	// panicHandler will handle panics that occur during event handling
	panicHandler PanicHandler
	// pool manages concurrent execution of event handlers
	pool Pool
	// Indicates whether the soiree is closed
	closed atomic.Value
	// Size of the buffer for the error channel in Emit
	errChanBufferSize int
	// client is the client for the event pool
	client any
	// store is used to persist events and results
	store EventStore
	// queueCancel cancels the queue consumer when present
	queueCancel context.CancelFunc
	// maxRetries specifies how many times a listener should be retried on error
	maxRetries int
	// backOffFactory creates a backoff policy for retries
	backOffFactory func() backoff.BackOff
}

// NewEventPool initializes a new EventPool with optional configuration options
func NewEventPool(opts ...EventPoolOption) *EventPool {
	m := &EventPool{
		topics:            sync.Map{},
		errorHandler:      DefaultErrorHandler,
		idGenerator:       DefaultIDGenerator,
		panicHandler:      DefaultPanicHandler,
		errChanBufferSize: 10, //nolint:mnd
		maxRetries:        1,
		backOffFactory: func() backoff.BackOff {
			return backoff.NewExponentialBackOff()
		},
	}

	m.closed.Store(false)

	// Apply each provided option to the soiree to configure it
	for _, opt := range opts {
		opt(m)
	}

	register(m)

	if q, ok := m.store.(EventQueue); ok {
		m.startQueueConsumer(q)
	}

	return m
}

// SetClient sets a client that can be used as a part of the event pool
func (m *EventPool) SetClient(client any) {
	m.client = client
}

// GetClient fetches the set client on the event pool
func (m *EventPool) GetClient() any {
	return m.client
}

// On subscribes a listener to a topic with the given name; returns a unique listener ID
func (m *EventPool) On(topicName string, listener Listener, opts ...ListenerOption) (string, error) {
	if listener == nil {
		return "", ErrNilListener
	}

	if !isValidTopicName(topicName) {
		return "", ErrInvalidTopicName
	}

	topic := m.EnsureTopic(topicName)
	listenerID := m.idGenerator()
	topic.AddListener(listenerID, listener, opts...)

	return listenerID, nil
}

// Off unsubscribes a listener from a topic using the listener's unique ID
func (m *EventPool) Off(topicName string, listenerID string) error {
	topic, err := m.GetTopic(topicName)
	if err != nil {
		return err
	}

	return topic.RemoveListener(listenerID)
}

// Emit asynchronously dispatches an event to all the subscribers of the event's topic
// It returns a channel that will receive any errors encountered during event handling
func (m *EventPool) Emit(eventName string, payload any) <-chan error {
	errChan := make(chan error, m.errChanBufferSize)

	// Before starting new goroutine, check if Soiree is closed
	if m.closed.Load().(bool) {
		errChan <- ErrEmitterClosed

		close(errChan)

		return errChan
	}

	event, ok := payload.(Event)
	if !ok {
		event = NewBaseEvent(eventName, payload)
	}

	if m.store != nil {
		if err := m.store.SaveEvent(event); err != nil {
			errChan <- err
		}
	}

	if _, ok := m.store.(EventQueue); ok {
		close(errChan)

		return errChan
	}

	if m.pool != nil {
		m.pool.Submit(func() {
			defer close(errChan)

			m.handleEvents(eventName, event, func(err error) {
				errChan <- err
			})
		})
	} else {
		go func() {
			defer close(errChan)

			m.handleEvents(eventName, event, func(err error) {
				errChan <- err
			})
		}()
	}

	return errChan
}

// EmitSync dispatches an event synchronously to all subscribers of the event's topic; his method will block until all notifications are completed
func (m *EventPool) EmitSync(eventName string, payload any) []error {
	var errs []error

	if m.closed.Load().(bool) {
		return []error{ErrEmitterClosed}
	}

	event, ok := payload.(Event)
	if !ok {
		event = NewBaseEvent(eventName, payload)
	}

	if m.store != nil {
		if err := m.store.SaveEvent(event); err != nil {
			errs = append(errs, err)

			return errs
		}
	}

	if _, ok := m.store.(EventQueue); ok {
		return nil
	}

	m.handleEvents(eventName, event, func(err error) {
		errs = append(errs, err)
	})

	return errs
}

// handleEvents is an internal method that processes an event and notifies all registered listeners
func (m *EventPool) handleEvents(topicName string, payload any, errorHandler func(error)) {
	defer func() {
		if r := recover(); r != nil && m.panicHandler != nil {
			m.panicHandler(r)
		}
	}()

	event, ok := payload.(Event)
	if !ok {
		event = NewBaseEvent(topicName, payload)
	}

	m.topics.Range(func(key, value any) bool {
		topicPattern := key.(string)
		if matchTopicPattern(topicPattern, topicName) {
			topic := value.(*Topic)
			topicErrors := m.triggerWithRetry(topic, event)
			m.handleTopicErrors(event, topicErrors, errorHandler)
		}

		return true
	})
}

// handleTopicErrors handles the errors returned by a topic's Trigger method
func (m *EventPool) handleTopicErrors(event Event, topicErrors []error, errorHandler func(error)) {
	for _, err := range topicErrors {
		if m.errorHandler != nil {
			err = m.errorHandler(event, err)
		}

		if err != nil {
			errorHandler(err)
		}
	}
}

// GetTopic retrieves a topic by its name. If the topic does not exist, it returns an error
func (m *EventPool) GetTopic(topicName string) (*Topic, error) {
	topic, ok := m.topics.Load(topicName)
	if !ok {
		return nil, fmt.Errorf("%w: unable to find topic '%s'", ErrTopicNotFound, topicName)
	}

	return topic.(*Topic), nil
}

// EnsureTopic retrieves or creates a new topic by its name
func (m *EventPool) EnsureTopic(topicName string) *Topic {
	topic, _ := m.topics.LoadOrStore(topicName, NewTopic())

	return topic.(*Topic)
}

// InterestedIn reports whether any listeners are registered for the given topic.
func (m *EventPool) InterestedIn(topicName string) bool {
	topic, err := m.GetTopic(topicName)
	if err != nil {
		return false
	}

	return topic.HasListeners()
}

// SetErrorHandler sets the error handler for the event pool
func (m *EventPool) SetErrorHandler(handler func(Event, error) error) {
	if handler != nil {
		m.errorHandler = handler
	}
}

// SetIDGenerator sets the ID generator for the event pool
func (m *EventPool) SetIDGenerator(generator func() string) {
	if generator != nil {
		m.idGenerator = generator
	}
}

// SetPool sets the pool for the event pool
func (m *EventPool) SetPool(pool Pool) {
	m.pool = pool
}

// SetPanicHandler sets the panic handler for the event pool
func (m *EventPool) SetPanicHandler(panicHandler PanicHandler) {
	if panicHandler != nil {
		m.panicHandler = panicHandler
	}
}

// SetErrChanBufferSize sets the buffer size for the error channel for the event pool
func (m *EventPool) SetErrChanBufferSize(size int) {
	m.errChanBufferSize = size
}

// SetEventStore sets the event persistence store for the pool
func (m *EventPool) SetEventStore(store EventStore) {
	if m.queueCancel != nil {
		m.queueCancel()
		m.queueCancel = nil
	}

	m.store = store

	if q, ok := m.store.(EventQueue); ok && !m.closed.Load().(bool) {
		m.startQueueConsumer(q)
	}
}

// SetRetry configures the retry attempts and backoff factory
func (m *EventPool) SetRetry(retries int, factory func() backoff.BackOff) {
	if retries > 0 {
		m.maxRetries = retries
	}

	if factory != nil {
		m.backOffFactory = factory
	}
}

// RegisterListeners registers all provided listener bindings on the pool and returns the generated listener IDs
func (m *EventPool) RegisterListeners(bindings ...ListenerBinding) ([]string, error) {
	ids := make([]string, 0, len(bindings))

	for _, binding := range bindings {
		id, err := binding.registerWith(m)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// triggerWithRetry executes topic listeners with retry and persistence
func (m *EventPool) triggerWithRetry(topic *Topic, event Event) []error {
	topic.mu.RLock()
	defer topic.mu.RUnlock()

	var errs []error

	for _, id := range topic.sortedListenerIDs {
		item, ok := topic.listeners[id]
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

// runListenerWithRetry executes a listener with retry logic and persistence.
func (m *EventPool) runListenerWithRetry(event Event, id string, item *listenerItem) error {
	retries := m.maxRetries
	if retries <= 0 {
		retries = 1
	}

	backOff := m.backOffFactory()

	var lastErr error

	for i := 0; i < retries; i++ {
		ctx := newEventContext(event)

		lastErr = item.call(ctx)
		if m.store != nil {
			_ = m.store.SaveHandlerResult(ctx.Event(), id, lastErr)
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

// Close terminates the soiree, ensuring all pending events are processed; it performs cleanup and releases resources
func (m *EventPool) Close() error {
	if m.closed.Load().(bool) {
		return ErrEmitterAlreadyClosed
	}

	m.closed.Store(true)

	// tidy it up
	m.topics.Range(func(key, _ any) bool {
		m.topics.Delete(key)
		return true
	})

	if m.pool != nil {
		m.pool.Release()
	}

	if m.queueCancel != nil {
		m.queueCancel()
	}

	deregister(m)

	return nil
}

// startQueueConsumer starts a goroutine that continuously consumes events from the queue
// It will stop when the context is cancelled or the event pool is closed
// If a pool is set, it will use the pool to submit the consumer function
func (m *EventPool) startQueueConsumer(q EventQueue) {
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

// consumeQueue continuously dequeues events from the queue and processes them
// It will stop when the context is cancelled or the event pool is closed
// This method is designed to be run in a separate goroutine
func (m *EventPool) consumeQueue(ctx context.Context, q EventQueue) {
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

		m.handleEvents(evt.Topic(), evt, func(error) {})
	}
}
