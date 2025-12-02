package soiree

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alitto/pond/v2"
)

func mustSubscribe(pool *EventPool, topic string, listener TypedListener[Event], opts ...ListenerOption) string {
	id, err := BindListener(typedEventTopic(topic), listener, opts...).Register(pool)
	if err != nil {
		panic(err)
	}

	return id
}

func emitPayload(pool *EventPool, topic string, payload any) <-chan error {
	return pool.Emit(topic, Event(NewBaseEvent(topic, payload)))
}

func TestEmitEventWithPool(t *testing.T) {
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(10))))

	var processedEvents int32

	listenerID := mustSubscribe(soiree, "testEvent", func(_ *EventContext, event Event) error {
		atomic.AddInt32(&processedEvents, 1)
		time.Sleep(10 * time.Millisecond) // Simulating work

		return nil
	})

	errChan := emitPayload(soiree, "testEvent", nil)

	// Collect all errors from the channel
	var errors []error

	go func() {
		for err := range errChan {
			if err != nil {
				errors = append(errors, err)
			}
		}
	}()

	// Wait for a short duration to ensure event processing has a chance to complete
	time.Sleep(100 * time.Millisecond)

	// Check for errors reported by the listener
	if len(errors) > 0 {
		t.Fatalf("Listener reported errors: %v", errors)
	}

	// Unregister the listener as cleanup
	if err := soiree.Off("testEvent", listenerID); err != nil {
		t.Errorf("Failed to unregister listener: %v", err)
	}

	// Final assertion after cleanup
	if atomic.LoadInt32(&processedEvents) != 1 {
		t.Fatalf("Expected 1 event to be processed, but got %d", processedEvents)
	}
}

func TestEmitMultipleEventsWithPool(t *testing.T) {
	// Create a EventPool instance with a PondPool.
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(10))))

	// Define the number of concurrent events to emit
	numConcurrentEvents := 10

	// Define a wait group to wait for all events to be processed
	var wg sync.WaitGroup

	wg.Add(numConcurrentEvents)

	// Define a variable to keep track of any errors encountered during event processing
	var processingError error

	// Add an event listener to handle "testEvent" and increment the processedEvents count
	mustSubscribe(soiree, "testEvent", func(_ *EventContext, event Event) error {
		// Simulate some processing
		time.Sleep(100 * time.Millisecond)

		// Decrement the wait group to signal event processing completion
		wg.Done()

		return nil
	})

	// Emit multiple events concurrently
	for i := 0; i < numConcurrentEvents; i++ {
		go func() {
			// Emit an event using the soiree
			errChan := emitPayload(soiree, "testEvent", nil)

			// Wait for the event to be processed
			for err := range errChan {
				if err != nil {
					// Capture the first error encountered during event processing
					processingError = err
					break
				}
			}
		}()
	}

	// Wait for all events to be processed
	wg.Wait()

	// Check if any errors occurred during event processing
	if processingError != nil {
		t.Errorf("Error processing event: %v", processingError)
	}
}

func TestMultipleListenersOnSameTopic(t *testing.T) {
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(5))))

	var count1, count2 int32

	mustSubscribe(soiree, "topic", func(_ *EventContext, e Event) error {
		atomic.AddInt32(&count1, 1)
		return nil
	})
	mustSubscribe(soiree, "topic", func(_ *EventContext, e Event) error {
		atomic.AddInt32(&count2, 1)
		return nil
	})

	errChan := emitPayload(soiree, "topic", nil)
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
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(2))))

	var count int32
	listenerID := mustSubscribe(soiree, "topic", func(_ *EventContext, e Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if err := soiree.Off("topic", listenerID); err != nil {
		t.Fatalf("Error removing listener: %v", err)
	}

	errChan := emitPayload(soiree, "topic", nil)
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
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(2))))
	errChan := emitPayload(soiree, "noListeners", nil)
	for err := range errChan {
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}

