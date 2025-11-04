//go:build examples

package main

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	uuidGenerator := func() string {
		return uuid.NewString()
	}

	pool := soiree.NewEventPool(soiree.WithIDGenerator(uuidGenerator))
	topic := "user.created"

	listener := func(ctx *soiree.EventContext) error {
		fmt.Printf("I have become aware of an event: %s with payload: %+v\n", ctx.Event().Topic(), ctx.Payload())
		return nil
	}

	listenerID, err := pool.On(topic, listener)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listener with ID %s subscribed to topic '%s'\n", listenerID, topic)

	pool.Emit(topic, soiree.NewBaseEvent(topic, "John Snow"))
}
