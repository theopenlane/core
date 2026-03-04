package azuresecuritycenter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

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
	return auth.DefaultClientDescriptors(TypeAzureSecurityCenter, ClientAzureSecurityCenterAPI, "Azure management API client for Defender for Cloud", buildAzureSecurityClient)
}

// buildAzureSecurityClient constructs an Azure Security Center client from credential payload.
func buildAzureSecurityClient(_ context.Context, payload types.CredentialPayload, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := auth.OAuthTokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	subscriptionID, _ := payload.Data.ProviderData["subscriptionId"].(string)
	if subscriptionID == "" {
		return types.EmptyClientInstance(), ErrSubscriptionIDMissing
	}

	client, err := armsecurity.NewPricingsClient(staticAzureCredential{token: token}, nil)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(&azurePricingsClient{
		client: client,
		scope:  fmt.Sprintf("%s%s", azureSubscriptionScopePrefix, subscriptionID),
	}), nil
}
