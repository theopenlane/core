package azuresecuritycenter

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/providers/helpers"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	azureSecurityHealth  types.OperationName = "health.default"
	azureSecurityPricing types.OperationName = "security.pricing_overview"

	maxSampleSize = 5
)

func azureSecurityOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        azureSecurityHealth,
			Kind:        types.OperationKindHealth,
			Description: "Call Azure Security Center pricings API to verify access.",
			Run:         runAzureSecurityHealth,
		},
		{
			Name:        azureSecurityPricing,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect plan/pricing metadata for Microsoft Defender for Cloud.",
			Run:         runAzureSecurityPricing,
		},
	}
}

func runAzureSecurityHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeAzureSecurityCenter))
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := listSecurityPricings(ctx, token)
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Azure Security Center pricing fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Retrieved %d pricing entries", len(resp.Value)),
		Details: map[string]any{
			"count": len(resp.Value),
		},
	}, nil
}

func runAzureSecurityPricing(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeAzureSecurityCenter))
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := listSecurityPricings(ctx, token)
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Azure Security Center pricing fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	samples := make([]map[string]any, 0, maxSampleSize)
	for _, item := range resp.Value {
		if len(samples) >= cap(samples) {
			break
		}
		samples = append(samples, map[string]any{
			"name":      item.Name,
			"tier":      item.Properties.PricingTier,
			"subPlan":   item.Properties.SubPlan,
			"freeTrial": item.Properties.FreeTrialRemainingTime,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d Defender pricing records", len(resp.Value)),
		Details: map[string]any{
			"count":   len(resp.Value),
			"samples": samples,
		},
	}, nil
}

type defenderPricingResponse struct {
	Value []defenderPricing `json:"value"`
}

type defenderPricing struct {
	Name       string                 `json:"name"`
	Properties defenderPricingDetails `json:"properties"`
}

type defenderPricingDetails struct {
	PricingTier            string `json:"pricingTier"`
	SubPlan                string `json:"subPlan"`
	FreeTrialRemainingTime string `json:"freeTrialRemainingTime"`
}

func listSecurityPricings(ctx context.Context, token string) (defenderPricingResponse, error) {
	const endpoint = "https://management.azure.com/providers/Microsoft.Security/pricings?api-version=2024-01-01"
	var resp defenderPricingResponse
	if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp); err != nil {
		return defenderPricingResponse{}, err
	}

	return resp, nil
}
