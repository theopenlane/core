package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// AssessmentsCollect collects Azure Defender for Cloud assessment findings for ingest
type AssessmentsCollect struct {
	types.OperationSettings
	// SubscriptionIDs limits collection to these subscriptions; when empty every accessible subscription is collected
	SubscriptionIDs []string `json:"subscriptionIds,omitempty" jsonschema:"title=Subscription IDs,description=Optional list of Azure subscription IDs to collect from; leave empty to collect from every subscription Openlane has been granted access to"`
}

// runAssessmentsCollect collects all unhealthy security posture assessments across the configured subscriptions
func runAssessmentsCollect(ctx context.Context, _ types.OperationRequest, client *Client, cfg AssessmentsCollect) ([]types.IngestPayloadSet, error) {
	envelopes, err := client.collect(ctx, findingsQuery, cfg.SubscriptionIDs, recommendationFindings)
	if err != nil {
		return nil, err
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaFinding.Name,
			Envelopes: envelopes,
		},
	}, nil
}
