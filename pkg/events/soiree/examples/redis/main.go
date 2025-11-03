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

	e := soiree.NewEventPool(
		soiree.WithRedisStore(client),
		soiree.WithRetry(3, func() backoff.BackOff {
			return backoff.NewConstantBackOff(500 * time.Millisecond)
		}),
	)
	taskCreated := soiree.NewEventTopic("task.created")

	soiree.MustOn(e, taskCreated, soiree.TypedListener[soiree.Event](func(evt soiree.Event) error {
		// Handle the event and return an error to trigger retries
		return nil
	}))

	soiree.EmitTopic(e, taskCreated, soiree.Event(soiree.NewBaseEvent(taskCreated.Name(), "some payload")))
}
