package hooks

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	mutationType = "mutation_type"
	mutationOp   = "mutation_op"
)

var entLabels = []string{mutationType, mutationOp}

// initOpsProcessedTotal creates a collector for total operations counter
func initOpsProcessedTotal() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ent_operation_total",
			Help: "Number of ent mutation operations",
		},
		entLabels,
	)
}

// initOpsProcessedError creates a collector for error counter
func initOpsProcessedError() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ent_operation_error",
			Help: "Number of failed ent mutation operations",
		},
		entLabels,
	)
}

// initOpsDuration creates a collector for duration histogram
func initOpsDuration() *prometheus.HistogramVec {
	return promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "ent_operation_duration_seconds",
			Help: "Time in seconds per operation",
		},
		entLabels,
	)
}

// initOpsTotalDuration creates a collector for total duration counter
func initOpsTotalDuration() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ent_operation_total_duration_seconds",
			Help: "Total time in seconds per operation type and mutation",
		},
		entLabels,
	)
}

// initialize the collectors, prometheus will register them automatically
var (
	opsProcessedTotal = initOpsProcessedTotal()
	opsProcessedError = initOpsProcessedError()
	opsDuration       = initOpsDuration()
	opsTotalDuration  = initOpsTotalDuration()
)

// MetricsHook inits the collectors with count total at beginning, error on mutation error and a duration after the mutation
func MetricsHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Before mutation, start measuring time
			start := time.Now()

			labels := prometheus.Labels{mutationType: m.Type(), mutationOp: m.Op().String()}

			opsProcessedTotal.With(labels).Inc()

			v, err := next.Mutate(ctx, m)
			if err != nil {
				// in case of error, increment the error counter
				opsProcessedError.With(labels).Inc()
			}

			duration := time.Since(start)
			opsDuration.With(labels).Observe(duration.Seconds())
			opsTotalDuration.With(labels).Add(duration.Seconds())

			return v, err
		})
	}
}
