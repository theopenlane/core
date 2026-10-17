package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// VulnerabilitiesCollect collects Azure Defender for Cloud vulnerability recommendations for ingest
type VulnerabilitiesCollect struct {
	types.OperationSettings
}

// runVulnerabilitiesCollect collects all unhealthy software update and service upgrade recommendations across the installation's subscriptions
func runVulnerabilitiesCollect(ctx context.Context, req types.OperationRequest, client *Client, _ VulnerabilitiesCollect) ([]types.IngestPayloadSet, error) {
	subscriptionIDs, err := requestedSubscriptions(req)
	if err != nil {
		return nil, err
	}

	envelopes, err := client.collect(ctx, vulnerabilitiesQuery, subscriptionIDs, splitCVEs)
	if err != nil {
		return nil, err
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaVulnerability.Name,
			Envelopes: envelopes,
		},
	}, nil
}
