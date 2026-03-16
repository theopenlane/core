package githubapp

import (
	"context"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`

	// RepositoryFilter limits repository collection to matching repositories
	RepositoryFilter string `json:"repositoryFilter,omitempty" jsonschema:"title=Repository Filter Expression"`

	// SecurityOnly limits collection to security-focused data
	SecurityOnly bool `json:"securityOnly,omitempty" jsonschema:"title=Collect Security Signals Only"`
}

// Builder returns the GitHub App definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Auth: &types.AuthRegistration{
				StartPath:    "/v1/integrations/github/app/install",
				CallbackPath: "/v1/integrations/github/app/callback",
				Start:        Auth{Config: cfg}.Start,
				Complete:     Auth{Config: cfg}.Complete,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         GitHubClient.ID(),
					Description: "GitHub GraphQL client",
					Build:       Client{Config: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate the GitHub App installation is reachable",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   GitHubClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{Config: cfg}),
				},
				{
					Name:        RepositorySyncOperation.Name(),
					Description: "Collect repository inventory from the installation",
					Topic:       RepositorySyncOperation.Topic(Slug),
					ClientRef:   GitHubClient.ID(),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      RepositorySync{}.Handle(Client{Config: cfg}),
				},
				{
					Name:         VulnerabilityCollectOperation.Name(),
					Description:  "Collect vulnerability alerts from the installation",
					Topic:        VulnerabilityCollectOperation.Topic(Slug),
					ClientRef:    GitHubClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[VulnerabilityCollectConfig](),
					Policy:       types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Ingest: []types.IngestContract{
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaVulnerability,
							EnsurePayloads: true,
						},
					},
					Handle: VulnerabilityCollect{}.Handle(Client{Config: cfg}),
				},
			},
			Mappings: githubAppMappings(),
			Webhooks: []types.WebhookRegistration{
				{
					Name:   "installation.events",
					Verify: Webhook{Config: cfg}.Verify,
				},
			},
		}, nil
	})
}
