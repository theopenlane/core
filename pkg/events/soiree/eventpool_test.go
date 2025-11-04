package soiree

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func mustOnTopic(pool *EventPool, topic string, listener TypedListener[Event], opts ...ListenerOption) string {
	id, err := OnTopic(pool, NewEventTopic(topic), listener, opts...)
	if err != nil {
		panic(err)
	}

	return id
}

func emitTopicWithPayload(pool *EventPool, topic string, payload any) <-chan error {
	return EmitTopic(pool, NewEventTopic(topic), Event(NewBaseEvent(topic, payload)))
}

func emitExistingEvent(pool *EventPool, event Event) <-chan error {
	return EmitTopic(pool, NewEventTopic(event.Topic()), event)
}

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
	topic := "testTopic"

	listener := func(_ *EventContext, e Event) error {
		return nil
	}

	// On to a topic
	id := mustOnTopic(soiree, topic, listener)

	if id == "" {
		t.Fatal("Onrned an empty ID")
	}

	// Now unsubscribe and ensure the listener is removed
	if err := soiree.Off(topic, id); err != nil {
		t.Fatalf("Off() failed with error: %v", err)
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

// TestEmitSyncSuccess tests emitting to a topic
func TestEmitSyncSuccess(t *testing.T) {
	soiree := NewEventPool()
	received := make(chan string, 1) // Buffered channel to receive one message
	topic := "testTopic"
	event := NewTestEvent(topic, "testPayload")

	// Prepare the listener
	listener := createTestListener(received)

	mustOnTopic(soiree, topic, listener)

	emitExistingEvent(soiree, event)

	// Wait for the listener to handle the event or timeout after a specific duration
	select {
	case topic := <-received:
		if topic != event.Topic() {
			t.Fatalf("Expected to receive event on '%s', got '%v'", event.Topic(), topic)
		}

	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for the event to be received")
	}
}

// TestEmitSyncFailure tests the synchronous EmitSync method for event handling that returns an error
func TestEmitSyncFailure(t *testing.T) {
	soiree := NewEventPool()
	topic := "testTopic"
	event := NewTestEvent(topic, "testPayload")

	// Create a listener that returns an error
	listener := func(_ *EventContext, e Event) error {
		return errors.New("listener error") // nolint: err113
	}

	mustOnTopic(soiree, topic, listener)

	// Emit the event synchronously and collect errors
	errsSync := EmitTopicSync(soiree, NewEventTopic(event.Topic()), Event(event))

	// Check that the errors juicy slice is not empty
	if len(errsSync) == 0 {
		t.Error("EmitSync() should have resulted in errors, but none were returned")
	}
}

// TestGetTopic tests getting a topic
func TestGetTopic(t *testing.T) {
	soiree := NewEventPool()
	event := NewTestEvent("testTopic", "testPayload")

	// Creating a topic by subscribing to it
	mustOnTopic(soiree, event.Topic(), func(_ *EventContext, e Event) error { return nil })

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
		_, err := OnTopic(soiree, NewEventTopic(event.Topic()), func(ctx *EventContext, e Event) error {
			// Record the event in the receivedEvents map
			eventKey := e.Topic()
			t.Logf("Listener received event on topic: %s with payload: %s", topicName, eventKey)
			payloadEvents, _ := receivedEvents.LoadOrStore(eventKey, new(sync.Map))
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
		emitExistingEvent(soiree, event) // Use the eventKey as the payload for identification
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
	listener1 := func(_ *EventContext, e Event) error { return nil }
	mustOnTopic(soiree, topic1, listener1)

	topic2 := "topic2"
	listener2 := func(_ *EventContext, e Event) error { return nil }
	mustOnTopic(soiree, topic2, listener2)

	var err error

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
	errChan := emitTopicWithPayload(soiree, topic1, "payload")
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

func createTestListener(received chan<- string) TypedListener[Event] {
	return func(_ *EventContext, e Event) error {
		// Send the topic to the received channel
		received <- e.Topic()
		return nil
	}
}
