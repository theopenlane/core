package cloudflare

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck identifies the default health check operation
type HealthCheck struct{}

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0CFLARE00000000000000001")
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
)

const Slug = "cloudflare"

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
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

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
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         clientRef.ID(),
					Description: "Cloudflare REST API client",
					Build:       buildCloudflareClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify Cloudflare API token via /user/tokens/verify",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
			},
		}, nil
	})
}
