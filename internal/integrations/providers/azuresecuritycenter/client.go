package azuresecuritycenter

import (
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
	return auth.DefaultClientDescriptors(TypeAzureSecurityCenter, ClientAzureSecurityCenterAPI, "Azure management API client for Defender for Cloud", auth.TokenClientBuilder(auth.OAuthTokenFromPayload, nil))
}
