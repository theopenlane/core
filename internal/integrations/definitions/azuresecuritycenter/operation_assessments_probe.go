package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// assessmentsProbeReason is the user-facing reason recorded when the assessments probe fails
const assessmentsProbeReason = "the connection cannot read Defender for Cloud assessments; assign the Security Reader role to the Openlane app on at least one Azure subscription"

// probeAssessments verifies the installation's subscriptions are accessible and their assessments can be read
func probeAssessments(ctx context.Context, req types.OperationRequest, client *Client) error {
	requested, err := requestedSubscriptions(req)
	if err != nil {
		return err
	}

	subscriptionIDs, err := client.subscriptions(ctx, requested)
	if err != nil {
		return types.Degraded(err, assessmentsProbeReason)
	}

	if _, err := client.arm.QueryResourceGraph(ctx, assessmentsProbeQuery, subscriptionIDs); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("assessments probe failed")

		return types.Degraded(ErrAssessmentFetchFailed, assessmentsProbeReason)
	}

	return nil
}
