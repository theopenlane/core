package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// These variables are stubs for now ad will be added throughout the codebase
var (
	// WorkerExecutions records the total number of worker executions completed
	// this is currently not being recorded anywhere
	WorkerExecutions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_worker_executions_total",
		Help: "The total number of worker executions completed",
	}, []string{"name"})

	// WorkerExecutionErrors records the total number of worker execution errors
	// this is currently not being recorded anywhere
	WorkerExecutionErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_worker_execution_errors_total",
		Help: "The total number of worker execution errors",
	}, []string{"name"})

	// Logins records the total number of login attempts by success
	Logins = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_logins_total",
		Help: "The total number of login attempts",
	}, []string{"success"})

	// RequestValidations records the total number of request validations by type and success
	RequestValidations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_request_validations_total",
		Help: "The total number of request validations by type and success",
	}, []string{"request", "success"})

	// HandlerErrors records the total number of handler errors by function and status code
	HandlerErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_handler_errors_total",
		Help: "The total number of handler errors by function and status code",
	}, []string{"function", "code"})

	// HandlerResults records the total number of handler results by request type and success
	HandlerResults = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_handler_results_total",
		Help: "The total number of handler results by request type and success",
	}, []string{"request", "success"})

	// Registrations records the total number of user registrations
	Registrations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "openlane_registrations_total",
		Help: "The total number of user registrations",
	})

	// EmailValidations records the number of email validations by success and result type
	EmailValidations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_email_validations_total",
		Help: "The total number of email validations by success and result type",
	}, []string{"success", "result"})

	// QueueTasksPushed records the number of tasks pushed to the queue
	// this is currently not being recorded anywhere
	QueueTasksPushed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_pushed_total",
		Help: "The total number of tasks pushed to queue",
	}, []string{"queue", "task"})

	// QueueTasksPushFailures records the number of errors pushing a task to the queue
	// this is currently not being recorded anywhere
	QueueTasksPushFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_push_failures_total",
		Help: "The total number of errors pushing a task to the queue",
	}, []string{"queue", "task"})

	// QueueTasksProcessed records the number of tasks processed by the consumer
	// this is currently not being recorded anywhere
	QueueTasksProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_processed_total",
		Help: "The total number of tasks processed by the consumer",
	}, []string{"task"})

	// QueueTasksProcessedErrors records the number of errors encountered by the consumer
	// this is currently not being recorded anywhere
	QueueTasksProcessedErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_queue_tasks_process_errors_total",
		Help: "The total number of errors encountered by the consumer",
	}, []string{"task"})

	// QueueTasksProcessedDuration records the duration of task processing in seconds
	// this is currently not being recorded anywhere
	QueueTasksProcessedDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "openlane_queue_tasks_process_duration_seconds",
		Help: "The duration of task processing in seconds",
	}, []string{"task"})

	// GraphQLOperationTotal records the total number of GraphQL operations processed
	GraphQLOperationTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_graphql_operations_total",
		Help: "The total number of GraphQL operations processed",
	}, []string{"operation", "success"})

	// GraphQLOperationDuration records the duration of GraphQL operations in seconds
	GraphQLOperationDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "openlane_graphql_operation_duration_seconds",
		Help:    "The duration of GraphQL operations in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	// GraphQLQueryRejected records the number of rejected GraphQL queries by reason
	GraphQLQueryRejected = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_graphql_query_rejected_total",
		Help: "The total number of GraphQL queries rejected",
	}, []string{"reason"})

	// GraphQLQueryComplexity records the complexity of GraphQL queries by operation
	GraphQLQueryComplexity = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "openlane_graphql_query_complexity",
		Help:    "The complexity of GraphQL queries by operation",
		Buckets: []float64{5, 10, 25, 50, 100, 200, 500, 1000},
	}, []string{"operation"})

	// FileUploadsInFlight records the current number of file uploads in progress
	FileUploadsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openlane_file_uploads_in_flight",
		Help: "Current number of file uploads in progress",
	})

	// FileUploadsStarted records the number of file uploads started
	FileUploadsStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "openlane_file_uploads_started_total",
		Help: "Total number of file uploads started",
	})

	// FileUploadsCompleted records the number of completed file uploads by status
	FileUploadsCompleted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_file_uploads_completed_total",
		Help: "Total number of file uploads completed",
	}, []string{"status"})

	// FileUploadDuration records the duration of file uploads
	FileUploadDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "openlane_file_upload_duration_seconds",
		Help:    "Duration of file uploads in seconds",
		Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0},
	}, []string{"status"})

	// FileBufferingStrategy tracks the number of files buffered by strategy (memory vs disk)
	// this is currently not being recorded anywhere
	FileBufferingStrategy = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_file_buffering_strategy_total",
		Help: "Total number of files buffered by strategy (memory vs disk)",
	}, []string{"strategy"})

	// StorageProviderUploads tracks the number of upload operations per storage provider
	StorageProviderUploads = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_uploads_total",
		Help: "Total number of files uploaded per storage provider",
	}, []string{"provider"})

	// StorageProviderBytesUploaded tracks the total bytes uploaded per storage provider
	StorageProviderBytesUploaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_bytes_uploaded_total",
		Help: "Total bytes uploaded per storage provider",
	}, []string{"provider"})

	// StorageProviderDownloads tracks the number of download operations per storage provider
	StorageProviderDownloads = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_downloads_total",
		Help: "Total number of files downloaded per storage provider",
	}, []string{"provider"})

	// StorageProviderBytesDownloaded tracks the total bytes downloaded per storage provider
	StorageProviderBytesDownloaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_bytes_downloaded_total",
		Help: "Total bytes downloaded per storage provider",
	}, []string{"provider"})

	// StorageProviderDeletes tracks the number of delete operations per storage provider
	StorageProviderDeletes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_storage_provider_deletes_total",
		Help: "Total number of files deleted per storage provider",
	}, []string{"provider"})

	// AuthenticationAttempts tracks the number of authentication attempts by type
	AuthenticationAttempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_authentication_attempts_total",
		Help: "The total number of authentication attempts by type (jwt, jwt_anonymous, pat, api_token)",
	}, []string{"type"})

	// ActiveSubscriptions tracks the current number of active subscriptions
	// this should be the difference between opened and closed subscriptions
	ActiveSubscriptions = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openlane_subscriptions_active",
		Help: "Current number of active subscriptions",
	})

	// SubscriptionOpen tracks the total number of subscriptions opened
	SubscriptionOpen = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "openlane_subscription_open_total",
		Help: "Total number of subscriptions opened",
	})

	// SubscriptionClose tracks the total number of subscriptions closed
	SubscriptionClose = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "openlane_subscription_close_total",
		Help: "Total number of subscriptions closed",
	})

	// SubscriptionLifetime records the lifetime of subscriptions in seconds
	SubscriptionLifetime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "openlane_subscription_lifetime_seconds",
		Help:    "Lifetime of subscriptions in seconds",
		Buckets: prometheus.ExponentialBuckets(10, 2, 10), //nolint:mnd
	})

	// APIMetrics is the list of all API metrics collectors for the main server
	WorkflowOperationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_workflow_operations_total",
		Help: "The total number of workflow operations by origin, trigger, and success",
	}, []string{"operation", "origin", "trigger", "success"})

	WorkflowOperationDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "openlane_workflow_operation_duration_seconds",
		Help:    "The duration of workflow operations in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "origin", "trigger"})

	WorkflowEmitErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openlane_workflow_emit_errors_total",
		Help: "The total number of workflow emit errors by topic and origin",
	}, []string{"topic", "origin"})

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
		WorkflowOperationsTotal,
		WorkflowOperationDuration,
		WorkflowEmitErrorsTotal,
	}

	// SubscriptionMetrics is the list of all metrics specific to graph subscriptions and websocket connections
	SubscriptionMetrics = []prometheus.Collector{
		ActiveSubscriptions,
		SubscriptionOpen,
		SubscriptionClose,
		SubscriptionLifetime,
	}

	// QueueConsumerMetrics is the list of all metrics for queue consumers
	// this is currently not registered and metrics are not collected
	QueueConsumerMetrics = []prometheus.Collector{
		QueueTasksProcessedDuration,
		QueueTasksProcessed,
		QueueTasksProcessedErrors,
	}
)
