package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// runCheckSync collects AWS Config rules and check results
func runCheckSync(_ context.Context, _ types.OperationRequest, _ Client, _ CheckSync) ([]types.IngestPayloadSet, error) {

	return nil, nil
}
