package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// These variables are stubs for now ad will be added throughout the codebase
var (
	WorkerExecutions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_executions_count",
		Help: "The number of worker executions completed",
	}, []string{"name"})
	WorkerExecutionErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_execution_errors_count",
		Help: "The number of worker execution errors",
	}, []string{"name"})
	Logins = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "logins_count",
		Help: "The number of logins",
	}, []string{"success"})

	Registrations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registrations_count",
		Help: "The number of registrations",
	})
	QueueTasksPushed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_pushed_total",
		Help: "The number of tasks pushed to queue",
	}, []string{"queue", "task"})
	QueueTasksPushFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_pushed_failures_total",
		Help: "The number of errors pushing a task to the queue",
	}, []string{"queue", "task"})

	QueueTasksProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_processed_total",
		Help: "The number of tasks processed by the consumer",
	}, []string{"task"})
	QueueTasksProcessedErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_processed_errors_total",
		Help: "The number of errors encountered by the consumer",
	}, []string{"task"})
	QueueTasksProcessedDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "queue_tasks_processed_duration_seconds",
		Help: "The length of time taken for a task to be processed in seconds",
	}, []string{"task"})

	APIMetrics = []prometheus.Collector{
		WorkerExecutions,
		WorkerExecutionErrors,
		Logins,
		Registrations,
		QueueTasksPushed,
		QueueTasksPushFailures,
	}

	QueueConsumerMetrics = []prometheus.Collector{
		QueueTasksProcessedDuration,
		QueueTasksProcessed,
		QueueTasksProcessedErrors,
	}
)
