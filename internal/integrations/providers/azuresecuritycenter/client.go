package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAzureSecurityCenterAPI identifies the Azure management API client.
	ClientAzureSecurityCenterAPI types.ClientName = "api"
)

// azureSecurityCenterClientDescriptors returns the client descriptors published by Defender for Cloud.
func azureSecurityCenterClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeAzureSecurityCenter, ClientAzureSecurityCenterAPI, "Azure management API client for Defender for Cloud", buildAzureSecurityCenterClient)
}

// buildAzureSecurityCenterClient constructs an authenticated Azure management API client.
func buildAzureSecurityCenterClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := auth.OAuthTokenFromPayload(payload, string(TypeAzureSecurityCenter))
	if err != nil {
		return nil, err
	}

	return auth.NewAuthenticatedClient(token, nil), nil
}
