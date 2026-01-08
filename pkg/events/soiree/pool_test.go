package soiree

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func mustSubscribe(pool *EventBus, topic string, listener TypedListener[Event]) string {
	id, err := BindListener(typedEventTopic(topic), listener).Register(pool)
	if err != nil {
		panic(err)
	}
	return id
}

func emitPayload(pool *EventBus, topic string, payload any) <-chan error {
	return pool.Emit(topic, Event(NewBaseEvent(topic, payload)))
}

func TestEmitEventWithPool(t *testing.T) {
	pool := New(Workers(10))

	var processedEvents int32

	listenerID := mustSubscribe(pool, "testEvent", func(_ *EventContext, event Event) error {
		atomic.AddInt32(&processedEvents, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	errChan := emitPayload(pool, "testEvent", nil)

	var errs []error
	go func() {
		for err := range errChan {
			if err != nil {
				errs = append(errs, err)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)

	if len(errs) > 0 {
		t.Fatalf("Listener reported errors: %v", errs)
	}

	if err := pool.Off("testEvent", listenerID); err != nil {
		t.Errorf("Failed to unregister listener: %v", err)
	}

	if atomic.LoadInt32(&processedEvents) != 1 {
		t.Fatalf("Expected 1 event to be processed, but got %d", processedEvents)
	}
}

func TestEmitMultipleEventsWithPool(t *testing.T) {
	pool := New(Workers(10))

	numConcurrentEvents := 10
	var wg sync.WaitGroup
	wg.Add(numConcurrentEvents)

	var processingError error

	mustSubscribe(pool, "testEvent", func(_ *EventContext, event Event) error {
		time.Sleep(100 * time.Millisecond)
		wg.Done()
		return nil
	})

	for i := 0; i < numConcurrentEvents; i++ {
		go func() {
			errChan := emitPayload(pool, "testEvent", nil)
			for err := range errChan {
				if err != nil {
					processingError = err
					break
				}
			}
		}()
	}

	wg.Wait()

	if processingError != nil {
		t.Errorf("Error processing event: %v", processingError)
	}
}

func TestMultipleListenersOnSameTopic(t *testing.T) {
	pool := New(Workers(5))

	var count1, count2 int32

	mustSubscribe(pool, "topic", func(_ *EventContext, e Event) error {
		atomic.AddInt32(&count1, 1)
		return nil
	})
	mustSubscribe(pool, "topic", func(_ *EventContext, e Event) error {
		atomic.AddInt32(&count2, 1)
		return nil
	})

	errChan := emitPayload(pool, "topic", nil)
	for err := range errChan {
		if err != nil {
			t.Fatalf("Listener error: %v", err)
		}
	}

	time.Sleep(20 * time.Millisecond)

	if atomic.LoadInt32(&count1) != 1 || atomic.LoadInt32(&count2) != 1 {
		t.Fatalf("Expected both listeners to be called once, got %d and %d", count1, count2)
	}
}

func TestRemoveListener(t *testing.T) {
	pool := New(Workers(2))

	var count int32
	listenerID := mustSubscribe(pool, "topic", func(_ *EventContext, e Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if err := pool.Off("topic", listenerID); err != nil {
		t.Fatalf("Error removing listener: %v", err)
	}

	errChan := emitPayload(pool, "topic", nil)
	for err := range errChan {
		if err != nil {
			t.Fatalf("Listener error: %v", err)
		}
	}

	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&count) != 0 {
		t.Fatalf("Expected removed listener to not be called, got %d", count)
	}
}

func TestEmitNoListeners(t *testing.T) {
	pool := New(Workers(2))
	errChan := emitPayload(pool, "noListeners", nil)
	for err := range errChan {
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}

func TestListenerErrorReporting(t *testing.T) {
	pool := New(Workers(2))

	someErr := errors.New("listener error")
	mustSubscribe(pool, "topic", func(_ *EventContext, e Event) error {
		return someErr
	})

	errChan := emitPayload(pool, "topic", nil)
	var gotErr error
	for err := range errChan {
		if err != nil {
			gotErr = err
		}
	}

	if gotErr == nil || gotErr.Error() != someErr.Error() {
		t.Fatalf("Expected error '%v', got '%v'", someErr, gotErr)
	}
}

func TestPoolWithPoolName(t *testing.T) {
	pool := NewPool(WithPoolName("test-pool"))
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	pool.Release()
}

func TestPoolResize(t *testing.T) {
	pool := NewPool(WithWorkers(5))

	pool.Resize(10)

	pool.Release()
}

func TestOffNonexistentTopic(t *testing.T) {
	bus := New()
	err := bus.Off("nonexistent", "some-id")
	if err == nil {
		t.Error("expected error when removing listener from nonexistent topic")
	}
}

func TestOffNonexistentListener(t *testing.T) {
	bus := New()
	_, err := bus.On("topic", func(_ *EventContext) error { return nil })
	if err != nil {
		t.Fatalf("On() failed: %v", err)
	}

	err = bus.Off("topic", "nonexistent-id")
	if err == nil {
		t.Error("expected error when removing nonexistent listener")
	}
}

func TestEmitToClosedBus(t *testing.T) {
	bus := New()
	if err := bus.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	errChan := bus.Emit("topic", "payload")
	var gotErr error
	for err := range errChan {
		gotErr = err
	}

	if gotErr != ErrEmitterClosed {
		t.Errorf("expected ErrEmitterClosed, got %v", gotErr)
	}
}

func TestEmitInvalidTopicName(t *testing.T) {
	bus := New()

	errChan := bus.Emit("", "payload")
	var gotErr error
	for err := range errChan {
		gotErr = err
	}

	if gotErr == nil {
		t.Error("expected error for empty topic name")
	}
}

func TestInterestedInNoListeners(t *testing.T) {
	bus := New()

	if bus.InterestedIn("nonexistent") {
		t.Error("expected InterestedIn to return false for topic with no listeners")
	}
}
