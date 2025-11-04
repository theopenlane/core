//go:build examples

package main

import (
	"fmt"

	"github.com/theopenlane/core/pkg/events/soiree"
)

// CustomPanicHandler logs the panic information and performs necessary cleanup
func CustomPanicHandler(recoveredPanic any) {
	fmt.Printf("Recovered from panic: %v", recoveredPanic)
	// Additional panic recovery logic can go here
	// For example, you might want to notify an administrator or restart the operation that caused the panic or murder the misbehaving service to set an example to the others
}

func main() {
	// Create a new soiree instance with the custom panic handler
	e := soiree.NewEventPool(soiree.WithPanicHandler(CustomPanicHandler))
	userCreated := soiree.NewEventTopic("user.created")

	// Define an event listener that intentionally causes a panic
	listener := func(_ *soiree.EventContext, evt soiree.Event) error {
		// Simulating a panic situation
		panic(fmt.Sprintf("George Costanza when there's a fire: %s", evt.Topic()))
	}

	// Subscribe the listener to a topic
	if _, err := soiree.OnTopic(e, userCreated, listener); err != nil {
		panic(err)
	}

	// Emit an event which will cause the listener to panic
	// Normally, you would check for errors and handle the error channel
	soiree.EmitTopic(e, userCreated, soiree.Event(soiree.NewBaseEvent(userCreated.Name(), "sfunk")))

	// Assuming there's additional application logic that continues after event emission,
	// it would carry on uninterrupted thanks to our handy panic handler
	fmt.Println("Application continues running despite the panic - isn't programming great")
}
