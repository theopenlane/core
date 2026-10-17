package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// VulnerabilitiesCollect collects Azure Defender for Cloud vulnerability recommendations for ingest
type VulnerabilitiesCollect struct {
	types.OperationSettings
	// SubscriptionIDs limits collection to these subscriptions; when empty every accessible subscription is collected
	SubscriptionIDs []string `json:"subscriptionIds,omitempty" jsonschema:"title=Subscription IDs,description=Optional list of Azure subscription IDs to collect from; leave empty to collect from every subscription Openlane has been granted access to"`
}

// runVulnerabilitiesCollect collects all unhealthy software update and service upgrade recommendations across the configured subscriptions
func runVulnerabilitiesCollect(ctx context.Context, _ types.OperationRequest, client *Client, cfg VulnerabilitiesCollect) ([]types.IngestPayloadSet, error) {
	envelopes, err := client.collect(ctx, vulnerabilitiesQuery, cfg.SubscriptionIDs, splitCVEs)
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
