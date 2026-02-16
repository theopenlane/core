//go:build examples

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/theopenlane/core/pkg/middleware/ratelimit"
)

func main() {
	limitedKey := "key"
	windowSize := 1 * time.Minute
	// create map data store for rate limiter and set each element's expiration time to 2*windowSize and old data flush interval to 10*time.Second
	dataStore := ratelimit.NewMapLimitStore(
		context.Background(),
		2*windowSize,
		10*time.Second,
	)

	var maxLimit int64 = 5
	// allow 5 requests per windowSize (1 minute)
	rateLimiter := ratelimit.New(dataStore, maxLimit, windowSize)

	for i := 0; i < 10; i++ {
		limitStatus, err := rateLimiter.Check(limitedKey)
		if err != nil {
			log.Fatal(err)
		}

		if limitStatus.IsLimited {
			fmt.Printf("too high rate for key: %s: rate: %f, limit: %d\nsleep: %s", limitedKey, limitStatus.CurrentRate, maxLimit, *limitStatus.LimitDuration)
			time.Sleep(*limitStatus.LimitDuration)
		} else {
			err := rateLimiter.Inc(limitedKey)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
