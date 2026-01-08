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
		WithWrap(func(e Event) Event { return e }),
		WithUnwrap(func(e Event) (Event, error) { return e, nil }),
	)
}

func subscribeDefaultTopic(bus *EventBus, listener TypedListener[Event]) (string, error) {
	return BindListener(defaultTopic(), listener).Register(bus)
}

func emitDefaultTopicAndCollect(bus *EventBus, payload any) []error {
	topic := defaultTopic()
	return collectEmitErrors(bus.Emit(topic.Name(), NewBaseEvent(topic.Name(), payload)))
}

func emitDefaultTopicAsync(bus *EventBus, payload any) <-chan error {
	topic := defaultTopic()
	return bus.Emit(topic.Name(), NewBaseEvent(topic.Name(), payload))
}

const errorMessage = "On() failed with error: %v"

// TestErrorHandler tests that the custom error handler is called on error
func TestErrorHandler(t *testing.T) {
	var handlerCalled bool

	customError := errors.New("custom error") //nolint:err113

	customErrorHandler := func(event Event, err error) error {
		if errors.Is(err, customError) {
			handlerCalled = true

			t.Logf("Custom error handler called with event: %s and error: %s", event.Topic(), err.Error())
		}

		return nil
	}

	soiree := New(ErrorHandler(customErrorHandler))

	listener := func(_ *EventContext, e Event) error {
		return customError
	}

	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf(errorMessage, err)
	}

	emitDefaultTopicAndCollect(soiree, "testPayload")

	if !handlerCalled {
		t.Fatalf("Custom error handler was not called on listener error")
	}
}

func TestErrorHandlerAsync(t *testing.T) {
	var handlerCalled bool
	var handlerMutex sync.Mutex

	customError := errors.New("custom error") //nolint:err113

	customErrorHandler := func(event Event, err error) error {
		handlerMutex.Lock()

		defer handlerMutex.Unlock()

		if errors.Is(err, customError) {
			handlerCalled = true
		}

		return nil
	}

	soiree := New(ErrorHandler(customErrorHandler))

	listener := func(_ *EventContext, e Event) error {
		return customError
	}

	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf("Error occurred: %v", err)
	}

	errChan := emitDefaultTopicAsync(soiree, "testPayload")

	for err := range errChan {
		if err != nil {
			t.Errorf("Expected nil error due to custom handler, got: %v", err)
		}
	}

	handlerMutex.Lock()
	wasHandlerCalled := handlerCalled
	handlerMutex.Unlock()

	if !wasHandlerCalled {
		t.Fatalf("Custom error handler was not called on listener error")
	}
}

func TestPanicHandler(t *testing.T) {
	var panicHandlerInvoked bool

	customPanicHandler := func(p any) {
		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	soiree := New(Panics(customPanicHandler))

	listener := func(_ *EventContext, e Event) error {
		panic("test panic")
	}

	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf("errorMessage: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	emitDefaultTopicAndCollect(soiree, "testPayload")

	if !panicHandlerInvoked {
		t.Fatalf("Custom panic handler was not called on listener panic")
	}
}

func TestPanicHandlerAsync(t *testing.T) {
	var panicHandlerInvoked bool

	var panicHandlerMutex sync.Mutex

	customPanicHandler := func(p any) {
		panicHandlerMutex.Lock()
		defer panicHandlerMutex.Unlock()

		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	soiree := New(Panics(customPanicHandler))

	listener := func(_ *EventContext, e Event) error {
		panic("test panic")
	}

	if _, err := subscribeDefaultTopic(soiree, listener); err != nil {
		t.Fatalf(errorMessage, err)
	}

	errChan := emitDefaultTopicAsync(soiree, "testPayload")

	for range errChan {
	}

	panicHandlerMutex.Lock()
	wasPanicHandlerInvoked := panicHandlerInvoked
	panicHandlerMutex.Unlock()

	if !wasPanicHandlerInvoked {
		t.Fatalf("Custom panic handler was not called on listener panic")
	}
}

func TestIDGenerator(t *testing.T) {
	customID := "customID"

	customIDGenerator := func() string {
		return customID
	}

	soiree := New(IDGenerator(customIDGenerator))

	listener := func(_ *EventContext, e Event) error {
		return nil
	}

	returnedID, err := subscribeDefaultTopic(soiree, listener)
	if err != nil {
		t.Fatalf(errorMessage, err)
	}

	if returnedID != customID {
		t.Fatalf("Expected ID to be '%s', but got '%s'", customID, returnedID)
	}
}

func TestAllEventBusOptions(t *testing.T) {
	dummyRedis := newTestRedis(t)
	dummyClient := struct{}{}

	bus1 := New(Workers(10))
	if bus1 == nil {
		t.Fatal("Workers option failed")
	}

	bus2 := New(WithRedisStore(dummyRedis))
	if bus2 == nil {
		t.Fatal("WithRedisStore option failed")
	}

	bus3 := New(Retry(2, func() backoff.BackOff { return backoff.NewConstantBackOff(1) }))
	if bus3 == nil {
		t.Fatal("Retry option failed")
	}

	bus4 := New(ErrChanBufferSize(5))
	if bus4 == nil {
		t.Fatal("ErrChanBufferSize option failed")
	}

	bus5 := New(Client(dummyClient))
	if bus5 == nil {
		t.Fatal("Client option failed")
	}
}
