package main

import (
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/redis/go-redis/v9"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	e := soiree.NewEventPool(
		soiree.WithRedisStore(client),
		soiree.WithRetry(3, func() backoff.BackOff {
			return backoff.NewConstantBackOff(500 * time.Millisecond)
		}),
	)

	e.On("task.created", func(evt soiree.Event) error {
		// Handle the event and return an error to trigger retries
		return nil
	})

	e.Emit("task.created", "some payload")
}
