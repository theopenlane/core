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
	validateOrderListener := func(_ *soiree.EventContext, evt soiree.Event) error {
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
	processPaymentListener := func(_ *soiree.EventContext, evt soiree.Event) error {
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
	sendConfirmationEmailListener := func(_ *soiree.EventContext, evt soiree.Event) error {
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
	if _, err := soiree.OnTopic(e, orderCreated, validateOrderListener, soiree.WithPriority(soiree.Highest)); err != nil {
		panic(err)
	}
	if _, err := soiree.OnTopic(e, orderCreated, processPaymentListener, soiree.WithPriority(soiree.Normal)); err != nil {
		panic(err)
	}
	if _, err := soiree.OnTopic(e, orderCreated, sendConfirmationEmailListener, soiree.WithPriority(soiree.Low)); err != nil {
		panic(err)
	}

	// Emit events for order creation
	fmt.Println("Emitting event for order creation...")
	soiree.EmitTopic(e, orderCreated, soiree.Event(soiree.NewBaseEvent(orderCreated.Name(), "order123")))
	soiree.EmitTopic(e, orderCreated, soiree.Event(soiree.NewBaseEvent(orderCreated.Name(), "order456")))

	// Allow time for events to be processed
	time.Sleep(1 * time.Second) // Replace with proper synchronization in production
}
