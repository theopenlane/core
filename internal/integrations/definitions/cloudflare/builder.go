package cloudflare

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Cloudflare definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "cloudflare",
				DisplayName: "Cloudflare",
				Description: "Validate Cloudflare account access and collect security-relevant account and zone context for posture workflows.",
				Category:    "security",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/cloudflare/overview",
				Labels:      map[string]string{"vendor": "cloudflare", "product": "zero-trust"},
				Active:      true,
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
					Ref:         CloudflareClient.ID(),
					Description: "Cloudflare REST API client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify Cloudflare API token via /user/tokens/verify",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   CloudflareClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
			},
		}, nil
	})
}
