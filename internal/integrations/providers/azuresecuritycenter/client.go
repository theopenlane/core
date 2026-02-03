package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAzureSecurityCenterAPI identifies the Azure management API client.
	ClientAzureSecurityCenterAPI types.ClientName = "api"
)

// azureSecurityCenterClientDescriptors returns the client descriptors published by Defender for Cloud.
func azureSecurityCenterClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeAzureSecurityCenter,
			Name:         ClientAzureSecurityCenterAPI,
			Description:  "Azure management API client for Defender for Cloud",
			Build:        buildAzureSecurityCenterClient,
			ConfigSchema: map[string]any{"type": "object"},
		},
	}
}

// buildAzureSecurityCenterClient constructs an authenticated Azure management API client.
func buildAzureSecurityCenterClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.OAuthTokenFromPayload(payload, string(TypeAzureSecurityCenter))
	if err != nil {
		return nil, err
	}

	return helpers.NewAuthenticatedClient(token, nil), nil
}
