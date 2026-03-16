package azuresecuritycenter

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const securityPricingSampleSize = 10

// SecurityPricingSample holds a representative Defender pricing record.
type SecurityPricingSample struct {
	Name      string `json:"name"`
	Tier      string `json:"tier"`
	SubPlan   string `json:"subPlan"`
	FreeTrial string `json:"freeTrial"`
}

// SecurityPricingOverview summarizes Defender pricing metadata.
type SecurityPricingOverview struct {
	Count   int                     `json:"count"`
	Samples []SecurityPricingSample `json:"samples"`
}

// Handle adapts the pricing overview to the generic operation registration boundary
func (s SecurityPricingOverview) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		apc, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return s.Run(ctx, apc)
	}
}

// Run collects Defender pricing metadata
func (SecurityPricingOverview) Run(ctx context.Context, client *azurePricingsClient) (json.RawMessage, error) {
	resp, err := client.client.List(ctx, client.scope, nil)
	if err != nil {
		return nil, ErrPricingFetchFailed
	}

	count := min(len(resp.Value), securityPricingSampleSize)
	samples := lo.Map(resp.Value[:count], func(item *armsecurity.Pricing, _ int) SecurityPricingSample {
		sample := SecurityPricingSample{
			Name: lo.FromPtrOr(item.Name, ""),
		}
		if item.Properties != nil {
			if item.Properties.PricingTier != nil {
				sample.Tier = string(*item.Properties.PricingTier)
			}
			sample.SubPlan = lo.FromPtrOr(item.Properties.SubPlan, "")
			sample.FreeTrial = lo.FromPtrOr(item.Properties.FreeTrialRemainingTime, "")
		}
		return sample
	})

	return providerkit.EncodeResult(SecurityPricingOverview{
		Count:   len(resp.Value),
		Samples: samples,
	}, ErrResultEncode)
}
