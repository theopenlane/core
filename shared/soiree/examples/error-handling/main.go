//go:build examples

package main

import (
	"fmt"
	"log"

	"github.com/theopenlane/shared/soiree"
)

func CustomErrorHandler(event soiree.Event, err error) error {
	log.Printf("Error processing event: %s with payload: %v - error: %s\n", event.Topic(), event.Payload(), err.Error())
	return nil
}

func main() {
	pool := soiree.NewEventPool(soiree.WithErrorHandler(CustomErrorHandler))
	topic := "user.created"

	listener := func(ctx *soiree.EventContext) error {
		return fmt.Errorf("simulated error in listener for event: %s", ctx.Event().Topic())
	}

	if _, err := pool.On(topic, listener); err != nil {
		panic(err)
	}

	errChan := pool.Emit(topic, soiree.NewBaseEvent(topic, "Lady Sansa Stark of Winterfell"))
	for err := range errChan {
		if err != nil {
			log.Printf("Error received from error channel: %v", err)
		}
	}
}
