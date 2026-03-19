package azuresecuritycenter

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck reports whether Defender pricing data can be listed
type HealthCheck struct {
	// Count is the number of Defender pricing records returned by the health probe
	Count int `json:"count"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		apc, err := SecurityCenterClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, apc)
	}
}

// Run verifies access by fetching Defender pricing data
func (HealthCheck) Run(ctx context.Context, client *azurePricingsClient) (json.RawMessage, error) {
	resp, err := client.client.List(ctx, client.scope, nil)
	if err != nil {
		return nil, ErrPricingFetchFailed
	}

	return providerkit.EncodeResult(HealthCheck{Count: len(resp.Value)}, ErrResultEncode)
}
