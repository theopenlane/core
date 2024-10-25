package soiree

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEmitEventWithPool(t *testing.T) {
	soiree := NewEventPool(WithPool(NewPondPool(10)))

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
	soiree := NewEventPool(WithPool(NewPondPool(10)))

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
