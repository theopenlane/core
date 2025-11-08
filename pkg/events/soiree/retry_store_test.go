package soiree

import (
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
)

func TestEventPersistence(t *testing.T) {
	store := NewInMemoryStore()
	soiree := NewEventPool(WithEventStore(store))

	topic := typedEventTopic("topic")
	if _, err := BindListener(topic, func(_ *EventContext, e Event) error { return nil }).Register(soiree); err != nil {
		t.Fatalf("On() error: %v", err)
	}

	soiree.EmitSync(topic.Name(), NewBaseEvent(topic.Name(), "payload"))

	if len(store.Events()) == 0 {
		t.Fatalf("expected event to be stored")
	}

	if len(store.Results()) == 0 {
		t.Fatalf("expected handler result to be stored")
	}
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	soiree := NewEventPool(
		WithRetry(3, func() backoff.BackOff { return backoff.NewConstantBackOff(10 * time.Millisecond) }),
	)

	topic := typedEventTopic("topic")
	if _, err := BindListener(topic, func(_ *EventContext, e Event) error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	}).Register(soiree); err != nil {
		t.Fatalf("On() error: %v", err)
	}

	errs := soiree.EmitSync(topic.Name(), NewBaseEvent(topic.Name(), "data"))
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
