package okta

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientOktaAPI identifies the Okta HTTP API client.
	ClientOktaAPI types.ClientName = "api"
)

// oktaClientDescriptors returns the client descriptors published by Okta.
func oktaClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeOkta,
			Name:         ClientOktaAPI,
			Description:  "Okta REST API client",
			Build:        buildOktaClient,
			ConfigSchema: map[string]any{"type": "object"},
		},
	}
}

// buildOktaClient constructs an authenticated Okta API client.
func buildOktaClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	apiToken, err := helpers.APITokenFromPayload(payload, string(TypeOkta))
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Authorization": "SSWS " + apiToken,
	}

	return helpers.NewAuthenticatedClient("", headers), nil
}
