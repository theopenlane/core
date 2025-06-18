package soiree

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics returns a metrics handler that registers default Prometheus collectors for the given pool
func PrometheusMetrics() MetricsHandler {
	return func(pl Pool) {
		p, ok := pl.(*PondPool)
		if !ok {
			return
		}

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
}

// NewStatsCollector is retained for backward compatibility and will register
// the default Prometheus metrics handler with the pool
func (p *PondPool) NewStatsCollector() {
	p.RegisterMetricsHandler(PrometheusMetrics())
}
