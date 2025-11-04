//go:build examples

package main

import (
	"fmt"
	"log"

	"github.com/theopenlane/core/pkg/events/soiree"
)

// CustomErrorHandler logs and handles errors that occur during event processing
func CustomErrorHandler(event soiree.Event, err error) error {
	// Log the error with additional context or send it to an error tracking service
	log.Printf("Error processing event: %s with payload: %v - error: %s\n", event.Topic(), event.Payload(), err.Error())

	// Here you can decide whether to return the error or handle it so that
	// the soiree considers it resolved
	// Returning nil will effectively 'swallow' the error, indicating it's been handled just like your backlog
	return nil
}

func main() {
	// Create a new soiree instance with the custom error handler
	e := soiree.NewEventPool(soiree.WithErrorHandler(CustomErrorHandler))
	userCreated := soiree.NewEventTopic("user.created")

	// Define an event listener that intentionally causes an error
	listener := func(_ *soiree.EventContext, evt soiree.Event) error {
		// Simulate an error
		return fmt.Errorf("simulated error in listener for event: %s", evt.Topic())
	}

	// Subscribe the listener to a topic
	if _, err := soiree.OnTopic(e, userCreated, listener); err != nil {
		panic(err)
	}

	// Emit an event which will cause the listener to error
	errChan := soiree.EmitTopic(e, userCreated, soiree.Event(soiree.NewBaseEvent(userCreated.Name(), "Lady Sansa Stark of Witerfell")))

	// Wait and collect errors from the error channel
	for err := range errChan {
		if err != nil {
			log.Printf("Error received from error channel: %v", err)
		}
	}
}
