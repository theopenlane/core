package runner

import "errors"

var (
	// ErrOperationContextMissing indicates required operation dependencies are missing
	ErrOperationContextMissing = errors.New("runner: integration operation context missing")
	// ErrIntegrationRunIDRequired indicates the run identifier is missing
	ErrIntegrationRunIDRequired = errors.New("runner: integration run id required")
	// ErrIntegrationRecordMissing indicates the integration record is missing for the run
	ErrIntegrationRecordMissing = errors.New("runner: integration record missing for run")
	// ErrIntegrationProviderUnknown indicates the integration provider could not be resolved
	ErrIntegrationProviderUnknown = errors.New("runner: integration provider unknown")
	// ErrOperationNameRequired indicates the operation name is missing on the run
	ErrOperationNameRequired = errors.New("runner: operation name required for run")
	// ErrOperationFailed indicates the operation failed to execute successfully
	ErrOperationFailed = errors.New("runner: integration operation failed")
	// ErrAlertPayloadsMissing indicates alert payloads are missing from operation output
	ErrAlertPayloadsMissing = errors.New("runner: alert payloads missing")
)
