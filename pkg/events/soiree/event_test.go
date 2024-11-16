package soiree

import (
	"reflect"
	"testing"
)

func TestNewBaseEvent(t *testing.T) {
	payload := map[string]string{"key": "value"} // Payload is a map
	event := NewBaseEvent("test_topic", payload)

	if event.Topic() != "test_topic" {
		t.Errorf("NewBaseEvent() topic = %s; want test_topic", event.Topic())
	}

	retrievedPayload, ok := event.Payload().(map[string]string)
	if !ok {
		t.Fatalf("Payload is not of type map[string]string")
	}

	if retrievedPayload["key"] != "value" {
		t.Errorf("NewBaseEvent() payload = %v; want %v", event.Payload(), payload)
	}
}

func TestBaseEventSetAbortedAndIsAborted(t *testing.T) {
	type Payload struct {
		Data string
	}

	event := NewBaseEvent("test_topic", Payload{Data: "some data"}) // Payload is a struct

	if event.IsAborted() {
		t.Errorf("Newly created event should not be aborted")
	}

	event.SetAborted(true)

	if !event.IsAborted() {
		t.Errorf("BaseEvent.Abort(true) did not abort the event")
	}

	event.SetAborted(false)

	if event.IsAborted() {
		t.Errorf("BaseEvent.Abort(false) did not unabort the event")
	}
}

func TestPropertiesSimple(t *testing.T) {
	text := "ABC"
	number := 0.5

	tests := map[string]struct {
		ref Properties
		run func(Properties)
	}{
		"revenue":  {Properties{"revenue": number}, func(p Properties) { p.Set("revenue", number) }},
		"currency": {Properties{"currency": text}, func(p Properties) { p.Set("currency", text) }},
	}

	for name, test := range tests {
		prop := NewProperties()
		test.run(prop)

		if !reflect.DeepEqual(prop, test.ref) {
			t.Errorf("%s: invalid properties produced: %#v\n", name, prop)
		}
	}
}

func TestPropertiesMulti(t *testing.T) {
	p0 := Properties{"title": "A", "value": 0.5}
	p1 := NewProperties().Set("title", "A").Set("value", 0.5)

	if !reflect.DeepEqual(p0, p1) {
		t.Errorf("invalid properties produced by chained setters:\n- expected %#v\n- found: %#v", p0, p1)
	}
}

func TestEventProperties(t *testing.T) {
	event := NewTestEvent("test.event", "test payload")
	event.Properties().Set("key", "value")

	if event.Properties().GetKey("key") != "value" {
		t.Errorf("expected property 'key' to be 'value', got '%v'", event.Properties().GetKey("key"))
	}
}

func TestEventPropertiesSet(t *testing.T) {
	event := NewTestEvent("test.event", "test payload")
	event.SetProperties(Properties{"key": "value"})

	if event.Properties().GetKey("key") != "value" {
		t.Errorf("expected property 'key' to be 'value', got '%v'", event.Properties().GetKey("key"))
	}
}

func TestEventPropertiesSetNil(t *testing.T) {
	event := NewTestEvent("test.event", "test payload")
	event.SetProperties(nil)

	if event.Properties() == nil {
		t.Errorf("expected properties to be initialized")
	}
}

func TestEventPropertiesSetNilMap(t *testing.T) {
	event := NewTestEvent("test.event", "test payload")
	event.SetProperties(Properties(nil))

	if event.Properties() == nil {
		t.Errorf("expected properties to be initialized")
	}
}
