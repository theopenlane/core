package soiree

import (
	"errors"
	"sync"
	"testing"
)

const errorMessage = "On() failed with error: %v"

// TestWithErrorHandler tests that the custom error handler is called on error
func TestWithErrorHandler(t *testing.T) {
	// Define a variable to determine if the custom error handler was called
	var handlerCalled bool

	// Define a custom error to be returned by a listener
	customError := errors.New("custom error") // nolint: goerr113

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
	listener := func(e Event) error {
		return customError
	}

	_, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf(errorMessage, err)
	}

	// Emit the event synchronously to trigger the error
	soiree.EmitSync("testTopic", NewBaseEvent("testTopic", "testPayload"))

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
	customError := errors.New("custom error") // nolint: goerr113

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
	listener := func(e Event) error {
		return customError
	}

	// Subscribe the listener to a topic
	_, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf("Error occurred: %v", err)
	}

	// Emit the event asynchronously to trigger the error
	errChan := soiree.Emit("testTopic", NewBaseEvent("testTopic", "testPayload"))

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
	customPanicHandler := func(p interface{}) {
		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	// Create a new EventPool with the custom panic handler
	soiree := NewEventPool(WithPanicHandler(customPanicHandler))

	// Define a listener that panics
	listener := func(e Event) error {
		panic("test panic")
	}

	// Subscribe the listener to a topic
	_, err := soiree.On("testTopic", listener)
	if err != nil {
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
	soiree.EmitSync("testTopic", "testPayload")

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
	customPanicHandler := func(p interface{}) {
		panicHandlerMutex.Lock()
		defer panicHandlerMutex.Unlock()

		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	// Create a new EventPool with the custom panic handler.
	soiree := NewEventPool(WithPanicHandler(customPanicHandler))

	// Define a listener that panics
	listener := func(e Event) error {
		panic("test panic")
	}

	// Subscribe the listener to a topic
	_, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf(errorMessage, err)
	}

	// Emit the event asynchronously to trigger the panic
	errChan := soiree.Emit("testTopic", "testPayload")

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
	listener := func(e Event) error {
		return nil
	}

	// Subscribe the listener to a topic and capture the returned ID
	returnedID, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf(errorMessage, err)
	}

	// Check if the returned ID matches the custom ID
	if returnedID != customID {
		t.Fatalf("Expected ID to be '%s', but got '%s'", customID, returnedID)
	}
}
