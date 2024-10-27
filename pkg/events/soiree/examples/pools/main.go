package main

import (
	"fmt"
	"time"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Initialize a goroutine pool with 5 workers and a maximum capacity of 1000 tasks
	pool := soiree.NewPondPool(100)

	// Create a new soiree instance using the custom pool
	e := soiree.NewEventPool(soiree.WithPool(pool))

	// Define a listener that simulates a time-consuming task - dealing with humans usually
	userSignupListener := func(evt soiree.Event) error {
		fmt.Printf("Processing event: %s with payload: %v\n", evt.Topic(), evt.Payload())
		// Simulate some work with a sleep
		time.Sleep(2 * time.Second)
		fmt.Printf("Finished processing event: %s\n", evt.Topic())

		return nil
	}

	// Subscribe a listener to a topic
	e.On("user.signup", userSignupListener)

	// Emit several events concurrently
	for i := 0; i < 100; i++ {
		go func(index int) {
			payload := fmt.Sprintf("User #%d", index)
			e.Emit("user.signup", payload)
		}(i)
	}

	// Wait for all events to be processed before shutting down
	time.Sleep(10 * time.Second)

	// Release the resources used by the pool
	pool.Release()
	fmt.Println("All events have been processed and the pool has been released")
}
