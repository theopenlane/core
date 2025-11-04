//go:build examples

package main

import (
	"fmt"
	"time"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	pool := soiree.NewEventPool()
	topic := "order.created"

	validateOrder := func(ctx *soiree.EventContext) error {
		orderID := ctx.Payload().(string)
		fmt.Printf("Validating order: %s\n", orderID)
		if orderID == "order123" {
			fmt.Println("validation failed - aborting event...")
			ctx.Event().SetAborted(true)
		}

		return nil
	}

	processPayment := func(ctx *soiree.EventContext) error {
		if ctx.Event().IsAborted() {
			fmt.Println("Payment processing skipped due to previous validation failure")
			return nil
		}

		orderID := ctx.Payload().(string)
		fmt.Printf("Processing payment for order: %s\n", orderID)

		return nil
	}

	sendEmail := func(ctx *soiree.EventContext) error {
		if ctx.Event().IsAborted() {
			fmt.Println("Confirmation email not sent due to event abort")
			return nil
		}

		orderID := ctx.Payload().(string)
		fmt.Printf("Sending confirmation email for order: %s\n", orderID)

		return nil
	}

	if _, err := pool.On(topic, validateOrder, soiree.WithPriority(soiree.Highest)); err != nil {
		panic(err)
	}
	if _, err := pool.On(topic, processPayment, soiree.WithPriority(soiree.Normal)); err != nil {
		panic(err)
	}
	if _, err := pool.On(topic, sendEmail, soiree.WithPriority(soiree.Low)); err != nil {
		panic(err)
	}

	fmt.Println("Emitting event for order creation...")
	pool.Emit(topic, soiree.NewBaseEvent(topic, "order123"))
	pool.Emit(topic, soiree.NewBaseEvent(topic, "order456"))

	time.Sleep(time.Second)
}
