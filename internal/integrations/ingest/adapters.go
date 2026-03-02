package ingest

import "context"

// VulnerabilityIngestFunc returns an IngestFunc that delegates to VulnerabilityAlerts.
// The IngestSummary fields map 1:1 from VulnerabilityIngestSummary.
func VulnerabilityIngestFunc() IngestFunc {
	return func(ctx context.Context, req IngestRequest) (IngestResult, error) {
		concrete := VulnerabilityIngestRequest{
			OrgID:             req.OrgID,
			IntegrationID:     req.IntegrationID,
			Provider:          req.Provider,
			Operation:         req.Operation,
			IntegrationConfig: req.IntegrationConfig,
			ProviderState:     req.ProviderState,
			OperationConfig:   req.OperationConfig,
			Envelopes:         req.Envelopes,
			DB:                req.DB,
		}

		result, err := VulnerabilityAlerts(ctx, concrete)
		return IngestResult{
			Summary: IngestSummary{
				Total:     result.Summary.Total,
				Mapped:    result.Summary.Mapped,
				Persisted: result.Summary.Persisted,
				Skipped:   result.Summary.Skipped,
				Failed:    result.Summary.Failed,
				Created:   result.Summary.Created,
				Updated:   result.Summary.Updated,
			},
			Errors: result.Errors,
		}, err
	}
}

// DirectoryAccountIngestFunc returns an IngestFunc that delegates to DirectoryAccounts.
// The IngestSummary fields map 1:1 from DirectoryAccountIngestSummary.
func DirectoryAccountIngestFunc() IngestFunc {
	return func(ctx context.Context, req IngestRequest) (IngestResult, error) {
		concrete := DirectoryAccountIngestRequest{
			OrgID:             req.OrgID,
			IntegrationID:     req.IntegrationID,
			Provider:          req.Provider,
			Operation:         req.Operation,
			IntegrationConfig: req.IntegrationConfig,
			ProviderState:     req.ProviderState,
			OperationConfig:   req.OperationConfig,
			Envelopes:         req.Envelopes,
			DB:                req.DB,
		}

		result, err := DirectoryAccounts(ctx, concrete)
		return IngestResult{
			Summary: IngestSummary{
				Total:     result.Summary.Total,
				Mapped:    result.Summary.Mapped,
				Persisted: result.Summary.Persisted,
				Skipped:   result.Summary.Skipped,
				Failed:    result.Summary.Failed,
				Created:   result.Summary.Created,
				Updated:   result.Summary.Updated,
			},
			Errors: result.Errors,
		}, err
	}
}
