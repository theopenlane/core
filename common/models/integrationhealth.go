package models

import (
	"io"
	"time"
)

// IntegrationHealth records the runtime health state of an installed integration
type IntegrationHealth struct {
	// UnhealthyReason is the user-facing reason the installation stopped syncing
	UnhealthyReason string `json:"unhealthyReason,omitempty"`
	// UnhealthyOperations maps failing operation names to their user-facing reasons
	UnhealthyOperations map[string]string `json:"unhealthyOperations,omitempty"`
	// LastSuccessfulHealthCheck is when the connection health check last passed
	LastSuccessfulHealthCheck *time.Time `json:"lastSuccessfulHealthCheck,omitempty"`
	// FailedRecords tracks ingest records a batched run could not persist, so later batched runs
	// stop requeuing them while they keep failing
	FailedRecords []FailedRecord `json:"failedRecords,omitempty"`
}

// FailedRecord is one ingest record a batched run could not persist, tracked so later batched runs
// stop requeuing it while it keeps failing
type FailedRecord struct {
	// Schema is the entityops schema name of the record
	Schema string `json:"schema"`
	// Key is the record's lookup key values in declared field order, joined by keyJoin (unexported const "\x1f")
	Key string `json:"key"`
	// RunID is the integration run that first recorded the failure
	RunID string `json:"runId"`
	// Attempts counts batched runs that failed the record since it was recorded
	Attempts int `json:"attempts"`
	// LastError is the most recent failure text
	LastError string `json:"lastError"`
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (h IntegrationHealth) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, h)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (h *IntegrationHealth) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, h)
}
