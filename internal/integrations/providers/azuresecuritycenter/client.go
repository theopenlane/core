package azuresecuritycenter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

type azureSubscriptionMetadata struct {
	SubscriptionID string `json:"subscriptionId"`
}

const (
	// ClientAzureSecurityCenterAPI identifies the Azure management API client.
	ClientAzureSecurityCenterAPI types.ClientName = "api"

	azureSubscriptionScopePrefix = "subscriptions/"
)

// azurePricingsClient wraps armsecurity.PricingsClient with the subscription scope baked in.
type azurePricingsClient struct {
	client *armsecurity.PricingsClient
	scope  string
}

// azureSecurityCenterClientDescriptors returns the client descriptors published by Defender for Cloud.
func azureSecurityCenterClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAzureSecurityCenter, ClientAzureSecurityCenterAPI, "Azure management API client for Defender for Cloud", buildAzureSecurityClient)
}

// newAzurePricingsClient constructs an azurePricingsClient from a bearer token and subscription ID.
func newAzurePricingsClient(token string, subscriptionID string) (*azurePricingsClient, error) {
	if subscriptionID == "" {
		return nil, ErrSubscriptionIDMissing
	}

	client, err := armsecurity.NewPricingsClient(staticAzureCredential{token: token}, nil)
	if err != nil {
		return nil, err
	}

	return &azurePricingsClient{
		client: client,
		scope:  fmt.Sprintf("%s%s", azureSubscriptionScopePrefix, subscriptionID),
	}, nil
}

// buildAzureSecurityClient constructs an Azure Security Center client from credential payload.
func buildAzureSecurityClient(_ context.Context, payload models.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := auth.OAuthTokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	var meta azureSubscriptionMetadata
	if err := json.Unmarshal(payload.ProviderData, &meta); err != nil {
		return types.EmptyClientInstance(), err
	}

	apc, err := newAzurePricingsClient(token, meta.SubscriptionID)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(apc), nil
}
