package soiree

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func (p *PondPool) NewStatsCollector() {
	// add a default name if none is provided
	name := p.name
	if name == "" {
		name = "default"
	}

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("pool_%s_workers_running", name),
			Help: "Number of running worker goroutines",
		},
		func() float64 {
			return float64(p.Running())
		}))

	// Task metrics
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_submitted_total", name),
			Help: "Number of tasks submitted",
		},
		func() float64 {
			return float64(p.SubmittedTasks())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("pool_%s_tasks_waiting_total", name),
			Help: "Number of tasks waiting in the queue",
		},
		func() float64 {
			return float64(p.WaitingTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_successful_total", name),
			Help: "Number of tasks that completed successfully",
		},
		func() float64 {
			return float64(p.SuccessfulTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_failed_total", name),
			Help: "Number of tasks that completed with panic",
		},
		func() float64 {
			return float64(p.FailedTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_completed_total", name),
			Help: "Number of tasks that completed either successfully or with panic",
		},
		func() float64 {
			return float64(p.CompletedTasks())
		}))
}
