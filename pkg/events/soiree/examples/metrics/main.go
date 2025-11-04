//go:build examples

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/theopenlane/core/pkg/events/soiree"
)

func main() {
	// Initialize a goroutine pool with 5 workers and a maximum capacity of 1000 tasks
	pool := soiree.NewPondPool(soiree.WithMaxWorkers(100))

	// Create a new soiree instance using the custom pool
	e := soiree.NewEventPool(soiree.WithPool(pool))
	userSignup := soiree.NewEventTopic("user.signup")

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pool_workers_running",
			Help: "Number of running worker goroutines",
		},
		func() float64 {
			return float64(pool.Running())
		}))

	// Task metrics
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_submitted_total",
			Help: "Number of tasks submitted",
		},
		func() float64 {
			return float64(pool.SubmittedTasks())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pool_tasks_waiting_total",
			Help: "Number of tasks waiting in the queue",
		},
		func() float64 {
			return float64(pool.WaitingTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_successful_total",
			Help: "Number of tasks that completed successfully",
		},
		func() float64 {
			return float64(pool.SuccessfulTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_failed_total",
			Help: "Number of tasks that completed with panic",
		},
		func() float64 {
			return float64(pool.FailedTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_completed_total",
			Help: "Number of tasks that completed either successfully or with panic",
		},
		func() float64 {
			return float64(pool.CompletedTasks())
		}))

	// Expose the registered metrics via HTTP
	http.Handle("/metrics", promhttp.Handler())

	// Define a listener that simulates a time-consuming task - dealing with humans usually
	userSignupListener := func(_ *soiree.EventContext, evt soiree.Event) error {
		fmt.Printf("Processing event: %s with payload: %v\n", evt.Topic(), evt.Payload())
		// Simulate some work with a sleep
		time.Sleep(2 * time.Second)
		fmt.Printf("Finished processing event: %s\n", evt.Topic())

		return nil
	}

	// Subscribe a listener to a topic
	if _, err := soiree.OnTopic(e, userSignup, userSignupListener); err != nil {
		panic(err)
	}

	// Emit several events concurrently
	for i := 0; i < 100; i++ {
		go func(index int) {
			payload := fmt.Sprintf("User #%d", index)
			soiree.EmitTopic(e, userSignup, soiree.Event(soiree.NewBaseEvent(userSignup.Name(), payload)))
		}(i)
	}

	http.ListenAndServe(":8084", nil)

	// Release the resources used by the pool
	defer pool.Release()

	fmt.Println("All events have been processed and the pool has been released")
}
