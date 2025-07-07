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

	_, err := soiree.On("topic", func(e Event) error { return nil })
	if err != nil {
		t.Fatalf("On() error: %v", err)
	}

	soiree.EmitSync("topic", "payload")

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

	_, err := soiree.On("topic", func(e Event) error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("On() error: %v", err)
	}

	errs := soiree.EmitSync("topic", "data")
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
