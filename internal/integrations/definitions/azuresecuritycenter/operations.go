package azuresecuritycenter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	azureSubscriptionScopePrefix = "subscriptions/"
	defaultAzureScope            = "https://management.azure.com/.default"
	azureTokenURLTemplate        = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

// staticAzureCredential implements azcore.TokenCredential for a pre-obtained bearer token
type staticAzureCredential struct {
	token string
}

// GetToken satisfies azcore.TokenCredential for a pre-obtained bearer token
func (s staticAzureCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

// azurePricingsClient wraps a PricingsClient with its subscription scope
type azurePricingsClient struct {
	client *armsecurity.PricingsClient
	scope  string
}

type azureHealthDetails struct {
	Count int `json:"count"`
}

type azurePricingSample struct {
	Name      string `json:"name"`
	Tier      string `json:"tier"`
	SubPlan   string `json:"subPlan"`
	FreeTrial string `json:"freeTrial"`
}

type azurePricingDetails struct {
	Count   int                  `json:"count"`
	Samples []azurePricingSample `json:"samples"`
}

// buildAzureSecurityClient constructs an Azure Defender for Cloud client via client credentials flow
func buildAzureSecurityClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	var cred credential
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, err
	}

	switch {
	case cred.TenantID == "":
		return nil, ErrTenantIDMissing
	case cred.ClientID == "":
		return nil, ErrClientIDMissing
	case cred.ClientSecret == "":
		return nil, ErrClientSecretMissing
	case cred.SubscriptionID == "":
		return nil, ErrSubscriptionIDMissing
	}

	tokenURL := fmt.Sprintf(azureTokenURLTemplate, cred.TenantID)
	cfg := clientcredentials.Config{
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{defaultAzureScope},
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	token, err := cfg.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchangeFailed, err)
	}

	pricingsClient, err := armsecurity.NewPricingsClient(staticAzureCredential{token: token.AccessToken}, nil)
	if err != nil {
		return nil, fmt.Errorf("azuresecuritycenter: pricings client build failed: %w", err)
	}

	return &azurePricingsClient{
		client: pricingsClient,
		scope:  fmt.Sprintf("%s%s", azureSubscriptionScopePrefix, cred.SubscriptionID),
	}, nil
}

// runHealthOperation verifies access by fetching Defender pricing data
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	apc, ok := client.(*azurePricingsClient)
	if !ok {
		return nil, ErrClientType
	}

	resp, err := apc.client.List(ctx, apc.scope, nil)
	if err != nil {
		return nil, fmt.Errorf("azuresecuritycenter: pricing fetch failed: %w", err)
	}

	return jsonx.ToRawMessage(azureHealthDetails{Count: len(resp.Value)})
}

// runSecurityPricingOperation collects Defender pricing metadata
func runSecurityPricingOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	apc, ok := client.(*azurePricingsClient)
	if !ok {
		return nil, ErrClientType
	}

	resp, err := apc.client.List(ctx, apc.scope, nil)
	if err != nil {
		return nil, fmt.Errorf("azuresecuritycenter: pricing fetch failed: %w", err)
	}

	const sampleSize = 10
	count := min(len(resp.Value), sampleSize)
	samples := lo.Map(resp.Value[:count], func(item *armsecurity.Pricing, _ int) azurePricingSample {
		sample := azurePricingSample{
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

	return jsonx.ToRawMessage(azurePricingDetails{
		Count:   len(resp.Value),
		Samples: samples,
	})
}
