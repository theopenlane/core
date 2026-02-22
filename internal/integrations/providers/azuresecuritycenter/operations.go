package azuresecuritycenter

import (
	"context"
	"fmt"
	"net/url"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	azureSecurityHealth  types.OperationName = "health.default"
	azureSecurityPricing types.OperationName = "security.pricing_overview"
)

// azureSecurityOperations registers the Defender for Cloud operations.
func azureSecurityOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(azureSecurityHealth, "Call Azure Security Center pricings API to verify access.", ClientAzureSecurityCenterAPI, runAzureSecurityHealth),
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
	client, token, err := auth.ClientAndOAuthToken(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	subscriptionID, err := subscriptionIDFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := listSecurityPricings(ctx, token, subscriptionID, client)
	if err != nil {
		return operations.OperationFailure("Azure Security Center pricing fetch failed", err, nil)
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
	client, token, err := auth.ClientAndOAuthToken(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	subscriptionID, err := subscriptionIDFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := listSecurityPricings(ctx, token, subscriptionID, client)
	if err != nil {
		return operations.OperationFailure("Azure Security Center pricing fetch failed", err, nil)
	}

	samples := lo.Map(resp.Value[:min(len(resp.Value), operations.DefaultSampleSize)], func(item defenderPricing, _ int) map[string]any {
		return map[string]any{
			"name":      item.Name,
			"tier":      item.Properties.PricingTier,
			"subPlan":   item.Properties.SubPlan,
			"freeTrial": item.Properties.FreeTrialRemainingTime,
		}
	})

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
	// Value holds pricing entries returned by the API
	Value []defenderPricing `json:"value"`
}

type defenderPricing struct {
	// Name is the pricing resource name
	Name string `json:"name"`
	// Properties holds pricing details for the resource
	Properties defenderPricingDetails `json:"properties"`
}

type defenderPricingDetails struct {
	// PricingTier is the Defender pricing tier
	PricingTier string `json:"pricingTier"`
	// SubPlan is the Defender sub-plan identifier
	SubPlan string `json:"subPlan"`
	// FreeTrialRemainingTime reports remaining free trial duration
	FreeTrialRemainingTime string `json:"freeTrialRemainingTime"`
}

// listSecurityPricings queries Defender pricing data for a subscription.
func listSecurityPricings(ctx context.Context, token string, subscriptionID string, client *auth.AuthenticatedClient) (defenderPricingResponse, error) {
	endpoint := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Security/pricings?api-version=2024-01-01", url.PathEscape(subscriptionID))
	var resp defenderPricingResponse
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return defenderPricingResponse{}, err
	}

	return resp, nil
}

// subscriptionIDFromPayload extracts the subscription ID from provider metadata.
func subscriptionIDFromPayload(payload types.CredentialPayload) (string, error) {
	subscriptionID, _ := payload.Data.ProviderData["subscriptionId"].(string)
	if subscriptionID == "" {
		return "", ErrSubscriptionIDMissing
	}

	return subscriptionID, nil
}
