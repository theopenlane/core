package okta

import (
	"context"
	"encoding/json"

	okta "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientOktaAPI identifies the Okta HTTP API client.
	ClientOktaAPI types.ClientName = "api"
)

// oktaClientDescriptors returns the client descriptors published by Okta.
func oktaClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeOkta, ClientOktaAPI, "Okta API client", buildOktaClient)
}

// buildOktaClient constructs an Okta SDK API client from credential payload.
func buildOktaClient(_ context.Context, payload types.CredentialPayload, _ json.RawMessage) (types.ClientInstance, error) {
	apiToken, err := auth.APITokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	orgURL, _ := payload.Data.ProviderData["orgUrl"].(string)
	if orgURL == "" {
		return types.EmptyClientInstance(), ErrCredentialsMissing
	}

	cfg, err := okta.NewConfiguration(
		okta.WithOrgUrl(orgURL),
		okta.WithToken(apiToken),
	)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(okta.NewAPIClient(cfg)), nil
}
