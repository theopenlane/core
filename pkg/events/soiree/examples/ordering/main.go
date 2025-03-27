package main

import (
	"fmt"
	"time"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Initialize the soiree - invite your friends
	e := soiree.NewEventPool()

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
	e.On("order.created", validateOrderListener, soiree.WithPriority(soiree.Highest))
	e.On("order.created", processPaymentListener, soiree.WithPriority(soiree.Normal))
	e.On("order.created", sendConfirmationEmailListener, soiree.WithPriority(soiree.Low))

	// Emit events for order creation
	fmt.Println("Emitting event for order creation...")
	e.Emit("order.created", "order123") // This order will fail validation
	e.Emit("order.created", "order456") // This order will pass validation

	// Allow time for events to be processed
	time.Sleep(1 * time.Second) // Replace with proper synchronization in production
}
