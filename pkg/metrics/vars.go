package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// These variables are stubs for now ad will be added throughout the codebase
var (
	WorkerExecutions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_worker_executions_total",
		Help: "The total number of worker executions completed",
	}, []string{"name"})

	WorkerExecutionErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_worker_execution_errors_total",
		Help: "The total number of worker execution errors",
	}, []string{"name"})

	Logins = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_logins_total",
		Help: "The total number of login attempts",
	}, []string{"success"})

	RequestValidations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_request_validations_total",
		Help: "The total number of request validations by type and success",
	}, []string{"request", "success"})

	HandlerErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_handler_errors_total",
		Help: "The total number of handler errors by function and status code",
	}, []string{"function", "code"})

	HandlerResults = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_handler_results_total",
		Help: "The total number of handler results by request type and success",
	}, []string{"request", "success"})

	Registrations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "openlane_registrations_total",
		Help: "The total number of user registrations",
	})

	EmailValidations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_email_validations_total",
		Help: "The total number of email validations by success and result type",
	}, []string{"success", "result"})

	QueueTasksPushed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_pushed_total",
		Help: "The total number of tasks pushed to queue",
	}, []string{"queue", "task"})

	QueueTasksPushFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_push_failures_total",
		Help: "The total number of errors pushing a task to the queue",
	}, []string{"queue", "task"})

	QueueTasksProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_processed_total",
		Help: "The total number of tasks processed by the consumer",
	}, []string{"task"})

	QueueTasksProcessedErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_process_errors_total",
		Help: "The total number of errors encountered by the consumer",
	}, []string{"task"})

	QueueTasksProcessedDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "openlane_queue_tasks_process_duration_seconds",
		Help: "The duration of task processing in seconds",
	}, []string{"task"})

	GraphQLOperationTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_graphql_operations_total",
		Help: "The total number of GraphQL operations processed",
	}, []string{"operation", "success"})

	GraphQLOperationDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "openlane_graphql_operation_duration_seconds",
		Help:    "The duration of GraphQL operations in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	// Impersonations counts impersonation events by action type (start/end)
	Impersonations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_impersonations_total",
		Help: "The total number of impersonation events by action type",
	}, []string{"type"})

	APIMetrics = []prometheus.Collector{
		WorkerExecutions,
		WorkerExecutionErrors,
		Logins,
		Registrations,
		EmailValidations,
		QueueTasksPushed,
		QueueTasksPushFailures,
		GraphQLOperationTotal,
		GraphQLOperationDuration,
		Impersonations,
		RequestValidations,
		HandlerErrors,
		HandlerResults,
	}

	QueueConsumerMetrics = []prometheus.Collector{
		QueueTasksProcessedDuration,
		QueueTasksProcessed,
		QueueTasksProcessedErrors,
	}
)
