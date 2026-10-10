package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// runAssetSync collects AWS assets
func runAssetSync(_ context.Context, _ types.OperationRequest, _ Client, _ AssetSync) ([]types.IngestPayloadSet, error) {
	return nil, nil
}
