package ingest

import (
	"context"
	"encoding/json"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	integrationstate "github.com/theopenlane/core/internal/integrations/state"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

// IngestFunc materializes operation result envelopes into the database;
// implementations receive a unified IngestRequest and return an IngestResult.
// The function signature is stable across ingest categories — the operation name
// and provider type inside IngestRequest determine which mapping schema is applied
//
//revive:disable-next-line
type IngestFunc func(ctx context.Context, req IngestRequest) (IngestResult, error)

// IngestRequest is the unified request type passed to all IngestFunc implementations
//
//revive:disable-next-line
type IngestRequest struct {
	// OrgID identifies the organization that owns the ingested records
	OrgID string
	// IntegrationID identifies the integration record
	IntegrationID string
	// Provider identifies the integration provider
	Provider integrationtypes.ProviderType
	// Operation identifies the operation that produced the envelopes
	Operation integrationtypes.OperationName
	// IntegrationConfig supplies integration-level configuration for mapping
	IntegrationConfig openapi.IntegrationConfig
	// ProviderState carries provider-specific state for mapping
	ProviderState integrationstate.IntegrationProviderState
	// OperationConfig supplies operation-level configuration for mapping
	OperationConfig json.RawMessage
	// MappingIndex resolves provider-registered default mappings.
	MappingIndex integrationtypes.MappingIndex
	// Envelopes holds the alert payloads to ingest
	Envelopes []integrationtypes.AlertEnvelope
	// DB provides access to the persistence layer
	DB *generated.Client
}

// Validate checks that required fields are present
func (r *IngestRequest) Validate() error {
	if r.OrgID == "" {
		return ErrIngestOrgIDRequired
	}
	if r.IntegrationID == "" {
		return ErrIngestIntegrationRequired
	}
	if r.Provider == integrationtypes.ProviderUnknown {
		return ErrIngestProviderUnknown
	}
	if r.Operation == "" {
		return ErrIngestOperationRequired
	}
	if r.DB == nil {
		return ErrDBClientRequired
	}

	return nil
}

// IngestSummary reports mapping and persistence statistics for a single ingest run
//
//revive:disable-next-line
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
//
//revive:disable-next-line
type IngestResult struct {
	// Summary aggregates ingest totals
	Summary IngestSummary
	// Errors captures per-envelope error messages
	Errors []string
}
