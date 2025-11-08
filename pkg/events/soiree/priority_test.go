package soiree

import (
	"sync"
	"testing"
)

func emitEvent(pool *EventPool, event Event) {
	for err := range pool.Emit(event.Topic(), event) {
		if err != nil {
			panic(err)
		}
	}
}

// TestPriorityOrdering checks if the Soiree calls listeners in the correct order of their priorities
func TestPriorityOrdering(t *testing.T) {
	pool := NewEventPool()

	var (
		mu        sync.Mutex
		callOrder []Priority
		wg        sync.WaitGroup
	)

	event := NewTestEvent("test_priority_topic", "test_payload")

	subscribe := func(priority Priority) {
		wg.Add(1)

		_, err := BindListener(typedEventTopic(event.Topic()), func(_ *EventContext, e Event) error {
			defer wg.Done()

			mu.Lock()
			callOrder = append(callOrder, priority)
			mu.Unlock()

			return nil
		}, WithPriority(priority)).Register(pool)
		if err != nil {
			t.Fatalf("failed to register listener: %v", err)
		}
	}

	// Set up listeners with different priorities
	subscribe(High)
	subscribe(Low)
	subscribe(Normal)
	subscribe(Lowest)
	subscribe(Highest)

	emitEvent(pool, event)
	wg.Wait()

	expectedOrder := []Priority{Highest, High, Normal, Low, Lowest}

	mu.Lock()
	defer mu.Unlock()

	if len(callOrder) != len(expectedOrder) {
		t.Fatalf("call order length %d, expected %d", len(callOrder), len(expectedOrder))
	}

	for i, priority := range expectedOrder {
		if callOrder[i] != priority {
			t.Errorf("expected priority %v at index %d, got %v", priority, i, callOrder[i])
		}
	}
}

// TestEmitSyncWithAbort tests the synchronous EmitSync method with a listener that aborts the event
func TestEmitSyncWithAbort(t *testing.T) {
	pool := NewEventPool()
	topic := "testTopic"

	highPriorityListener := func(_ *EventContext, e Event) error { return nil }
	abortingListener := func(_ *EventContext, e Event) error {
		e.SetAborted(true)
		return nil
	}
	lowPriorityListener := func(_ *EventContext, e Event) error {
		t.Error("low priority listener should not be invoked after abort")
		return nil
	}

	if _, err := BindListener(typedEventTopic(topic), lowPriorityListener, WithPriority(Low)).Register(pool); err != nil {
		t.Fatalf("register low priority listener: %v", err)
	}
	if _, err := BindListener(typedEventTopic(topic), abortingListener, WithPriority(Normal)).Register(pool); err != nil {
		t.Fatalf("register aborting listener: %v", err)
	}
	if _, err := BindListener(typedEventTopic(topic), highPriorityListener, WithPriority(High)).Register(pool); err != nil {
		t.Fatalf("register high priority listener: %v", err)
	}

	pool.EmitSync(topic, NewBaseEvent(topic, "testPayload"))
}

// TestEmitWithAbort tests the asynchronous Emit method with a listener that aborts the event
func TestEmitWithAbort(t *testing.T) {
	pool := NewEventPool()
	topic := "testTopic"

	highPriorityListener := func(_ *EventContext, e Event) error { return nil }
	abortingListener := func(_ *EventContext, e Event) error {
		e.SetAborted(true)
		return nil
	}

	var lowPriorityListenerCalled bool
	lowPriorityListener := func(_ *EventContext, e Event) error {
		lowPriorityListenerCalled = true
		return nil
	}

	if _, err := BindListener(typedEventTopic(topic), lowPriorityListener, WithPriority(Low)).Register(pool); err != nil {
		t.Fatalf("register low priority listener: %v", err)
	}
	if _, err := BindListener(typedEventTopic(topic), abortingListener, WithPriority(Normal)).Register(pool); err != nil {
		t.Fatalf("register aborting listener: %v", err)
	}
	if _, err := BindListener(typedEventTopic(topic), highPriorityListener, WithPriority(High)).Register(pool); err != nil {
		t.Fatalf("register high priority listener: %v", err)
	}

	for err := range pool.Emit(topic, NewBaseEvent(topic, "testPayload")) {
		if err != nil {
			t.Fatalf("unexpected emit error: %v", err)
		}
	}

	if lowPriorityListenerCalled {
		t.Error("low priority listener should not have been called")
	}
}
