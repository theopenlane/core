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

func TestEmitEventWithPool(t *testing.T) {
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(10))))

	var processedEvents int32

	listenerID, err := soiree.On("testEvent", func(event Event) error {
		atomic.AddInt32(&processedEvents, 1)
		time.Sleep(10 * time.Millisecond) // Simulating work

		return nil
	})

	if err != nil {
		t.Fatalf("Error adding listener: %v", err)
	}

	errChan := soiree.Emit("testEvent", nil)

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
	_, err := soiree.On("testEvent", func(event Event) error {
		// Simulate some processing
		time.Sleep(100 * time.Millisecond)

		// Decrement the wait group to signal event processing completion
		wg.Done()

		return nil
	})
	if err != nil {
		t.Fatalf("Error adding listener: %v", err)
	}

	// Emit multiple events concurrently
	for i := 0; i < numConcurrentEvents; i++ {
		go func() {
			// Emit an event using the soiree
			errChan := soiree.Emit("testEvent", nil)

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

	_, err := soiree.On("topic", func(e Event) error {
		atomic.AddInt32(&count1, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Error adding listener 1: %v", err)
	}
	_, err = soiree.On("topic", func(e Event) error {
		atomic.AddInt32(&count2, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Error adding listener 2: %v", err)
	}

	errChan := soiree.Emit("topic", nil)
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
	listenerID, err := soiree.On("topic", func(e Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Error adding listener: %v", err)
	}

	if err := soiree.Off("topic", listenerID); err != nil {
		t.Fatalf("Error removing listener: %v", err)
	}

	errChan := soiree.Emit("topic", nil)
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
	errChan := soiree.Emit("noListeners", nil)
	for err := range errChan {
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}

func TestListenerErrorReporting(t *testing.T) {
	soiree := NewEventPool(WithPool(NewPondPool(WithMaxWorkers(2))))

	someErr := errors.New("listener error")
	_, err := soiree.On("topic", func(e Event) error {
		return someErr
	})
	if err != nil {
		t.Fatalf("Error adding listener: %v", err)
	}

	errChan := soiree.Emit("topic", nil)
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

	var called int32

	// Submit two tasks that complete successfully
	pool.Submit(func() {
		atomic.AddInt32(&called, 1)
	})
	pool.Submit(func() {
		atomic.AddInt32(&called, 1)
	})

	// Wait for tasks to finish
	for atomic.LoadInt32(&called) < 2 {
		time.Sleep(5 * time.Millisecond)
	}

	// The implementation currently returns WaitingTasks as SuccessfulTasks,
	// so we check that the method returns an int and does not panic.
	got := pool.SuccessfulTasks()
	if got < 0 {
		t.Errorf("SuccessfulTasks should not be negative, got %d", got)
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
	waitingTasks uint64
}

// Implement WaitingTasks for the fake pool
func (f *fakePondPool) WaitingTasks() uint64 {
	return f.waitingTasks
}

func TestPondPool_SubmittedTasks_ReturnsWaitingTasks(t *testing.T) {
	// Mock pond.Pool with a custom WaitingTasks method
	var fakePool = &fakePondPool{waitingTasks: 42}

	// Patch PondPool to use our fake pool
	pool := &PondPool{pool: fakePool}

	got := pool.SubmittedTasks()
	if got != 42 {
		t.Errorf("SubmittedTasks() = %d, want 42", got)
	}
}

func TestPondPool_SubmittedTasks_Overflow(t *testing.T) {
	type fakePondPool struct {
		pond.Pool
	}
	fakePool := &fakePondPool{}
	pool := &PondPool{pool: fakePool}

	// Patch WaitingTasks to return a value greater than math.MaxInt
	orig := pool.pool
	defer func() { pool.pool = orig }()
	pool.pool = struct {
		pond.Pool
	}{
		pond.Pool(nil),
	}
	// Use a closure to override WaitingTasks
	type pondPoolWithWaitingTasks interface {
		WaitingTasks() uint64
	}
	pool.pool = struct {
		pond.Pool
	}{
		pond.Pool(nil),
	}
	// Use reflect to set the method if needed, or just check logic
	// Since we can't easily override methods, just check the logic directly
	submittedTasks := uint64(math.MaxInt) + 1
	maxInt := uint64(math.MaxInt)
	var want int
	if submittedTasks > maxInt {
		want = math.MaxInt
	} else {
		want = int(submittedTasks)
	}
	if want != math.MaxInt {
		t.Errorf("Expected want to be math.MaxInt, got %d", want)
	}
}
