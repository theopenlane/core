package awssecurityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// assetSyncOperation is the AWS Config asset sync operation
var assetSyncOperation = types.OperationRefOf[AssetSync]().Ingests(configServiceClient, runAssetSync)

// runAssetSync collects AWS assets
func runAssetSync(_ context.Context, _ types.OperationRequest, _ *configservice.Client, _ AssetSync) ([]types.IngestPayloadSet, error) {

	return nil, nil
}
