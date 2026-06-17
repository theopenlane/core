package onedrive

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// HealthCheck holds the result of a OneDrive health check
type HealthCheck struct {
	// DriveID is the resolved identifier of the user's default OneDrive
	DriveID string `json:"driveId"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(oneDriveClient, h.Run)
}

// Run executes the health check by verifying the user's drive is accessible
func (HealthCheck) Run(ctx context.Context, c *DriveClient) (json.RawMessage, error) {
	drive, err := c.Graph.Me().Drive().Get(ctx, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("onedrive: health check drive.get failed")
		return nil, ErrHealthCheckFailed
	}

	return providerkit.EncodeResult(HealthCheck{DriveID: lo.FromPtr(drive.GetId())}, ErrResultEncode)
}
