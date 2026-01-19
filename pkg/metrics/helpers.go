package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// LabelValues provides a structured way to define metric labels
// This helps avoid typos and ensures consistency across the codebase
type LabelValues struct {
	// Common labels
	Success   string
	Operation string
	Method    string
	Status    string

	// Service-specific labels
	WorkerName   string
	QueueName    string
	TaskName     string
	RequestType  string
	FunctionName string
	ErrorCode    string

	// Resource labels
	TableName   string
	CacheName   string
	ServiceName string
}

// StandardLabels returns a slice of label values in the correct order
// for the given metric. This ensures consistency and reduces errors.
func (l LabelValues) ForWorker() []string {
	return []string{l.WorkerName}
}

func (l LabelValues) ForLogin() []string {
	return []string{l.Success}
}

func (l LabelValues) ForRequestValidation() []string {
	return []string{l.RequestType, l.Success}
}

func (l LabelValues) ForHandlerError() []string {
	return []string{l.FunctionName, l.ErrorCode}
}

func (l LabelValues) ForHandlerResult() []string {
	return []string{l.RequestType, l.Success}
}

func (l LabelValues) ForQueueTask() []string {
	return []string{l.QueueName, l.TaskName}
}

func (l LabelValues) ForQueueTaskProcessed() []string {
	return []string{l.TaskName}
}

func (l LabelValues) ForGraphQLOperation() []string {
	return []string{l.Operation, l.Success}
}

// MetricNameBuilder helps construct consistent metric names
type MetricNameBuilder struct {
	namespace string
	subsystem string
}

// NewMetricNameBuilder creates a new builder with default namespace
func NewMetricNameBuilder() *MetricNameBuilder {
	return &MetricNameBuilder{
		namespace: "openlane",
	}
}

// WithSubsystem sets the subsystem for the metric
func (b *MetricNameBuilder) WithSubsystem(subsystem string) *MetricNameBuilder {
	b.subsystem = subsystem
	return b
}

// Build constructs the full metric name
func (b *MetricNameBuilder) Build(name string) string {
	if b.subsystem != "" {
		return prometheus.BuildFQName(b.namespace, b.subsystem, name)
	}

	return prometheus.BuildFQName(b.namespace, "", name)
}

// Common label values to use across metrics
const (
	// Success values
	LabelSuccess = "true"
	LabelFailure = "false"

	// Common operations
	OperationCreate = "create"
	OperationRead   = "read"
	OperationUpdate = "update"
	OperationDelete = "delete"
	OperationList   = "list"

	// Queue names
	QueueEmail     = "email"
	QueueWebhook   = "webhook"
	QueueAnalytics = "analytics"

	// Task types
	TaskEmailWelcome      = "welcome_email"
	TaskEmailVerification = "verification_email"
	TaskEmailReset        = "reset_email"
	TaskWebhookDelivery   = "webhook_delivery"

	// Authentication types
	AuthTypeJWT          = "jwt"
	AuthTypeJWTAnonymous = "jwt_anonymous"
	AuthTypePAT          = "pat"
	AuthTypeAPIToken     = "api_token"
)

// RecordWorkerExecution records a worker execution with appropriate labels
func RecordWorkerExecution(workerName string, err error) {
	if err != nil {
		WorkerExecutionErrors.WithLabelValues(workerName).Inc()
	} else {
		WorkerExecutions.WithLabelValues(workerName).Inc()
	}
}

// RecordLogin records a login attempt
func RecordLogin(success bool) {
	successStr := LabelFailure
	if success {
		successStr = LabelSuccess
	}

	Logins.WithLabelValues(successStr).Inc()
}

// RecordRegistration records a registration attempt
func RecordRegistration() {
	Registrations.Inc()
}

// RecordEmailValidation records an email validation result
func RecordEmailValidation(success bool, result string) {
	successStr := LabelFailure
	if success {
		successStr = LabelSuccess
	}

	EmailValidations.WithLabelValues(successStr, result).Inc()
}

// RecordRequestValidation records a request validation result
func RecordRequestValidation(requestType string, success bool) {
	successStr := LabelFailure
	if success {
		successStr = LabelSuccess
	}

	RequestValidations.WithLabelValues(requestType, successStr).Inc()
}

