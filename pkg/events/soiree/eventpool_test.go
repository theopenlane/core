package soiree

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func mustOnTopic(pool *EventBus, topic string, listener TypedListener[Event]) string {
	id, err := BindListener(typedEventTopic(topic), listener).Register(pool)
	if err != nil {
		panic(err)
	}
	return id
}

func emitTopicWithPayload(pool *EventBus, topic string, payload any) <-chan error {
	return pool.Emit(topic, Event(NewBaseEvent(topic, payload)))
}

func emitExistingEvent(pool *EventBus, event Event) <-chan error {
	return pool.Emit(event.Topic(), event)
}

func TestEventPoolBasics(t *testing.T) {
	pool := New()
	if pool == nil {
		t.Fatal("New() should not return nil")
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
}

func TestInterestedInWildcard(t *testing.T) {
	pool := New()

	_, err := pool.On("event.*", func(_ *EventContext) error { return nil })
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	if !pool.InterestedIn("event.created") {
		t.Fatal("expected InterestedIn to return true for wildcard match")
	}

	if pool.InterestedIn("other.created") {
		t.Fatal("expected InterestedIn to return false for non-matching topic")
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

func TestEmitCollectErrors(t *testing.T) {
	pool := New()
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
				return errors.New("listener error") //nolint:err113
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
			errs := collectEmitErrors(pool.Emit(event.Topic(), Event(event)))

			if tc.expectErr && len(errs) == 0 {
				t.Fatal("expected emit to return errors")
			}

			if !tc.expectErr && len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestEventContextPayloadAccess(t *testing.T) {
	pool := New()
	payload := map[string]string{"name": "Ada"}

	_, err := pool.On("payload", func(ctx *EventContext) error {
		if ctx.Event() == nil {
			t.Fatal("expected Event() to return an event")
		}

		if ctx.Payload() == nil {
			t.Fatal("expected Payload() to return a payload")
		}

		got, ok := PayloadAs[map[string]string](ctx)
		if !ok {
			t.Fatal("expected PayloadAs to succeed")
		}

		if got["name"] != "Ada" {
			t.Fatalf("unexpected payload contents: %v", got)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	errs := collectEmitErrors(pool.Emit("payload", payload))
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestWildcardSubscriptionAndEmitting(t *testing.T) {
	soiree := New()

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

	// Calculate total expected listener invocations
	expectedCalls := 0
	for _, matches := range expectedMatches {
		expectedCalls += len(matches)
	}

	var wg sync.WaitGroup
	wg.Add(expectedCalls)

	receivedEvents := new(sync.Map) // A concurrent map to store received events

	// On the mock listener to all topics
	for _, topic := range topics {
		event := NewTestEvent(topic, "testPayload")
		topicName := topic
		_, err := BindListener(typedEventTopic(event.Topic()), func(ctx *EventContext, e Event) error {
			defer wg.Done()
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

	// Wait for all listeners to be invoked
	wg.Wait()

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
	pool := New(Workers(10))

	topic1 := "topic1"
	listener1 := func(_ *EventContext, e Event) error { return nil }
	mustOnTopic(pool, topic1, listener1)

	topic2 := "topic2"
	listener2 := func(_ *EventContext, e Event) error { return nil }
	mustOnTopic(pool, topic2, listener2)

	if err := pool.Close(); err != nil {
		t.Fatalf("Close() failed with error: %v", err)
	}

	// Verify that no new events can be emitted
	errChan := emitTopicWithPayload(pool, topic1, "payload")
	select {
	case err := <-errChan:
		if err == nil {
			t.Errorf("Emit() should return an error after Close()")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for the error to be received")
	}
}

func TestListenerCanRemoveListener(t *testing.T) {
	pool := New()

	var id string
	var err error
	id, err = pool.On("topic", func(_ *EventContext) error {
		return pool.Off("topic", id)
	})
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	errCh := make(chan []error, 1)
	go func() {
		errCh <- collectEmitErrors(pool.Emit("topic", "payload"))
	}()

	select {
	case errs := <-errCh:
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
	case <-time.After(time.Second):
		t.Fatal("Emit deadlocked while removing listener")
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

func TestEventBusClient(t *testing.T) {
	type testClient struct {
		name string
	}

	client := &testClient{name: "test"}
	bus := New(Client(client))

	if bus.Client() != client {
		t.Fatal("expected client to be set on bus")
	}
}

func TestRegisterListeners(t *testing.T) {
	bus := New()
	topic := typedEventTopic("test.topic")

	binding1 := BindListener(topic, func(_ *EventContext, _ Event) error { return nil })
	binding2 := BindListener(topic, func(_ *EventContext, _ Event) error { return nil })

	ids, err := bus.RegisterListeners(binding1, binding2)
	if err != nil {
		t.Fatalf("RegisterListeners failed: %v", err)
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 listener IDs, got %d", len(ids))
	}

	for _, id := range ids {
		if id == "" {
			t.Fatal("expected non-empty listener ID")
		}
	}
}

func TestEventContextProperties(t *testing.T) {
	bus := New()

	done := make(chan struct{}, 1)
	_, err := bus.On("props.test", func(ctx *EventContext) error {
		props := ctx.Properties()
		if props == nil {
			t.Error("expected properties to be non-nil")
		}

		props.Set("key1", "value1")
		props.Set("key2", 42)

		val, ok := ctx.Property("key1")
		if !ok || val != "value1" {
			t.Errorf("expected key1=value1, got %v", val)
		}

		str, ok := ctx.PropertyString("key1")
		if !ok || str != "value1" {
			t.Errorf("expected string value1, got %v", str)
		}

		_, ok = ctx.PropertyString("key2")
		if ok {
			t.Error("expected PropertyString to fail for non-string value")
		}

		_, ok = ctx.Property("nonexistent")
		if ok {
			t.Error("expected Property to return false for nonexistent key")
		}

		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("On() failed: %v", err)
	}

	bus.Emit("props.test", "payload")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener did not run")
	}
}

func TestClientAs(t *testing.T) {
	type testClient struct {
		name string
	}

	client := &testClient{name: "myClient"}
	bus := New(Client(client))

	done := make(chan struct{}, 1)
	_, err := bus.On("client.test", func(ctx *EventContext) error {
		c, ok := ClientAs[*testClient](ctx)
		if !ok {
			t.Error("expected ClientAs to succeed")
		}
		if c.name != "myClient" {
			t.Errorf("expected client name myClient, got %s", c.name)
		}

		_, ok = ClientAs[string](ctx)
		if ok {
			t.Error("expected ClientAs to fail for wrong type")
		}

		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("On() failed: %v", err)
	}

	bus.Emit("client.test", "payload")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener did not run")
	}
}

func TestEventContextNilSafety(t *testing.T) {
	var ctx *EventContext

	if ctx.Context() == nil {
		t.Error("expected Context() to return background context for nil EventContext")
	}

	if ctx.Event() != nil {
		t.Error("expected Event() to return nil for nil EventContext")
	}

	if ctx.Payload() != nil {
		t.Error("expected Payload() to return nil for nil EventContext")
	}

	props := ctx.Properties()
	if props == nil {
		t.Error("expected Properties() to return empty map for nil EventContext")
	}

	_, ok := ClientAs[string](ctx)
	if ok {
		t.Error("expected ClientAs to return false for nil context")
	}

	_, ok = PayloadAs[string](ctx)
	if ok {
		t.Error("expected PayloadAs to return false for nil context")
	}
}

func TestSetPayload(t *testing.T) {
	event := NewBaseEvent("test", "original")
	if event.Payload() != "original" {
		t.Fatal("expected original payload")
	}

	event.SetPayload("modified")
	if event.Payload() != "modified" {
		t.Fatal("expected modified payload")
	}
}

func TestTypedTopicWrap(t *testing.T) {
	type UserPayload struct {
		ID   string
		Name string
	}

	topic := NewTypedTopic[UserPayload]("user.created")
	payload := UserPayload{ID: "123", Name: "Test"}

	event, err := topic.Wrap(payload)
	if err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}

	if event.Topic() != "user.created" {
		t.Errorf("expected topic user.created, got %s", event.Topic())
	}

	unwrapped, ok := event.Payload().(UserPayload)
	if !ok {
		t.Fatal("expected payload to be UserPayload")
	}

	if unwrapped.ID != "123" {
		t.Errorf("expected ID 123, got %s", unwrapped.ID)
	}
}

func TestUnwrapPayloadJSONDeserialization(t *testing.T) {
	type TestPayload struct {
		Value string `json:"value"`
	}

	jsonData := []byte(`{"value":"test"}`)
	event := NewBaseEvent("test", jsonData)

	decoded, err := UnwrapPayload[TestPayload](event)
	if err != nil {
		t.Fatalf("UnwrapPayload failed: %v", err)
	}

	if decoded.Value != "test" {
		t.Errorf("expected value test, got %s", decoded.Value)
	}
}

func TestUnwrapPayloadErrors(t *testing.T) {
	_, err := UnwrapPayload[string](nil)
	if err != ErrNilPayload {
		t.Errorf("expected ErrNilPayload for nil event, got %v", err)
	}

	event := NewBaseEvent("test", nil)
	_, err = UnwrapPayload[string](event)
	if err != ErrNilPayload {
		t.Errorf("expected ErrNilPayload for nil payload, got %v", err)
	}
}
