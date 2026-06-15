package googledrive

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const healthCheckMaxResults = int64(1)

// HealthCheck holds the result of a Google Drive health check
type HealthCheck struct {
	// FileCount is the number of files returned by the probe
	FileCount int `json:"fileCount"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(driveClient, h.Run)
}

// Run executes the health check by listing files accessible via the Drive API
func (HealthCheck) Run(ctx context.Context, c DriveClient) (json.RawMessage, error) {
	resp, err := c.Svc.Files.List().
		PageSize(healthCheckMaxResults).
		Fields("files(id)").
		Context(ctx).
		Do()
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("googledrive: health check files.list failed")
		return nil, fmt.Errorf("%w: %v", ErrHealthCheckFailed, err)
	}

	return providerkit.EncodeResult(HealthCheck{FileCount: len(resp.Files)}, ErrResultEncode)
}
