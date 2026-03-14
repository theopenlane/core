package cloudflare

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label   string   `json:"label,omitempty"   jsonschema:"title=Installation Label"`
	ZoneIDs []string `json:"zoneIds,omitempty" jsonschema:"title=Zone IDs"`
}

// credential holds the Cloudflare API credentials for one installation
type credential struct {
	APIToken  string   `json:"apiToken"          jsonschema:"required,title=API Token"`
	AccountID string   `json:"accountId"         jsonschema:"required,title=Account ID"`
	ZoneIDs   []string `json:"zoneIds,omitempty" jsonschema:"title=Zone IDs"`
}

// Builder returns the Cloudflare definition builder
func Builder() definition.Builder {
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0CFLARE00000000000000001",
				Slug:        "cloudflare",
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
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema:   providerkit.SchemaFrom[credential](),
				Persist:  types.CredentialPersistModeKeystore,
				Validate: providerkit.ValidateAPIKeyCredential(),
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "api",
					Description: "Cloudflare REST API client",
					Build:       buildCloudflareClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Verify Cloudflare API token via /user/tokens/verify",
					Topic:       gala.TopicName("integration.cloudflare.health.default"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
			},
		}, nil
	})
}
