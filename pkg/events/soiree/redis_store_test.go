package soiree

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	t.Cleanup(mr.Close)

	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func TestRedisEventPersistence(t *testing.T) {
	client := newTestRedis(t)
	store := NewRedisStore(client)
	soiree := NewEventPool(WithEventStore(store))

	done := make(chan struct{}, 1)
	_, err := soiree.On("topic", func(e Event) error {
		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("On() error: %v", err)
	}

	soiree.Emit("topic", "data")

	select {
	case <-done:
		// Give the result time to persist
		time.Sleep(10 * time.Millisecond)
		// Now safe to check for results
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

	soiree := NewEventPool(
		WithEventStore(store),
		WithRetry(2, func() backoff.BackOff { return backoff.NewConstantBackOff(10 * time.Millisecond) }),
	)

	attempts := 0
	_, err := soiree.On("topic", func(e Event) error {
		attempts++
		if attempts < 2 {
			return errors.New("fail")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("On() error: %v", err)
	}

	soiree.Emit("topic", "data")

	time.Sleep(100 * time.Millisecond)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRedisMetrics(t *testing.T) {
	client := newTestRedis(t)

	// Use a custom Prometheus registry for this test
	reg := prometheus.NewRegistry()
	metrics := newRedisMetrics(reg)
	store := NewRedisStoreWithMetrics(client, metrics)
	soiree := NewEventPool(WithEventStore(store))

	_, err := soiree.On("topic", func(e Event) error { return nil })
	if err != nil {
		t.Fatalf("On() error: %v", err)
	}

	soiree.Emit("topic", "data")

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
