//go:build examples

package main

import (
	"fmt"
	"time"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Initialize the soiree - invite your friends
	e := soiree.NewEventPool()
	orderCreated := soiree.NewEventTopic("order.created")

	// High-priority listener for order validation
	validateOrderListener := func(evt soiree.Event) error {
		orderID := evt.Payload().(string)
		// Perform validation logic...
		fmt.Printf("Validating order: %s\n", orderID)
		// Simulate order validation failure
		if orderID == "order123" {
			fmt.Println("validation failed - aborting event...")
			evt.SetAborted(true)
		}

		return nil
	}

	// Listener for processing the payment
	processPaymentListener := func(evt soiree.Event) error {
		if evt.IsAborted() {
			fmt.Println("Payment processing skipped due to previous validation failure")
			return nil
		}

		orderID := evt.Payload().(string)
		// Process payment logic...
		fmt.Printf("Processing payment for order: %s\n", orderID)

		return nil
	}

	// Listener for sending confirmation email
	sendConfirmationEmailListener := func(evt soiree.Event) error {
		if evt.IsAborted() {
			fmt.Println("Confirmation email not sent due to event abort")
			return nil
		}

		orderID := evt.Payload().(string)
		// Send email logic...
		fmt.Printf("Sending confirmation email for order: %s\n", orderID)

		return nil
	}

	// Subscribe listeners with specified priorities
	soiree.MustOn(e, orderCreated, soiree.TypedListener[soiree.Event](validateOrderListener), soiree.WithPriority(soiree.Highest))
	soiree.MustOn(e, orderCreated, soiree.TypedListener[soiree.Event](processPaymentListener), soiree.WithPriority(soiree.Normal))
	soiree.MustOn(e, orderCreated, soiree.TypedListener[soiree.Event](sendConfirmationEmailListener), soiree.WithPriority(soiree.Low))

	// Emit events for order creation
	fmt.Println("Emitting event for order creation...")
	soiree.EmitTopic(e, orderCreated, soiree.Event(soiree.NewBaseEvent(orderCreated.Name(), "order123")))
	soiree.EmitTopic(e, orderCreated, soiree.Event(soiree.NewBaseEvent(orderCreated.Name(), "order456")))

	// Allow time for events to be processed
	time.Sleep(1 * time.Second) // Replace with proper synchronization in production
}
