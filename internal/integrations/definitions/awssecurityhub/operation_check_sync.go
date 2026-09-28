package awssecurityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// checkSyncOperation is the AWS Config check sync operation
var checkSyncOperation = types.OperationRefOf[CheckSync]().Ingests(configServiceClient, runCheckSync)

// runCheckSync collects AWS Config rules and check results
func runCheckSync(_ context.Context, _ types.OperationRequest, _ *configservice.Client, _ CheckSync) ([]types.IngestPayloadSet, error) {

	return nil, nil
}
