package soiree

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// TestNewEventPool tests the creation of a new EventPool
func TestNewEventPool(t *testing.T) {
	soiree := NewEventPool()
	if soiree == nil {
		t.Fatal("NewEventPool() should not return nil")
	}
}

// TestOnOff tests subscribing to and unsubscribing from a topic
func TestOnOff(t *testing.T) {
	soiree := NewEventPool()

	listener := func(e Event) error {
		return nil
	}

	// On to a topic
	id, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	if id == "" {
		t.Fatal("Onrned an empty ID")
	}

	// Now unsubscribe and ensure the listener is removed
	if err := soiree.Off("testTopic", id); err != nil {
		t.Fatalf("Off() failed with error: %v", err)
	}
}

// TestEmitAsyncSuccess tests the asynchronous Emit method for successful event handling
func TestEmitAsyncSuccess(t *testing.T) {
	soiree := NewEventPool()

	// Create a listener that does not return an error
	listener := func(e Event) error {
		// Simulate some work.
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	// Subscribe the listener to the "testTopic"
	_, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Emit the event asynchronously
	errChan := soiree.Emit("testTopic", "testPayload")

	// Collect errors from the error channel
	var emitErrors []error

	for err := range errChan {
		if err != nil {
			emitErrors = append(emitErrors, err)
		}
	}

	// Check that there were no errors during emission
	if len(emitErrors) != 0 {
		t.Errorf("Emit() resulted in errors: %v", emitErrors)
	}
}

type TestEvent struct {
	*BaseEvent
}

func NewTestEvent(topic string, payload any) *TestEvent {
	return &TestEvent{
		BaseEvent: NewBaseEvent(topic, payload),
	}
}

// TestEmitAsyncFailure tests the asynchronous Emit method for event handling that returns an error
func TestEmitAsyncFailure(t *testing.T) {
	soiree := NewEventPool()
	event := NewTestEvent("testTopic", "test payload")

	// Create a listener that returns an error
	listener := func(e Event) error {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		return errors.New("listener error") // nolint: err113
	}

	// Subscribe the listener to the "testTopic"
	_, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Emit the event asynchronously
	errChan := soiree.Emit(event.Topic(), event)

	// Collect errors from the error channel
	var emitErrors []error

	for err := range errChan {
		if err != nil {
			emitErrors = append(emitErrors, err)
		}
	}

	// Check that the errors slice is not empty, indicating that an error was returned by the listener
	if len(emitErrors) == 0 {
		t.Error("Emit() should have resulted in errors, but none were returned")
	}
}

// TestEmitSyncSuccess tests emitting to a topic
func TestEmitSyncSuccess(t *testing.T) {
	soiree := NewEventPool()
	received := make(chan string, 1) // Buffered channel to receive one message
	event := NewTestEvent("testTopic", "testPayload")

	// Prepare the listener
	listener := createTestListener(received)

	_, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	soiree.Emit(event.Topic(), event)

	// Wait for the listener to handle the event or timeout after a specific duration
	select {
	case topic := <-received:
		if topic != "testTopic" {
			t.Fatalf("Expected to receive event on 'testTopic', got '%v'", topic)
		}

	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for the event to be received")
	}
}

// TestEmitSyncFailure tests the synchronous EmitSync method for event handling that returns an error
func TestEmitSyncFailure(t *testing.T) {
	soiree := NewEventPool()
	event := NewTestEvent("testTopic", "testPayload")

	// Create a listener that returns an error
	listener := func(e Event) error {
		return errors.New("listener error") // nolint: err113
	}

	_, err := soiree.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Emit the event synchronously and collect errors
	errors := soiree.EmitSync(event.Topic(), event)

	// Check that the errors juicy slice is not empty
	if len(errors) == 0 {
		t.Error("EmitSync() should have resulted in errors, but none were returned")
	}
}

// TestGetTopic tests getting a topic
func TestGetTopic(t *testing.T) {
	soiree := NewEventPool()
	event := NewTestEvent("testTopic", "testPayload")

	// Creating a topic by subscribing to it
	_, err := soiree.On("testTopic", func(e Event) error { return nil })
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	topic, err := soiree.GetTopic(event.Topic())

	if err != nil {
		t.Fatalf("GetTopic() failed with error: %v", err)
	}

	if topic == nil {
		t.Fatal("GetTopic() returned nil")
	}
}

// TestEnsureTopic tests getting or creating a topic
func TestEnsureTopic(t *testing.T) {
	soiree := NewEventPool()

	// Get or create a new topic
	topic := soiree.EnsureTopic("newTopic")
	if topic == nil {
		t.Fatal("EnsureTopic() should not return nil")
	}

	// Try to retrieve the same topic and check if it's the same instance
	sameTopic, err := soiree.GetTopic("newTopic")
	if err != nil {
		t.Fatalf("GetTopic() failed with error: %v", err)
	}

	if sameTopic != topic {
		t.Fatal("EnsureTopic() did not return the same instance of the topic")
	}
}

func TestWildcardSubscriptionAndEmitting(t *testing.T) {
	soiree := NewEventPool()

	topics := []string{
		"event.some.*.*",
		"event.some.*.run",
		"event.some.**",
		"**.thing.run",
	}

	expectedMatches := map[string][]string{
		"event.some.thing.run": {"event.some.*.*", "event.some.*.run", "event.some.**", "**.thing.run"},
		"event.some.thing.do":  {"event.some.*.*", "event.some.**"},
		"event.some.thing":     {"event.some.**"},
	}

	receivedEvents := new(sync.Map) // A concurrent map to store received events

	// On the mock listener to all topics
	for _, topic := range topics {
		event := NewTestEvent(topic, "testPayload")
		topicName := topic
		_, err := soiree.On(event.Topic(), func(e Event) error {
			// Record the event in the receivedEvents map
			eventPayload := e.Payload().(string)
			t.Logf("Listener received event on topic: %s with payload: %s", topicName, eventPayload)
			payloadEvents, _ := receivedEvents.LoadOrStore(eventPayload, new(sync.Map))
			payloadEvents.(*sync.Map).Store(topicName, struct{}{})

			return nil
		})

		if err != nil {
			t.Fatalf("Failed to subscribe to topic %s: %s", topic, err)
		}
	}

	// Emit events to all topics and check if the listeners are notified
	for eventKey := range expectedMatches {
		event := NewTestEvent(eventKey, eventKey)
		t.Logf("Emitting event: %s", eventKey)
		soiree.Emit(event.Topic(), event) // Use the eventKey as the payload for identification
	}

	// Allow some time for the events to be processed asynchronously
	time.Sleep(1 * time.Second) // use synchronization primitives instead of Sleep?

	// Verify that the correct listeners were notified
	for eventKey, expectedTopics := range expectedMatches {
		if topics, ok := receivedEvents.Load(eventKey); ok {
			receivedTopics := make([]string, 0)

			topics.(*sync.Map).Range(func(key, value any) bool {
				receivedTopic := key.(string)
				receivedTopics = append(receivedTopics, receivedTopic)

				return true
			})

			for _, expectedTopic := range expectedTopics {
				if !contains(receivedTopics, expectedTopic) {
					t.Errorf("Expected topic %s to be notified for event %s, but it was not", expectedTopic, eventKey)
				}
			}
		} else {
			t.Errorf("No events received for eventKey %s", eventKey)
		}
	}
}

func TestEventPoolClose(t *testing.T) {
	soiree := NewEventPool()

	// Set up topics and listeners
	topic1 := "topic1"
	listener1 := func(e Event) error { return nil }

	_, err := soiree.On(topic1, listener1)

	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	topic2 := "topic2"
	listener2 := func(e Event) error { return nil }

	_, err = soiree.On(topic2, listener2)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Use a Pool
	pool := NewPondPool(WithMaxWorkers(10))
	soiree.SetPool(pool)

	// Close the soiree
	if err := soiree.Close(); err != nil {
		t.Fatalf("Close() failed with error: %v", err)
	}

	// Verify topics have been removed
	_, err = soiree.GetTopic(topic1)
	if err == nil {
		t.Errorf("GetTopic() should return an error after Close()")
	}

	_, err = soiree.GetTopic(topic2)
	if err == nil {
		t.Errorf("GetTopic() should return an error after Close()")
	}

	// Verify the pool has been released
	if pool.Running() > 0 {
		t.Errorf("Pool should be released and have no running workers after Close()")
	}

	// Verify that no new events can be emitted
	errChan := soiree.Emit(topic1, "payload")
	select {
	case err := <-errChan:
		if err == nil {
			t.Errorf("Emit() should return an error after Close()")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for the error to be received")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

func createTestListener(received chan<- string) func(Event) error {
	return func(e Event) error {
		// Send the topic to the received channel
		received <- e.Topic()
		return nil
	}
}