// RecordHandlerError records a handler error using HTTP status code
// The function name is derived from http.StatusText() for consistency
func RecordHandlerError(statusCode int) {
	HandlerErrors.WithLabelValues(http.StatusText(statusCode), strconv.Itoa(statusCode)).Inc()
}

// RecordHandlerResult records a handler result
func RecordHandlerResult(requestType string, success bool) {
	successStr := LabelFailure
	if success {
		successStr = LabelSuccess
	}

	HandlerResults.WithLabelValues(requestType, successStr).Inc()
}

// RecordQueueTaskPushed records a task being pushed to queue
func RecordQueueTaskPushed(queueName, taskName string, err error) {
	if err != nil {
		QueueTasksPushFailures.WithLabelValues(queueName, taskName).Inc()
	} else {
		QueueTasksPushed.WithLabelValues(queueName, taskName).Inc()
	}
}

// RecordQueueTaskProcessed records a task being processed
func RecordQueueTaskProcessed(taskName string, duration float64, err error) {
	QueueTasksProcessedDuration.WithLabelValues(taskName).Observe(duration)

	if err != nil {
		QueueTasksProcessedErrors.WithLabelValues(taskName).Inc()
	} else {
		QueueTasksProcessed.WithLabelValues(taskName).Inc()
	}
}

// RecordGraphQLOperation records a GraphQL operation with complexity
func RecordGraphQLOperation(operation string, duration float64, complexity int, err error) {
	successStr := LabelSuccess
	if err != nil {
		successStr = LabelFailure
	}

	GraphQLOperationTotal.WithLabelValues(operation, successStr).Inc()
	GraphQLOperationDuration.WithLabelValues(operation).Observe(duration)
	GraphQLQueryComplexity.WithLabelValues(operation).Observe(float64(complexity))
}

// RecordGraphQLRejection records a rejected GraphQL query
func RecordGraphQLRejection(reason string) {
	GraphQLQueryRejected.WithLabelValues(reason).Inc()
}

// StartFileUpload records the start of a file upload
func StartFileUpload() {
	FileUploadsStarted.Inc()
	FileUploadsInFlight.Inc()
}

// FinishFileUpload records the completion of a file upload
func FinishFileUpload(status string, duration float64) {
	FileUploadsInFlight.Dec()
	FileUploadsCompleted.WithLabelValues(status).Inc()
	FileUploadDuration.WithLabelValues(status).Observe(duration)
}

// RecordFileBufferingStrategy records which buffering strategy was used
func RecordFileBufferingStrategy(strategy string) {
	FileBufferingStrategy.WithLabelValues(strategy).Inc()
}

// RecordStorageUpload records an upload operation for a storage provider
func RecordStorageUpload(provider string, bytes int64) {
	StorageProviderUploads.WithLabelValues(provider).Inc()
	StorageProviderBytesUploaded.WithLabelValues(provider).Add(float64(bytes))
}

// RecordStorageDownload records a download operation for a storage provider
func RecordStorageDownload(provider string, bytes int64) {
	StorageProviderDownloads.WithLabelValues(provider).Inc()
	StorageProviderBytesDownloaded.WithLabelValues(provider).Add(float64(bytes))
}

// RecordStorageDelete records a delete operation for a storage provider
func RecordStorageDelete(provider string) {
	StorageProviderDeletes.WithLabelValues(provider).Inc()
}

// RecordAuthentication records an authentication attempt by type
func RecordAuthentication(authType string) {
	AuthenticationAttempts.WithLabelValues(authType).Inc()
}

// RecordSubscriptionOpened records when a subscription is opened
func RecordSubscriptionOpened() {
	ActiveSubscriptions.Inc()
	SubscriptionOpen.Inc()
}

// RecordSubscriptionClosed records when a subscription is closed
func RecordSubscriptionClosed(lifetimeSeconds float64) {
	ActiveSubscriptions.Dec()
	SubscriptionClose.Inc()
	SubscriptionLifetime.Observe(lifetimeSeconds)
}
