package buildkite

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label            string `json:"label,omitempty"            jsonschema:"title=Installation Label"`
	OrganizationSlug string `json:"organizationSlug,omitempty" jsonschema:"title=Organization Slug"`
	TeamSlug         string `json:"teamSlug,omitempty"         jsonschema:"title=Team Slug"`
}

// credential holds the Buildkite API credentials for one installation
type credential struct {
	APIToken         string `json:"apiToken"               jsonschema:"required,title=API Token"`
	OrganizationSlug string `json:"organizationSlug"       jsonschema:"required,title=Organization Slug"`
	TeamSlug         string `json:"teamSlug,omitempty"     jsonschema:"title=Team Slug"`
	PipelineSlug     string `json:"pipelineSlug,omitempty" jsonschema:"title=Pipeline Slug"`
}

var (
	definitionSpec   = types.DefinitionSpec{
		ID:          "def_01K0BKITE000000000000000001",
		Slug:        "buildkite",
		Version:     "v1",
		Family:      "buildkite",
		DisplayName: "Buildkite",
		Description: "Collect Buildkite organization and pipeline context to support CI security and compliance posture reporting.",
		Category:    "ci",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/buildkite/overview",
		Labels:      map[string]string{"vendor": "buildkite", "product": "pipelines"},
		Active:      false,
		Visible:     true,
	}

	userInputSchema  = providerkit.SchemaFrom[userInput]()
	credentialSchema = providerkit.SchemaFrom[credential]()
)

// def implements definition.Assembler for the Buildkite integration
type def struct{}

// Builder returns the Buildkite definition builder
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
			Description: "Buildkite REST API client",
			Build:       buildBuildkiteClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "health.default",
			Kind:        types.OperationKindHealth,
			Description: "Validate Buildkite token by calling the /v2/user endpoint",
			Topic:       gala.TopicName("integration.buildkite.health.default"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      runHealthOperation,
		},
		{
			Name:        "organizations.collect",
			Kind:        types.OperationKindCollect,
			Description: "Collect Buildkite organizations for reporting",
			Topic:       gala.TopicName("integration.buildkite.organizations.collect"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      runOrganizationsCollectOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration { return nil }
func (d *def) Webhooks() []types.WebhookRegistration { return nil }
