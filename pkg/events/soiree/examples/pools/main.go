//go:build examples

package main

import (
	"fmt"
	"time"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	pool := soiree.NewPondPool(soiree.WithMaxWorkers(5))

	e := soiree.NewEventPool(soiree.WithPool(pool))
	topic := "user.signup"

	userSignupListener := func(ctx *soiree.EventContext) error {
		fmt.Printf("Processing event: %s with payload: %v\n", ctx.Event().Topic(), ctx.Payload())
		time.Sleep(2 * time.Second)
		fmt.Printf("Finished processing event: %s\n", ctx.Event().Topic())
		return nil
	}

	if _, err := e.On(topic, userSignupListener); err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		go func(index int) {
			payload := fmt.Sprintf("User #%d", index)
			e.Emit(topic, soiree.NewBaseEvent(topic, payload))
		}(i)
	}

	time.Sleep(10 * time.Second)

	pool.Release()
	fmt.Println("All events have been processed and the pool has been released")
}
