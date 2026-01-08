package soiree

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/cenkalti/backoff/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/redis/go-redis/v9"
)

func mustRegisterListener(t *testing.T, bus *EventBus, topic TypedTopic[Event], listener TypedListener[Event]) string {
	t.Helper()

	id, err := BindListener(topic, listener).Register(bus)
	if err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	return id
}

func newTestRedis(t *testing.T) *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		t.Skipf("skipping redis-backed tests: %v", err)
	}

	t.Cleanup(mr.Close)

	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func TestRedisEventPersistence(t *testing.T) {
	client := newTestRedis(t)
	store := NewRedisStore(client)
	soiree := New(EventStore(store))
	topic := typedEventTopic("topic")

	done := make(chan struct{}, 1)
	mustRegisterListener(t, soiree, topic, func(_ *EventContext, e Event) error {
		done <- struct{}{}
		return nil
	})

	soiree.Emit(topic.Name(), NewBaseEvent(topic.Name(), "data"))

	select {
	case <-done:
		time.Sleep(10 * time.Millisecond)
		evts, err := store.Events(context.Background())
		if err != nil || len(evts) == 0 {
			t.Fatalf("expected stored event, err=%v", err)
		}

		res, err := store.Results(context.Background())
		if err != nil || len(res) == 0 {
			t.Fatalf("expected stored result, err=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("listener did not run")
	}
}

func TestRedisRetryWithBackoff(t *testing.T) {
	client := newTestRedis(t)

	store := NewRedisStore(client)

	soiree := New(
		EventStore(store),
		Retry(2, func() backoff.BackOff { return backoff.NewConstantBackOff(10 * time.Millisecond) }),
	)

	attempts := 0
	topic := typedEventTopic("topic")
	mustRegisterListener(t, soiree, topic, func(_ *EventContext, e Event) error {
		attempts++
		if attempts < 2 {
			return errors.New("fail")
		}
		return nil
	})

	soiree.Emit(topic.Name(), NewBaseEvent(topic.Name(), "data"))

	time.Sleep(100 * time.Millisecond)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRedisMetrics(t *testing.T) {
	client := newTestRedis(t)

	reg := prometheus.NewRegistry()
	metrics := newRedisMetrics(reg)
	store := NewRedisStore(client, WithRedisMetrics(metrics))
	soiree := New(EventStore(store))
	topic := typedEventTopic("topic")

	mustRegisterListener(t, soiree, topic, func(_ *EventContext, e Event) error { return nil })

	soiree.Emit(topic.Name(), NewBaseEvent(topic.Name(), "data"))

	time.Sleep(100 * time.Millisecond)

	if v := testutil.ToFloat64(metrics.redisEventsPersisted); v != 1 {
		t.Fatalf("expected 1 event persisted, got %v", v)
	}
	if v := testutil.ToFloat64(metrics.redisEventsDequeued); v != 1 {
		t.Fatalf("expected 1 event dequeued, got %v", v)
	}
	if v := testutil.ToFloat64(metrics.redisResultsPersisted); v != 1 {
		t.Fatalf("expected 1 result persisted, got %v", v)
	}
}

func TestRedisStoreWithTTLOptions(t *testing.T) {
	client := newTestRedis(t)

	store := NewRedisStore(client,
		WithEventsTTL(time.Hour),
		WithResultsTTL(time.Hour),
		WithDedupTTL(time.Hour),
	)

	if store.eventsTTL != time.Hour {
		t.Fatalf("expected eventsTTL to be 1 hour, got %v", store.eventsTTL)
	}
	if store.resultsTTL != time.Hour {
		t.Fatalf("expected resultsTTL to be 1 hour, got %v", store.resultsTTL)
	}
	if store.dedupTTL != time.Hour {
		t.Fatalf("expected dedupTTL to be 1 hour, got %v", store.dedupTTL)
	}

	soiree := New(EventStore(store))
	topic := typedEventTopic("topic")

	done := make(chan struct{}, 1)
	mustRegisterListener(t, soiree, topic, func(_ *EventContext, e Event) error {
		done <- struct{}{}
		return nil
	})

	soiree.Emit(topic.Name(), NewBaseEvent(topic.Name(), "data"))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener did not run")
	}
}

func TestEmitWithContext(t *testing.T) {
	bus := New()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan context.Context, 1)
	_, err := bus.On("topic", func(ec *EventContext) error {
		done <- ec.Context()
		return nil
	})
	if err != nil {
		t.Fatalf("On() failed: %v", err)
	}

	bus.EmitWithContext(ctx, "topic", "payload")

	select {
	case receivedCtx := <-done:
		if receivedCtx == nil {
			t.Fatal("expected context to be set")
		}
	case <-time.After(time.Second):
		t.Fatal("listener did not run")
	}
}
