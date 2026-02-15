package gala

import (
	"testing"
)

type listenerRegistrationTestPayload struct {
	Message string `json:"message"`
}

func TestRegisterTopicAndAttachListeners(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[listenerRegistrationTestPayload]{Name: TopicName("listener.registration.topic")}

	if err := RegisterTopic(registry, Registration[listenerRegistrationTestPayload]{
		Topic: topic,
		Codec: JSONCodec[listenerRegistrationTestPayload]{},
	}); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	ids := make([]ListenerID, 0, 2)
	for _, definition := range []Definition[listenerRegistrationTestPayload]{
		{
			Topic: topic,
			Name:  "listener.registration.one",
			Handle: func(HandlerContext, listenerRegistrationTestPayload) error {
				return nil
			},
		},
		{
			Topic: topic,
			Name:  "listener.registration.two",
			Handle: func(HandlerContext, listenerRegistrationTestPayload) error {
				return nil
			},
		},
	} {
		id, err := AttachListener(registry, definition)
		if err != nil {
			t.Fatalf("unexpected listener registration error: %v", err)
		}

		ids = append(ids, id)
	}

	if len(ids) != 2 {
		t.Fatalf("expected two listener ids, got %d", len(ids))
	}

	if got := len(registry.registeredListeners(topic.Name)); got != 2 {
		t.Fatalf("expected two listeners attached, got %d", got)
	}
}

func TestRegisterTopicWithJSONCodecEncodesAndDecodes(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[listenerRegistrationTestPayload]{Name: TopicName("listener.registration.json_codec")}

	if err := RegisterTopic(registry, Registration[listenerRegistrationTestPayload]{
		Topic: topic,
		Codec: JSONCodec[listenerRegistrationTestPayload]{},
	}); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	if _, err := AttachListener(registry, Definition[listenerRegistrationTestPayload]{
		Topic: topic,
		Name:  "listener.registration.json_codec",
		Handle: func(HandlerContext, listenerRegistrationTestPayload) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("unexpected listener registration error: %v", err)
	}

	encoded, err := registry.EncodePayload(topic.Name, listenerRegistrationTestPayload{Message: "hello"})
	if err != nil {
		t.Fatalf("expected payload to encode with json codec: %v", err)
	}

	decoded, err := registry.DecodePayload(topic.Name, encoded)
	if err != nil {
		t.Fatalf("expected payload to decode with json codec: %v", err)
	}

	payload, ok := decoded.(listenerRegistrationTestPayload)
	if !ok {
		t.Fatalf("expected decoded payload type %T, got %T", listenerRegistrationTestPayload{}, decoded)
	}
	if payload.Message != "hello" {
		t.Fatalf("expected decoded message %q, got %q", "hello", payload.Message)
	}
}

func TestRegisterListenersRegistersTopicAndListener(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[listenerRegistrationTestPayload]{Name: TopicName("listener.registration.durable")}

	ids, err := RegisterListeners(registry, Definition[listenerRegistrationTestPayload]{
		Topic:  topic,
		Name:   "listener.registration.durable",
		Handle: func(HandlerContext, listenerRegistrationTestPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("unexpected durable listener registration error: %v", err)
	}

	if len(ids) != 1 {
		t.Fatalf("expected one listener id, got %d", len(ids))
	}

	if _, err := registry.EncodePayload(topic.Name, listenerRegistrationTestPayload{Message: "registered"}); err != nil {
		t.Fatalf("expected topic to be registered, got %v", err)
	}

	if got := len(registry.registeredListeners(topic.Name)); got != 1 {
		t.Fatalf("expected one listener attached, got %d", got)
	}
}
