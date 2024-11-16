package soiree

import (
	"fmt"
	"sync"
	"sync/atomic"
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
	client interface{}
}

// NewEventPool initializes a new EventPool with optional configuration options
func NewEventPool(opts ...EventPoolOption) *EventPool {
	m := &EventPool{
		topics:            sync.Map{},
		errorHandler:      DefaultErrorHandler,
		idGenerator:       DefaultIDGenerator,
		panicHandler:      DefaultPanicHandler,
		errChanBufferSize: 10, //nolint:mnd
	}

	m.closed.Store(false)

	// Apply each provided option to the soiree to configure it
	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *EventPool) SetClient(client interface{}) {
	m.client = client
}

func (m *EventPool) GetClient() interface{} {
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
func (m *EventPool) Emit(eventName string, payload interface{}) <-chan error {
	errChan := make(chan error, m.errChanBufferSize)

	// Before starting new goroutine, check if Soiree is closed
	if m.closed.Load().(bool) {
		errChan <- ErrEmitterClosed
		close(errChan)

		return errChan
	}

	if m.pool != nil {
		m.pool.Submit(func() {
			defer close(errChan)
			m.handleEvents(eventName, payload, func(err error) {
				errChan <- err
			})
		})
	} else {
		go func() {
			defer close(errChan)
			m.handleEvents(eventName, payload, func(err error) {
				errChan <- err
			})
		}()
	}

	return errChan
}

// EmitSync dispatches an event synchronously to all subscribers of the event's topic; his method will block until all notifications are completed
func (m *EventPool) EmitSync(eventName string, payload interface{}) []error {
	if m.closed.Load().(bool) {
		return []error{ErrEmitterClosed}
	}

	var errs []error

	m.handleEvents(eventName, payload, func(err error) {
		errs = append(errs, err)
	})

	return errs
}

// handleEvents is an internal method that processes an event and notifies all registered listeners
func (m *EventPool) handleEvents(topicName string, payload interface{}, errorHandler func(error)) {
	defer func() {
		if r := recover(); r != nil && m.panicHandler != nil {
			m.panicHandler(r)
		}
	}()

	event, ok := payload.(Event)
	if !ok {
		event = NewBaseEvent(topicName, payload)
	}

	m.topics.Range(func(key, value interface{}) bool {
		topicPattern := key.(string)
		if matchTopicPattern(topicPattern, topicName) {
			topic := value.(*Topic)
			topicErrors := topic.Trigger(event)
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

// Close terminates the soiree, ensuring all pending events are processed; it performs cleanup and releases resources
func (m *EventPool) Close() error {
	if m.closed.Load().(bool) {
		return ErrEmitterAlreadyClosed
	}

	m.closed.Store(true)

	// tidy it up
	m.topics.Range(func(key, value interface{}) bool {
		m.topics.Delete(key)
		return true
	})

	if m.pool != nil {
		m.pool.Release()
	}

	return nil
}
