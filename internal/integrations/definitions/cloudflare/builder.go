package cloudflare

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
	// ZoneIDs limits collection to specific Cloudflare zone identifiers
	ZoneIDs []string `json:"zoneIds,omitempty" jsonschema:"title=Zone IDs"`
}

// credential holds the Cloudflare API credentials for one installation
type CredentialSchema struct {
	APIToken  string   `json:"apiToken"          jsonschema:"required,title=API Token"`
	AccountID string   `json:"accountId"         jsonschema:"required,title=Account ID"`
	ZoneIDs   []string `json:"zoneIds,omitempty" jsonschema:"title=Zone IDs"`
}

// Builder returns the Cloudflare definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
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
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
			},
		}, nil
	})
}
