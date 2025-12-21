package soiree

import (
	"errors"
	"sync"
	"testing"

	"github.com/cenkalti/backoff/v5"
)

func defaultTopic() TypedTopic[Event] {
	return NewTypedTopic(
		"testTopic",
		func(e Event) Event { return e },
		func(e Event) (Event, error) { return e, nil },
	)
}

func subscribeDefaultTopic(pool *EventPool, listener TypedListener[Event], opts ...ListenerOption) (string, error) {
	return BindListener(defaultTopic(), listener, opts...).Register(pool)
}

func emitDefaultTopicSync(pool *EventPool, payload any) []error {
	topic := defaultTopic()
	return pool.EmitSync(topic.Name(), NewBaseEvent(topic.Name(), payload))
}

func emitDefaultTopicAsync(pool *EventPool, payload any) <-chan error {
	topic := defaultTopic()
	return pool.Emit(topic.Name(), NewBaseEvent(topic.Name(), payload))
}

const errorMessage = "On() failed with error: %v"

// TestWithErrorHandler tests that the custom error handler is called on error
func TestWithErrorHandler(t *testing.T) {
	// Define a variable to determine if the custom error handler was called
	var handlerCalled bool

	// Define a custom error to be returned by a listener
	customError := errors.New("custom error") //nolint:err113

	// Define a custom error handler that sets handlerCalled to true
	customErrorHandler := func(event Event, err error) error {
		if errors.Is(err, customError) {
			handlerCalled = true

			t.Logf("Custom error handler called with event: %s and error: %s", event.Topic(), err.Error())
		}

		return nil // Returning nil to indicate the error is handled
	}

	// Create a new EventPool with the custom error handler
	soiree := NewEventPool(WithErrorHandler(customErrorHandler))

	// Define a listener that returns the custom error
	listener := func(_ *EventContext, e Event) error {
		return customError
	}

	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf(errorMessage, err)
	}

	// Emit the event synchronously to trigger the error
	emitDefaultTopicSync(soiree, "testPayload")

	// Check if the custom error handler was called
	if !handlerCalled {
		t.Fatalf("Custom error handler was not called on listener error")
	}
}

func TestWithErrorHandlerAsync(t *testing.T) {
	// Define a variable to determine if the custom error handler was called
	var handlerCalled bool
	// To safely update handlerCalled from different goroutines
	var handlerMutex sync.Mutex

	// Define a custom error to be returned by a listener
	customError := errors.New("custom error") //nolint:err113

	// Define a custom error handler that sets handlerCalled to true
	customErrorHandler := func(event Event, err error) error {
		handlerMutex.Lock()

		defer handlerMutex.Unlock()

		if errors.Is(err, customError) {
			handlerCalled = true
		}

		return nil // Assume the error is handled and return nil
	}

	// Create a new EventPool with the custom error handler
	soiree := NewEventPool(WithErrorHandler(customErrorHandler))

	// Define a listener that returns the custom error
	listener := func(_ *EventContext, e Event) error {
		return customError
	}

	// Subscribe the listener to a topic
	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf("Error occurred: %v", err)
	}

	// Emit the event asynchronously to trigger the error
	errChan := emitDefaultTopicAsync(soiree, "testPayload")

	// Wait for all errors to be processed
	for err := range errChan {
		if err != nil {
			t.Errorf("Expected nil error due to custom handler, got: %v", err)
		}
	}

	// Check if the custom error handler was called
	handlerMutex.Lock()
	wasHandlerCalled := handlerCalled
	handlerMutex.Unlock()

	if !wasHandlerCalled {
		t.Fatalf("Custom error handler was not called on listener error")
	}
}

