package soiree

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func mustOnTopic(pool *EventPool, topic string, listener TypedListener[Event], opts ...ListenerOption) string {
	id, err := BindListener(typedEventTopic(topic), listener, opts...).Register(pool)
	if err != nil {
		panic(err)
	}

	return id
}

func emitTopicWithPayload(pool *EventPool, topic string, payload any) <-chan error {
	return pool.Emit(topic, Event(NewBaseEvent(topic, payload)))
}

func emitExistingEvent(pool *EventPool, event Event) <-chan error {
	return pool.Emit(event.Topic(), event)
}

func TestEventPoolBasics(t *testing.T) {
	pool := NewEventPool()
	if pool == nil {
		t.Fatal("NewEventPool() should not return nil")
	}

	topic := "testTopic"
	listener := func(_ *EventContext, e Event) error { return nil }

	id := mustOnTopic(pool, topic, listener)
	if id == "" {
		t.Fatal("expected listener id")
	}

	if err := pool.Off(topic, id); err != nil {
		t.Fatalf("Off() failed with error: %v", err)
	}

	// EnsureTopic should always return the same instance.
	created := pool.EnsureTopic("newTopic")
	if created == nil {
		t.Fatal("EnsureTopic() should not return nil")
	}

	retrieved, err := pool.GetTopic("newTopic")
	if err != nil {
		t.Fatalf("GetTopic() failed with error: %v", err)
	}
	if created != retrieved {
		t.Fatal("EnsureTopic() should return the same topic instance")
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

func TestEmitSync(t *testing.T) {
	pool := NewEventPool()
	topic := "testTopic"
	event := NewTestEvent(topic, "testPayload")

	cases := []struct {
		name      string
		listener  func(*EventContext, Event) error
		expectErr bool
	}{
		{
			name: "success",
			listener: func(_ *EventContext, e Event) error {
				return nil
			},
			expectErr: false,
		},
		{
			name: "failure",
			listener: func(_ *EventContext, e Event) error {
				return errors.New("listener error") // nolint: err113
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool.topics.Range(func(key, value any) bool {
				pool.topics.Delete(key)
				return true
			})

			mustOnTopic(pool, topic, tc.listener)
			errs := pool.EmitSync(event.Topic(), Event(event))

			if tc.expectErr && len(errs) == 0 {
				t.Fatal("expected emit sync to return errors")
			}

			if !tc.expectErr && len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
		})
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
		_, err := BindListener(typedEventTopic(event.Topic()), func(ctx *EventContext, e Event) error {
			// Record the event in the receivedEvents map
			eventKey := e.Topic()
			t.Logf("Listener received event on topic: %s with payload: %s", topicName, eventKey)
			payloadEvents, _ := receivedEvents.LoadOrStore(eventKey, new(sync.Map))
			payloadEvents.(*sync.Map).Store(topicName, struct{}{})

			return nil
		}).Register(soiree)

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