func TestListenerErrorReporting(t *testing.T) {
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(2))))

	someErr := errors.New("listener error")
	mustSubscribe(soiree, "topic", func(_ *EventContext, e Event) error {
		return someErr
	})

	errChan := emitPayload(soiree, "topic", nil)
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
func TestPondPool_SuccessfulTasks(t *testing.T) {
	pool := NewPondPool(WithMaxWorkers(2))
	defer pool.Release()

	const expected = 2
	var wg sync.WaitGroup
	wg.Add(expected)

	for i := 0; i < expected; i++ {
		pool.Submit(func() {
			defer wg.Done()
		})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("tasks did not complete in time")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if pool.SuccessfulTasks() == expected {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got := pool.SuccessfulTasks(); got != expected {
		t.Fatalf("SuccessfulTasks() = %d, want %d", got, expected)
	}
}
func TestWithName_SetsName(t *testing.T) {
	pool := &PondPool{}
	name := "test-pool"
	opt := WithName(name)
	opt(pool)
	if pool.name != name {
		t.Errorf("WithName did not set the name correctly, got %q, want %q", pool.name, name)
	}
}

func TestWithName_EmptyString(t *testing.T) {
	pool := &PondPool{name: "existing"}
	opt := WithName("")
	opt(pool)
	if pool.name != "" {
		t.Errorf("WithName with empty string did not clear the name, got %q", pool.name)
	}
}

func TestWithOptions_EmptyOptions(t *testing.T) {
	pool := &PondPool{}
	WithOptions()(pool)
	if len(pool.opts) != 0 {
		t.Errorf("Expected no options to be set, got %d", len(pool.opts))
	}
}
func TestPondPool_SubmitMultipleAndWait_AllTasksRun(t *testing.T) {
	pool := NewPondPool(WithMaxWorkers(4))
	defer pool.Release()

	var count int32
	tasks := make([]func(), 10)
	for i := range tasks {
		tasks[i] = func() {
			atomic.AddInt32(&count, 1)
			time.Sleep(10 * time.Millisecond)
		}
	}

	pool.SubmitMultipleAndWait(tasks)

	if atomic.LoadInt32(&count) != int32(len(tasks)) {
		t.Errorf("Expected %d tasks to run, got %d", len(tasks), count)
	}
}

func TestPondPool_SubmitMultipleAndWait_EmptySlice(t *testing.T) {
	pool := NewPondPool(WithMaxWorkers(2))
	defer pool.Release()

	done := make(chan struct{})
	go func() {
		pool.SubmitMultipleAndWait(nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("SubmitMultipleAndWait with nil slice did not return promptly")
	}

	done2 := make(chan struct{})
	go func() {
		pool.SubmitMultipleAndWait([]func(){})
		close(done2)
	}()

	select {
	case <-done2:
	case <-time.After(100 * time.Millisecond):
		t.Error("SubmitMultipleAndWait with empty slice did not return promptly")
	}
}

func TestPondPool_SubmitMultipleAndWait_TasksRunConcurrently(t *testing.T) {
	pool := NewPondPool(WithMaxWorkers(3))
	defer pool.Release()

	var mu sync.Mutex
	started := 0
	tasks := []func(){
		func() {
			mu.Lock()
			started++
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
		},
		func() {
			mu.Lock()
			started++
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
		},
		func() {
			mu.Lock()
			started++
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
		},
	}

	start := time.Now()
	pool.SubmitMultipleAndWait(tasks)
	elapsed := time.Since(start)

	if started != 3 {
		t.Errorf("Expected all tasks to start, got %d", started)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected tasks to run concurrently, took too long: %v", elapsed)
	}
}

type fakePondPool struct {
	pond.Pool
	waitingTasks    uint64
	submittedTasks  uint64
	successfulTasks uint64
	failedTasks     uint64
	completedTasks  uint64
}

func (f *fakePondPool) WaitingTasks() uint64 {
	return f.waitingTasks
}

func (f *fakePondPool) SubmittedTasks() uint64 {
	return f.submittedTasks
}

func (f *fakePondPool) SuccessfulTasks() uint64 {
	return f.successfulTasks
}

func (f *fakePondPool) FailedTasks() uint64 {
	return f.failedTasks
}

func (f *fakePondPool) CompletedTasks() uint64 {
	return f.completedTasks
}

func TestPondPool_SubmittedTasks_DelegatesToPool(t *testing.T) {
	var fakePool = &fakePondPool{submittedTasks: 42}

	// Patch PondPool to use our fake pool
	pool := &PondPool{pool: fakePool}

	got := pool.SubmittedTasks()
	if got != 42 {
		t.Errorf("SubmittedTasks() = %d, want 42", got)
	}
}

func TestPondPool_SubmittedTasks_Overflow(t *testing.T) {
	fakePool := &fakePondPool{submittedTasks: uint64(math.MaxInt) + 10}
	pool := &PondPool{pool: fakePool}

	got := pool.SubmittedTasks()
	if got != math.MaxInt {
		t.Errorf("SubmittedTasks() = %d, want %d", got, math.MaxInt)
	}
}
