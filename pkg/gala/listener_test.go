package gala

import (
	"testing"
)

func TestRegisterTopicAndAttachListeners(t *testing.T) {
	runtime := newTestGala(t, nil)
	topic := Topic[runtimeTestPayload]{Name: TopicName("listener.registration.topic")}

	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	ids := make([]ListenerID, 0, 2)
	for _, definition := range []Definition[runtimeTestPayload]{
		{
			Topic: topic,
			Name:  "listener.registration.one",
			Handle: func(HandlerContext, runtimeTestPayload) error {
				return nil
			},
		},
		{
			Topic: topic,
			Name:  "listener.registration.two",
			Handle: func(HandlerContext, runtimeTestPayload) error {
				return nil
			},
		},
	} {
		id, err := attachListener(runtime, definition)
		if err != nil {
			t.Fatalf("unexpected listener registration error: %v", err)
		}

		ids = append(ids, id)
	}

	if len(ids) != 2 {
		t.Fatalf("expected two listener ids, got %d", len(ids))
	}

	if got := len(runtime.registry.registeredListeners(topic.Name)); got != 2 {
		t.Fatalf("expected two listeners attached, got %d", got)
	}
}

func TestRegisterTopicWithJSONCodecEncodesAndDecodes(t *testing.T) {
	runtime := newTestGala(t, nil)
	topic := Topic[runtimeTestPayload]{Name: TopicName("listener.registration.json_codec")}

	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "listener.registration.json_codec",
		Handle: func(HandlerContext, runtimeTestPayload) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("unexpected listener registration error: %v", err)
	}

	registration, err := runtime.registry.topicRegistration(topic.Name)
	if err != nil {
		t.Fatalf("expected topic registration to resolve: %v", err)
	}

	encoded, err := registration.encode(runtimeTestPayload{Message: "hello"})
	if err != nil {
		t.Fatalf("expected payload to encode with json codec: %v", err)
	}

	decoded, err := registration.decode(encoded)
	if err != nil {
		t.Fatalf("expected payload to decode with json codec: %v", err)
	}

	payload, ok := decoded.(runtimeTestPayload)
	if !ok {
		t.Fatalf("expected decoded payload type %T, got %T", runtimeTestPayload{}, decoded)
	}
	if payload.Message != "hello" {
		t.Fatalf("expected decoded message %q, got %q", "hello", payload.Message)
	}
}

func TestRegisterRegistersTopicAndListener(t *testing.T) {
	runtime := newTestGala(t, nil)
	topic := Topic[runtimeTestPayload]{Name: TopicName("listener.registration.durable")}

	ids, err := Register(runtime, Definition[runtimeTestPayload]{
		Topic:  topic,
		Name:   "listener.registration.durable",
		Handle: func(HandlerContext, runtimeTestPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("unexpected durable listener registration error: %v", err)
	}

	if len(ids) != 1 {
		t.Fatalf("expected one listener id, got %d", len(ids))
	}

	if _, err := runtime.registry.topicRegistration(topic.Name); err != nil {
		t.Fatalf("expected topic to be registered, got %v", err)
	}

	if got := len(runtime.registry.registeredListeners(topic.Name)); got != 1 {
		t.Fatalf("expected one listener attached, got %d", got)
	}
}
