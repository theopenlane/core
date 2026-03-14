package vercel

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label     string `json:"label,omitempty"     jsonschema:"title=Installation Label"`
	TeamID    string `json:"teamId,omitempty"    jsonschema:"title=Team ID"`
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID"`
}

// credential holds the Vercel API credentials for one installation
type credential struct {
	APIToken  string `json:"apiToken"            jsonschema:"required,title=API Token"`
	TeamID    string `json:"teamId,omitempty"    jsonschema:"title=Team ID"`
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID"`
}

var (
	definitionSpec   = types.DefinitionSpec{
		ID:          "def_01K0VERCEL00000000000000001",
		Slug:        "vercel",
		Version:     "v1",
		Family:      "vercel",
		DisplayName: "Vercel",
		Description: "Collect Vercel project and deployment context to support devops posture and drift detection workflows.",
		Category:    "devops",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/vercel/overview",
		Labels:      map[string]string{"vendor": "vercel", "product": "deployment"},
		Active:      false,
		Visible:     true,
	}

	userInputSchema  = providerkit.SchemaFrom[userInput]()
	credentialSchema = providerkit.SchemaFrom[credential]()
)

// def implements definition.Assembler for the Vercel integration
type def struct{}

// Builder returns the Vercel definition builder
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
			Description: "Vercel REST API client",
			Build:       buildVercelClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "health.default",
			Kind:        types.OperationKindHealth,
			Description: "Call Vercel /v2/user to verify token and account",
			Topic:       gala.TopicName("integration.vercel.health.default"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      runHealthOperation,
		},
		{
			Name:        "projects.sample",
			Kind:        types.OperationKindCollect,
			Description: "Collect a sample of Vercel projects for drift detection",
			Topic:       gala.TopicName("integration.vercel.projects.sample"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      runProjectsSampleOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration { return nil }
func (d *def) Webhooks() []types.WebhookRegistration { return nil }
