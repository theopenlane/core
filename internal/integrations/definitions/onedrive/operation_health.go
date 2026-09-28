package onedrive

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// HealthCheck holds the result of a OneDrive health check
type HealthCheck struct {
	// DriveID is the resolved identifier of the user's default OneDrive
	DriveID string `json:"driveId"`
}

// checkHealth executes the health check by verifying the user's drive is accessible
func checkHealth(ctx context.Context, _ types.OperationRequest, c *DriveClient) (json.RawMessage, error) {
	drive, err := c.Graph.Me().Drive().Get(ctx, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("onedrive: health check drive.get failed")
		return nil, ErrHealthCheckFailed
	}

	return providerkit.EncodeResult(HealthCheck{DriveID: lo.FromPtr(drive.GetId())}, ErrResultEncode)
}
