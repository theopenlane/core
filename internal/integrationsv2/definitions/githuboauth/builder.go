package githuboauth

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
		ID:          "def_01K0GHOAUTH0000000000000001",
		Slug:        "github_oauth",
		Version:     "v1",
		Family:      "github",
		DisplayName: "GitHub",
		Description: "Collect GitHub repository metadata and security alerts (Dependabot, code scanning, and secret scanning) for exposure management.",
		Category:    "code",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/github/overview",
		Labels:      map[string]string{"vendor": "github"},
		Active:      true,
		Visible:     false,
	}

	configSchema    = providerkit.SchemaFrom[Config]()
	userInputSchema = providerkit.SchemaFrom[userInput]()
)

// def implements definition.Assembler for the GitHub OAuth integration
type def struct {
	cfg Config
}

// Builder returns the GitHub OAuth definition builder with the supplied operator config applied
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
		Start:    startInstallAuth,
		Complete: completeInstallAuth,
	}
}

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "rest",
			Description: "GitHub REST API client",
			Build:       buildRESTClient,
		},
		{
			Name:        "graphql",
			Description: "GitHub GraphQL API client",
			Build:       buildGraphQLClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "health.default",
			Kind:        types.OperationKindHealth,
			Description: "Validate GitHub OAuth token by calling the /user endpoint",
			Topic:       gala.TopicName("integration.github_oauth.health.default"),
			Client:      "rest",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      runHealthOperation,
		},
		{
			Name:        "repository.sync",
			Kind:        types.OperationKindSync,
			Description: "Collect repository metadata for the authenticated account",
			Topic:       gala.TopicName("integration.github_oauth.repository.sync"),
			Client:      "rest",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      runRepositorySyncOperation,
		},
		{
			Name:        "vulnerability.collect",
			Kind:        types.OperationKindCollect,
			Description: "Collect GitHub vulnerability alerts (Dependabot, code scanning, secret scanning) for accessible repositories",
			Topic:       gala.TopicName("integration.github_oauth.vulnerability.collect"),
			Client:      "rest",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      runVulnerabilityCollectOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema:  "repository",
			Variant: "default",
			Spec:    types.MappingOverride{Version: "v1"},
		},
		{
			Schema:  "vulnerability",
			Variant: "default",
			Spec:    types.MappingOverride{Version: "v1"},
		},
	}
}

func (d *def) Webhooks() []types.WebhookRegistration { return nil }