func TestWithPanicHandlerSync(t *testing.T) {
	// Flag to indicate panic handler invocation
	var panicHandlerInvoked bool

	// Define a custom panic handler
	customPanicHandler := func(p any) {
		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	// Create a new EventPool with the custom panic handler
	soiree := NewEventPool(WithPanicHandler(customPanicHandler))

	// Define a listener that panics
	listener := func(_ *EventContext, e Event) error {
		panic("test panic")
	}

	// Subscribe the listener to a topic
	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf("errorMessage: %v", err)
	}

	// Recover from panic to prevent test failure
	defer func() {
		if r := recover(); r != nil {
			// This is expected
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	// Emit the event synchronously to trigger the panic
	emitDefaultTopicSync(soiree, "testPayload")

	// Verify that the custom panic handler was invoked
	if !panicHandlerInvoked {
		t.Fatalf("Custom panic handler was not called on listener panic")
	}
}

func TestWithPanicHandlerAsync(t *testing.T) {
	// Flag to indicate panic handler invocation
	var panicHandlerInvoked bool

	var panicHandlerMutex sync.Mutex // To safely update panicHandlerInvoked from different goroutines

	// Define a custom panic handler
	customPanicHandler := func(p any) {
		panicHandlerMutex.Lock()
		defer panicHandlerMutex.Unlock()

		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	// Create a new EventPool with the custom panic handler.
	soiree := NewEventPool(WithPanicHandler(customPanicHandler))

	// Define a listener that panics
	listener := func(_ *EventContext, e Event) error {
		panic("test panic")
	}

	// Subscribe the listener to a topic
	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf(errorMessage, err)
	}

	// Emit the event asynchronously to trigger the panic
	errChan := emitDefaultTopicAsync(soiree, "testPayload")

	// Wait for all events to be processed (which includes recovering from panic)
	for range errChan {
		// Normally, you'd check for errors here, but in this case, we expect a panic, not an error
	}

	// Verify that the custom panic handler was invoked
	panicHandlerMutex.Lock()
	wasPanicHandlerInvoked := panicHandlerInvoked
	panicHandlerMutex.Unlock()

	if !wasPanicHandlerInvoked {
		t.Fatalf("Custom panic handler was not called on listener panic")
	}
}

func TestWithIDGenerator(t *testing.T) {
	// Custom ID to be returned by the custom ID generator
	customID := "customID"

	// Define a custom ID generator that returns the custom ID
	customIDGenerator := func() string {
		return customID
	}

	// Create a new EventPool with the custom ID generator
	soiree := NewEventPool(WithIDGenerator(customIDGenerator))

	// Define a no-op listener
	listener := func(_ *EventContext, e Event) error {
		return nil
	}

	// Subscribe the listener to a topic and capture the returned ID
	returnedID, err := subscribeDefaultTopic(soiree, listener)
	if err != nil {
		t.Fatalf(errorMessage, err)
	}

	// Check if the returned ID matches the custom ID
	if returnedID != customID {
		t.Fatalf("Expected ID to be '%s', but got '%s'", customID, returnedID)
	}
}

func TestAllEventPoolOptions(t *testing.T) {
	// Dummy implementations for required interfaces
	dummyPool := NewPondPool(WithMaxWorkers(1))
	dummyRedis := newTestRedis(t)
	dummyClient := struct{}{}

	// Test WithPool
	soiree := NewEventPool(WithPool(dummyPool))
	if soiree == nil {
		t.Fatal("WithPool option failed")
	}

	// Test WithRedisStore
	soiree2 := NewEventPool(WithRedisStore(dummyRedis))
	if soiree2 == nil {
		t.Fatal("WithRedisStore option failed")
	}

	// Test WithRetry
	soiree3 := NewEventPool(WithRetry(2, func() backoff.BackOff { return backoff.NewConstantBackOff(1) }))
	if soiree3 == nil {
		t.Fatal("WithRetry option failed")
	}

	// Test WithErrChanBufferSize
	soiree4 := NewEventPool(WithErrChanBufferSize(5))
	if soiree4 == nil {
		t.Fatal("WithErrChanBufferSize option failed")
	}

	// Test WithClient
	soiree5 := NewEventPool(WithClient(dummyClient))
	if soiree5 == nil {
		t.Fatal("WithClient option failed")
	}
}
