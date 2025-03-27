package main

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Custom ID generator using UUID v4
	uuidGenerator := func() string {
		return uuid.NewString()
	}

	// Start a soiree but invite UUID instead of ULID
	e := soiree.NewEventPool(soiree.WithIDGenerator(uuidGenerator))

	// Define an event listener
	listener := func(evt soiree.Event) error {
		// The listener does something with the event
		fmt.Printf("I have become aware of an event: %s with payload: %+v\n", evt.Topic(), evt.Payload())
		return nil
	}

	// Subscribe the listener to a topic and retrieve the listener's unique ID
	listenerID, err := e.On("user.created", listener)
	if err != nil {
		fmt.Printf("Error subscribing listener: %v\n", err)
		return
	}

	// The listenerID returned from the subscription is the unique UUID generated for the listener
	fmt.Printf("Listener with ID %s subscribed to topic 'user.created'\n", listenerID)

	// Emit an event
	e.Emit("user.created", "John Snow")
}
