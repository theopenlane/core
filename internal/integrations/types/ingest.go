package types

import (
	"context"
	"encoding/json"
)

// IngestFunc materializes operation result envelopes into the database;
// implementations receive a unified IngestRequest and return an IngestResult —
// the function signature is stable across ingest categories, the operation name
// and provider type inside IngestRequest determine which mapping schema is applied
type IngestFunc func(ctx context.Context, req IngestRequest) (IngestResult, error)

// IngestRequest is the unified request type passed to all IngestFunc implementations
type IngestRequest struct {
	// OrgID identifies the organization that owns the ingested records
	OrgID string
	// IntegrationID identifies the integration record
	IntegrationID string
	// Provider identifies the integration provider
	Provider ProviderType
	// Operation identifies the operation that produced the envelopes
	Operation OperationName
	// Config supplies integration-level configuration for mapping, including MappingOverrides
	Config IntegrationConfig
	// ProviderState carries provider-specific state for mapping
	ProviderState IntegrationProviderState
	// OperationConfig supplies operation-level configuration for mapping
	OperationConfig json.RawMessage
	// MappingIndex resolves provider-registered default mappings
	MappingIndex MappingIndex
	// Envelopes holds the alert payloads to ingest
	Envelopes []AlertEnvelope
}

// Validate checks that required fields are present
func (r *IngestRequest) Validate() error {
	if r.OrgID == "" {
		return ErrIngestOrgIDRequired
	}

	if r.IntegrationID == "" {
		return ErrIngestIntegrationRequired
	}

	if r.Provider == ProviderUnknown {
		return ErrIngestProviderUnknown
	}

	if r.Operation == "" {
		return ErrIngestOperationRequired
	}

	return nil
}

// IngestSummary reports mapping and persistence statistics for a single ingest run
type IngestSummary struct {
	// Total counts total envelopes processed
	Total int `json:"total"`
	// Mapped counts envelopes that were successfully mapped and persisted
	Mapped int `json:"mapped"`
	// Skipped counts envelopes filtered out by mapping
	Skipped int `json:"skipped"`
	// Failed counts envelopes that failed mapping or persistence
	Failed int `json:"failed"`
	// Created counts new records created
	Created int `json:"created"`
	// Updated counts existing records updated
	Updated int `json:"updated"`
}

// IngestResult captures the outcome of an IngestFunc call
type IngestResult struct {
	// Summary aggregates ingest totals
	Summary IngestSummary
	// Errors captures per-envelope error messages
	Errors []string
}
