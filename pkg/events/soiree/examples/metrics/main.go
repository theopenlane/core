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
	pool := soiree.NewPondPool(soiree.WithMaxWorkers(100))

	e := soiree.NewEventPool(soiree.WithPool(pool))
	topic := "user.signup"

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pool_workers_running",
			Help: "Number of running worker goroutines",
		},
		func() float64 { return float64(pool.Running()) },
	))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_submitted_total",
			Help: "Number of tasks submitted",
		},
		func() float64 { return float64(pool.SubmittedTasks()) },
	))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pool_tasks_waiting_total",
			Help: "Number of tasks waiting in the queue",
		},
		func() float64 { return float64(pool.WaitingTasks()) },
	))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_successful_total",
			Help: "Number of tasks that completed successfully",
		},
		func() float64 { return float64(pool.SuccessfulTasks()) },
	))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_failed_total",
			Help: "Number of tasks that completed with panic",
		},
		func() float64 { return float64(pool.FailedTasks()) },
	))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_completed_total",
			Help: "Number of tasks that completed either successfully or with panic",
		},
		func() float64 { return float64(pool.CompletedTasks()) },
	))

	http.Handle("/metrics", promhttp.Handler())

	userSignupListener := func(ctx *soiree.EventContext) error {
		fmt.Printf("Processing event: %s with payload: %v\n", ctx.Event().Topic(), ctx.Payload())
		time.Sleep(2 * time.Second)
		fmt.Printf("Finished processing event: %s\n", ctx.Event().Topic())
		return nil
	}

	if _, err := e.On(topic, userSignupListener); err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		go func(index int) {
			payload := fmt.Sprintf("User #%d", index)
			e.Emit(topic, soiree.NewBaseEvent(topic, payload))
		}(i)
	}

	defer pool.Release()

	if err := http.ListenAndServe(":8084", nil); err != nil {
		panic(err)
	}
}
