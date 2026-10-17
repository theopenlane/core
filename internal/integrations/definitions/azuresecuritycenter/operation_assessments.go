package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// AssessmentsCollect collects Azure Defender for Cloud assessment findings for ingest
type AssessmentsCollect struct {
	types.OperationSettings
}

// runAssessmentsCollect collects all unhealthy security posture assessments across the installation's subscriptions
func runAssessmentsCollect(ctx context.Context, req types.OperationRequest, client *Client, _ AssessmentsCollect) ([]types.IngestPayloadSet, error) {
	subscriptionIDs, err := requestedSubscriptions(req)
	if err != nil {
		return nil, err
	}

	envelopes, err := client.collect(ctx, findingsQuery, subscriptionIDs, recommendationFindings)
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
