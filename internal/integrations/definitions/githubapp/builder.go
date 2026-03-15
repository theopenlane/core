package githubapp

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

const (
	// DefinitionID is the canonical opaque identifier for the GitHub App definition
	DefinitionID = "def_01K0GHAPP000000000000000001"
	// DefinitionSlug is the human-readable slug for the GitHub App definition
	DefinitionSlug = "github_app"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label            string `json:"label,omitempty"            jsonschema:"title=Installation Label"`
	RepositoryFilter string `json:"repositoryFilter,omitempty" jsonschema:"title=Repository Filter Expression"`
	SecurityOnly     bool   `json:"securityOnly,omitempty"     jsonschema:"title=Collect Security Signals Only"`
}

// def holds operator config for the GitHub App integration
type def struct {
	cfg Config
}

// Builder returns the GitHub App definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
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
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Auth: &types.AuthRegistration{
				Start:    d.startInstallAuth,
				Complete: d.completeInstallAuth,
			},
			Clients: []types.ClientRegistration{
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
			},
			Operations: []types.OperationRegistration{
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
			},
			Mappings: githubAppMappings(),
			Webhooks: []types.WebhookRegistration{
				{
					Name:    "installation.events",
					Verify:  d.verifyWebhook,
					Resolve: d.resolveWebhook,
					Handle:  d.handleWebhook,
				},
			},
		}, nil
	})
}
