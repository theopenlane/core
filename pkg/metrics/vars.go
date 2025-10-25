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

	GraphQLQueryRejected = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_graphql_query_rejected_total",
		Help: "The total number of GraphQL queries rejected",
	}, []string{"reason"})

	GraphQLQueryComplexity = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "openlane_graphql_query_complexity",
		Help:    "The complexity of GraphQL queries by operation",
		Buckets: []float64{5, 10, 25, 50, 100, 200, 500, 1000},
	}, []string{"operation"})

	FileUploadsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openlane_file_uploads_in_flight",
		Help: "Current number of file uploads in progress",
	})

	FileUploadsStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "openlane_file_uploads_started_total",
		Help: "Total number of file uploads started",
	})

	FileUploadsCompleted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_file_uploads_completed_total",
		Help: "Total number of file uploads completed",
	}, []string{"status"})

	FileUploadDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "openlane_file_upload_duration_seconds",
		Help:    "Duration of file uploads in seconds",
		Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0},
	}, []string{"status"})

	FileBufferingStrategy = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_file_buffering_strategy_total",
		Help: "Total number of files buffered by strategy (memory vs disk)",
	}, []string{"strategy"})

	StorageProviderUploads = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_uploads_total",
		Help: "Total number of files uploaded per storage provider",
	}, []string{"provider"})

	StorageProviderBytesUploaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_bytes_uploaded_total",
		Help: "Total bytes uploaded per storage provider",
	}, []string{"provider"})

	StorageProviderDownloads = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_downloads_total",
		Help: "Total number of files downloaded per storage provider",
	}, []string{"provider"})

	StorageProviderBytesDownloaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_bytes_downloaded_total",
		Help: "Total bytes downloaded per storage provider",
	}, []string{"provider"})

	StorageProviderDeletes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_deletes_total",
		Help: "Total number of files deleted per storage provider",
	}, []string{"provider"})

	AuthenticationAttempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_authentication_attempts_total",
		Help: "The total number of authentication attempts by type (jwt, jwt_anonymous, pat, api_token)",
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
		GraphQLQueryRejected,
		GraphQLQueryComplexity,
		RequestValidations,
		HandlerErrors,
		HandlerResults,
		FileUploadsInFlight,
		FileUploadsStarted,
		FileUploadsCompleted,
		FileUploadDuration,
		FileBufferingStrategy,
		StorageProviderUploads,
		StorageProviderBytesUploaded,
		StorageProviderDownloads,
		StorageProviderBytesDownloaded,
		StorageProviderDeletes,
		AuthenticationAttempts,
	}

	QueueConsumerMetrics = []prometheus.Collector{
		QueueTasksProcessedDuration,
		QueueTasksProcessed,
		QueueTasksProcessedErrors,
	}
)
