package okta

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Okta definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "okta",
				DisplayName: "Okta",
				Description: "Collect Okta tenant and sign-on policy metadata for identity posture and access governance analysis.",
				Category:    "sso",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/okta/overview",
				Labels:      map[string]string{"vendor": "okta", "product": "identity"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[CredentialSchema](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         OktaClient.ID(),
					Description: "Okta API client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Okta user API to verify API token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   OktaClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        PoliciesCollectOperation.Name(),
					Description: "Collect sign-on policy metadata for posture analysis",
					Topic:       PoliciesCollectOperation.Topic(Slug),
					ClientRef:   OktaClient.ID(),
					Handle:      PoliciesCollect{}.Handle(),
				},
			},
		}, nil
	})
}
