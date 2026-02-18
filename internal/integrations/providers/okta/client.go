package okta

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientOktaAPI identifies the Okta HTTP API client.
	ClientOktaAPI types.ClientName = "api"
)

// oktaClientDescriptors returns the client descriptors published by Okta.
func oktaClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeOkta, ClientOktaAPI, "Okta REST API client", buildOktaClient)
}

// buildOktaClient constructs an authenticated Okta API client.
func buildOktaClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	apiToken, err := auth.APITokenFromPayload(payload)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Authorization": "SSWS " + apiToken,
	}

	return auth.NewAuthenticatedClient("", headers), nil
}
