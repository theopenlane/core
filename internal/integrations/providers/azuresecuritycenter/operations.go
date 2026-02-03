package azuresecuritycenter

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	azureSecurityHealth  types.OperationName = "health.default"
	azureSecurityPricing types.OperationName = "security.pricing_overview"

	maxSampleSize = 5
)

// azureSecurityOperations registers the Defender for Cloud operations.
func azureSecurityOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        azureSecurityHealth,
			Kind:        types.OperationKindHealth,
			Description: "Call Azure Security Center pricings API to verify access.",
			Client:      ClientAzureSecurityCenterAPI,
			Run:         runAzureSecurityHealth,
		},
		{
			Name:        azureSecurityPricing,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect plan/pricing metadata for Microsoft Defender for Cloud.",
			Client:      ClientAzureSecurityCenterAPI,
			Run:         runAzureSecurityPricing,
		},
	}
}

// runAzureSecurityHealth verifies access by fetching Defender pricing data.
func runAzureSecurityHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeAzureSecurityCenter))
	if err != nil {
		return types.OperationResult{}, err
	}

	subscriptionID, err := subscriptionIDFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := listSecurityPricings(ctx, token, subscriptionID, client)
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

// runAzureSecurityPricing collects Defender pricing metadata.
func runAzureSecurityPricing(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeAzureSecurityCenter))
	if err != nil {
		return types.OperationResult{}, err
	}

	subscriptionID, err := subscriptionIDFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := listSecurityPricings(ctx, token, subscriptionID, client)
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

// listSecurityPricings queries Defender pricing data for a subscription.
func listSecurityPricings(ctx context.Context, token string, subscriptionID string, client *helpers.AuthenticatedClient) (defenderPricingResponse, error) {
	endpoint := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Security/pricings?api-version=2024-01-01", url.PathEscape(subscriptionID))
	var resp defenderPricingResponse
	if client != nil {
		if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
			return defenderPricingResponse{}, err
		}
	} else if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp); err != nil {
		return defenderPricingResponse{}, err
	}

	return resp, nil
}

// subscriptionIDFromPayload extracts the subscription ID from provider metadata.
func subscriptionIDFromPayload(payload types.CredentialPayload) (string, error) {
	subscriptionID := helpers.StringValue(payload.Data.ProviderData, "subscriptionId")
	if strings.TrimSpace(subscriptionID) == "" {
		return "", ErrSubscriptionIDMissing
	}
	return subscriptionID, nil
}
