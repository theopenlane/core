package githubapp

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the GitHub App definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		app := App{Config: cfg}

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         GitHubAppCredential,
					Name:        "GitHub App Credential",
					Description: "Auth-managed credential slot used by the GitHub App client in this definition.",
				},
			},
			Installation: Installation.Registration(),
			Auth: &types.AuthRegistration{
				StartPath:    "/v1/integrations/github/app/install",
				CallbackPath: "/v1/integrations/github/app/callback",
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            GitHubClient.ID(),
					CredentialRefs: []types.CredentialRef{GitHubAppCredential},
					Description:    "GitHub GraphQL client",
					Build:          Client{APIURL: cfg.APIURL}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate the GitHub App installation is reachable",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   GitHubClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        RepositorySyncOperation.Name(),
					Description: "Collect repository inventory from the installation",
					Topic:       RepositorySyncOperation.Topic(Slug),
					ClientRef:   GitHubClient.ID(),
					Handle:      RepositorySync{}.Handle(),
				},
				{
					Name:         VulnerabilityCollectOperation.Name(),
					Description:  "Collect vulnerability alerts from the installation",
					Topic:        VulnerabilityCollectOperation.Topic(Slug),
					ClientRef:    GitHubClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[VulnerabilityCollectConfig](),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: VulnerabilityCollect{}.IngestHandle(),
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Collect organization members as directory accounts",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   GitHubClient.ID(),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
						},
					},
					IngestHandle: DirectorySync{}.IngestHandle(),
				},
			},
			Mappings: githubAppMappings(),
			Webhooks: []types.WebhookRegistration{
				{
					Name:   InstallationEventsWebhook.Name(),
					Verify: app.Verify,
					Event:  app.Event,
					Events: []types.WebhookEventRegistration{
						{
							Name:   PingWebhookEvent.Name(),
							Topic:  PingWebhookEvent.Topic(Slug),
							Handle: PingWebhook{}.Handle,
						},
						{
							Name:   InstallationCreatedWebhookEvent.Name(),
							Topic:  InstallationCreatedWebhookEvent.Topic(Slug),
							Handle: InstallationCreatedWebhook{}.Handle,
						},
						{
							Name:  DependabotAlertWebhookEvent.Name(),
							Topic: DependabotAlertWebhookEvent.Topic(Slug),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: DependabotAlertWebhook{}.Handle,
						},
						{
							Name:  CodeScanningAlertWebhookEvent.Name(),
							Topic: CodeScanningAlertWebhookEvent.Topic(Slug),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: CodeScanningAlertWebhook{}.Handle,
						},
						{
							Name:  SecretScanningAlertWebhookEvent.Name(),
							Topic: SecretScanningAlertWebhookEvent.Topic(Slug),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: SecretScanningAlertWebhook{}.Handle,
						},
					},
				},
			},
		}, nil
	})
}
