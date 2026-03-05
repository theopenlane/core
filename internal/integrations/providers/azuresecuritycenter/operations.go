package azuresecuritycenter

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	azureSecurityHealth  types.OperationName = types.OperationHealthDefault
	azureSecurityPricing types.OperationName = "security.pricing_overview"
)

type azureSecurityHealthDetails struct {
	Count int `json:"count"`
}

type azureSecurityPricingSample struct {
	Name      string `json:"name"`
	Tier      string `json:"tier"`
	SubPlan   string `json:"subPlan"`
	FreeTrial string `json:"freeTrial"`
}

type azureSecurityPricingDetails struct {
	Count   int                          `json:"count"`
	Samples []azureSecurityPricingSample `json:"samples"`
}

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

// resolveAzureSecurityClient returns a pooled Azure pricings client or builds one from the credential payload.
func resolveAzureSecurityClient(_ context.Context, input types.OperationInput) (*azurePricingsClient, error) {
	if c, ok := types.ClientInstanceAs[*azurePricingsClient](input.Client); ok {
		return c, nil
	}

	token, err := auth.OAuthTokenFromPayload(input.Credential)
	if err != nil {
		return nil, err
	}

	subscriptionID, _ := input.Credential.Data.ProviderData["subscriptionId"].(string)
	if subscriptionID == "" {
		return nil, ErrSubscriptionIDMissing
	}

	cred := staticAzureCredential{token: token}

	client, err := armsecurity.NewPricingsClient(cred, nil)
	if err != nil {
		return nil, err
	}

	return &azurePricingsClient{
		client: client,
		scope:  fmt.Sprintf("%s%s", azureSubscriptionScopePrefix, subscriptionID),
	}, nil
}

// runAzureSecurityHealth verifies access by fetching Defender pricing data.
func runAzureSecurityHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	apc, err := resolveAzureSecurityClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := apc.client.List(ctx, apc.scope, nil)
	if err != nil {
		return operations.OperationFailure("Azure Security Center pricing fetch failed", err, nil)
	}

	return operations.OperationSuccess(fmt.Sprintf("Retrieved %d pricing entries", len(resp.Value)), azureSecurityHealthDetails{
		Count: len(resp.Value),
	}), nil
}

// runAzureSecurityPricing collects Defender pricing metadata.
func runAzureSecurityPricing(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	apc, err := resolveAzureSecurityClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := apc.client.List(ctx, apc.scope, nil)
	if err != nil {
		return operations.OperationFailure("Azure Security Center pricing fetch failed", err, nil)
	}

	samples := lo.Map(resp.Value[:min(len(resp.Value), operations.DefaultSampleSize)], func(item *armsecurity.Pricing, _ int) azureSecurityPricingSample {
		sample := azureSecurityPricingSample{
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

	return operations.OperationSuccess(fmt.Sprintf("Collected %d Defender pricing records", len(resp.Value)), azureSecurityPricingDetails{
		Count:   len(resp.Value),
		Samples: samples,
	}), nil
}
