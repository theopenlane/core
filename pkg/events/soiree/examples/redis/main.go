//go:build examples

package main

import (
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/redis/go-redis/v9"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	pool := soiree.NewEventPool(
		soiree.WithRedisStore(client),
		soiree.WithRetry(3, func() backoff.BackOff {
			return backoff.NewConstantBackOff(500 * time.Millisecond)
		}),
	)

	topic := "task.created"

	if _, err := pool.On(topic, func(ctx *soiree.EventContext) error {
		// Handle the event and return an error to trigger retries
		return nil
	}); err != nil {
		panic(err)
	}

	pool.Emit(topic, soiree.NewBaseEvent(topic, "some payload"))
}
