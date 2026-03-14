package cloudflare

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
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

var (
	definitionSpec   = types.DefinitionSpec{
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
	}

	userInputSchema  = providerkit.SchemaFrom[userInput]()
	credentialSchema = providerkit.SchemaFrom[credential]()
)

// def implements definition.Assembler for the Cloudflare integration
type def struct{}

// Builder returns the Cloudflare definition builder
func Builder() definition.Builder {
	return definition.FromAssembler(&def{})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration { return nil }

func (d *def) UserInput() *types.UserInputRegistration {
	return &types.UserInputRegistration{Schema: userInputSchema}
}

func (d *def) Credentials() *types.CredentialRegistration {
	return &types.CredentialRegistration{
		Schema:  credentialSchema,
		Persist: types.CredentialPersistModeKeystore,
	}
}

func (d *def) Auth() *types.AuthRegistration { return nil }

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "api",
			Description: "Cloudflare REST API client",
			Build:       buildCloudflareClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "health.default",
			Kind:        types.OperationKindHealth,
			Description: "Verify Cloudflare API token via /user/tokens/verify",
			Topic:       gala.TopicName("integration.cloudflare.health.default"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      runHealthOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration { return nil }
func (d *def) Webhooks() []types.WebhookRegistration { return nil }
