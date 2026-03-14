package githubapp

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label            string `json:"label,omitempty"            jsonschema:"title=Installation Label"`
	RepositoryFilter string `json:"repositoryFilter,omitempty" jsonschema:"title=Repository Filter Expression"`
	SecurityOnly     bool   `json:"securityOnly,omitempty"     jsonschema:"title=Collect Security Signals Only"`
}

var (
	definitionSpec = types.DefinitionSpec{
		ID:          "def_01K0GHAPP000000000000000001",
		Slug:        "github_app",
		Version:     "v1",
		Family:      "github",
		DisplayName: "GitHub App",
		Description: "Install the Openlane GitHub App to collect repository metadata and security alerts",
		Category:    "code",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/github_app/overview",
		Labels:      map[string]string{"vendor": "github"},
		Active:      true,
		Visible:     true,
	}

	configSchema    = providerkit.SchemaFrom[Config]()
	userInputSchema = providerkit.SchemaFrom[userInput]()
)

// def implements definition.Assembler for the GitHub App integration
type def struct {
	cfg Config
}

// Builder returns the GitHub App definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.FromAssembler(&def{cfg: cfg})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration {
	return &types.OperatorConfigRegistration{Schema: configSchema}
}

func (d *def) UserInput() *types.UserInputRegistration {
	return &types.UserInputRegistration{Schema: userInputSchema}
}

func (d *def) Credentials() *types.CredentialRegistration { return nil }

func (d *def) Auth() *types.AuthRegistration {
	return &types.AuthRegistration{
		Start:    d.startInstallAuth,
		Complete: d.completeInstallAuth,
	}
}

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "rest",
			Description: "GitHub REST client",
			Build:       d.buildRESTClient,
		},
		{
			Name:        "graphql",
			Description: "GitHub GraphQL client",
			Build:       d.buildGraphQLClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "health.default",
			Kind:        types.OperationKindHealth,
			Description: "Validate the GitHub App installation is reachable",
			Topic:       gala.TopicName("integration.github_app.health.default"),
			Client:      "rest",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      d.runHealthOperation,
		},
		{
			Name:        "repository.sync",
			Kind:        types.OperationKindSync,
			Description: "Collect repository inventory from the installation",
			Topic:       gala.TopicName("integration.github_app.repository.sync"),
			Client:      "rest",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      d.runRepositorySyncOperation,
		},
		{
			Name:        "vulnerability.collect",
			Kind:        types.OperationKindCollect,
			Description: "Collect vulnerability data from the installation",
			Topic:       gala.TopicName("integration.github_app.vulnerability.collect"),
			Client:      "rest",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      d.runVulnerabilityCollectionOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration {
	return githubAppMappings()
}

func (d *def) Webhooks() []types.WebhookRegistration {
	return []types.WebhookRegistration{
		{
			Name:    "installation.events",
			Verify:  d.verifyWebhook,
			Resolve: d.resolveWebhook,
			Handle:  d.handleWebhook,
		},
	}
}
