package gala

import (
	"testing"
)

type listenerRegistrationTestPayload struct {
	Message string `json:"message"`
}

func TestRegisterListenersRegistersTopicAndListeners(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[listenerRegistrationTestPayload]{Name: TopicName("listener.registration.topic")}

	ids, err := RegisterListeners(registry,
		ListenerRegistration[listenerRegistrationTestPayload]{
			Topic: topic,
			Name:  "listener.registration.one",
			Policy: TopicPolicy{
				EmitMode:   EmitModeDurable,
				QueueClass: QueueClassGeneral,
			},
			Handle: func(HandlerContext, listenerRegistrationTestPayload) error { return nil },
		},
		ListenerRegistration[listenerRegistrationTestPayload]{
			Topic:  topic,
			Name:   "listener.registration.two",
			Handle: func(HandlerContext, listenerRegistrationTestPayload) error { return nil },
		},
	)
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	if len(ids) != 2 {
		t.Fatalf("expected two listener ids, got %d", len(ids))
	}

	policy, ok := registry.TopicPolicy(topic.Name)
	if !ok {
		t.Fatalf("expected topic policy to be registered")
	}
	if policy.QueueClass != QueueClassGeneral {
		t.Fatalf("expected queue class %q, got %q", QueueClassGeneral, policy.QueueClass)
	}

	if got := len(registry.Listeners(topic.Name)); got != 2 {
		t.Fatalf("expected two listeners attached, got %d", got)
	}
}

func TestRegisterListenerDefaultsCodecToJSON(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[listenerRegistrationTestPayload]{Name: TopicName("listener.registration.default_codec")}

	_, err := RegisterListener(registry, ListenerRegistration[listenerRegistrationTestPayload]{
		Topic:  topic,
		Name:   "listener.registration.default_codec",
		Handle: func(HandlerContext, listenerRegistrationTestPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	encoded, _, err := registry.EncodePayload(topic.Name, listenerRegistrationTestPayload{Message: "hello"})
	if err != nil {
		t.Fatalf("expected payload to encode with default codec: %v", err)
	}

	decoded, err := registry.DecodePayload(topic.Name, encoded)
	if err != nil {
		t.Fatalf("expected payload to decode with default codec: %v", err)
	}

	payload, ok := decoded.(listenerRegistrationTestPayload)
	if !ok {
		t.Fatalf("expected decoded payload type %T, got %T", listenerRegistrationTestPayload{}, decoded)
	}
	if payload.Message != "hello" {
		t.Fatalf("expected decoded message %q, got %q", "hello", payload.Message)
	}
}

func TestRegisterDurableListenersAppliesQueueClass(t *testing.T) {
	registry := NewRegistry()
	topic := Topic[listenerRegistrationTestPayload]{Name: TopicName("listener.registration.durable")}

	ids, err := RegisterDurableListeners(registry, QueueClassWorkflow, Definition[listenerRegistrationTestPayload]{
		Topic:  topic,
		Name:   "listener.registration.durable",
		Handle: func(HandlerContext, listenerRegistrationTestPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	if len(ids) != 1 {
		t.Fatalf("expected one listener id, got %d", len(ids))
	}

	policy, ok := registry.TopicPolicy(topic.Name)
	if !ok {
		t.Fatalf("expected topic policy to be registered")
	}
	if policy.EffectiveEmitMode() != EmitModeDurable {
		t.Fatalf("expected emit mode %q, got %q", EmitModeDurable, policy.EffectiveEmitMode())
	}
	if policy.QueueClass != QueueClassWorkflow {
		t.Fatalf("expected queue class %q, got %q", QueueClassWorkflow, policy.QueueClass)
	}
}
