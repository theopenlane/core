package okta

import (
	"context"
	"encoding/json"

	okta "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// ClientOktaAPI identifies the Okta HTTP API client.
	ClientOktaAPI types.ClientName = "api"
)

// oktaClientDescriptors returns the client descriptors published by Okta.
func oktaClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeOkta, ClientOktaAPI, "Okta API client", buildOktaClient)
}

type oktaProviderMetadata struct {
	OrgURL string `json:"orgUrl"`
}

// buildOktaClient constructs an Okta SDK API client from credential payload.
func buildOktaClient(_ context.Context, payload models.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	apiToken, err := auth.APITokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	var meta oktaProviderMetadata
	if err := jsonx.UnmarshalIfPresent(payload.ProviderData, &meta); err != nil {
		return types.EmptyClientInstance(), err
	}

	orgURL := meta.OrgURL
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
