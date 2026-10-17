package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// assessmentsProbeReason is the user-facing reason recorded when the assessments probe fails
	assessmentsProbeReason = "the connection cannot read Defender for Cloud assessments; assign the Security Reader role to the Openlane app on at least one Azure subscription"
	// assessmentsProbeQuery reads a single assessment to confirm Resource Graph access
	assessmentsProbeQuery = `securityresources | where type =~ "microsoft.security/assessments" | take 1`
)

// probeAssessments verifies at least one subscription is accessible and its assessments can be read
func probeAssessments(ctx context.Context, _ types.OperationRequest, client *Client) error {
	subscriptionIDs, err := client.subscriptions(ctx, nil)
	if err != nil {
		return types.Degraded(err, assessmentsProbeReason)
	}

	if _, err := client.arm.QueryResourceGraph(ctx, assessmentsProbeQuery, subscriptionIDs); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("assessments probe failed")

		return types.Degraded(ErrAssessmentFetchFailed, assessmentsProbeReason)
	}

	return nil
}
