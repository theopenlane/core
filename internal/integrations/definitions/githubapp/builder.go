package githubapp

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the GitHub App definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		app := App{Config: cfg}

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
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
					Ref:         GitHubAppCredential.ID(),
					Name:        "GitHub App Credential",
					Description: "Auth-managed credential slot used by the GitHub App client in this definition.",
					Schema:      gitHubAppCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       GitHubAppCredential.ID(),
					Name:                "GitHub App Installation",
					Description:         "Install the Openlane GitHub App into a GitHub organization or repository owner account.",
					CredentialRefs:      []types.CredentialSlotID{GitHubAppCredential.ID()},
					ClientRefs:          []types.ClientID{GitHubClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Auth:                appInstallAuthRegistration(cfg),
					Disconnect:          appInstallDisconnectRegistration(),
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            GitHubClient.ID(),
					CredentialRefs: []types.CredentialSlotID{GitHubAppCredential.ID()},
					Description:    "GitHub GraphQL client",
					Build:          Client{APIURL: cfg.APIURL}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate the GitHub App installation is reachable",
					Topic:       types.OperationTopic(DefinitionID.ID(), HealthDefaultOperation.Name()),
					ClientRef:   GitHubClient.ID(),
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        RepositorySyncOperation.Name(),
					Description: "Collect repository inventory from the installation as assets",
					Topic:       types.OperationTopic(DefinitionID.ID(), RepositorySyncOperation.Name()),
					ClientRef:   GitHubClient.ID(),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaAsset,
						},
					},
					IngestHandle: RepositorySync{}.IngestHandle(),
				},
				{
					Name:         VulnerabilityCollectOperation.Name(),
					Description:  "Collect vulnerability alerts from the installation",
					Topic:        types.OperationTopic(DefinitionID.ID(), VulnerabilityCollectOperation.Name()),
					ClientRef:    GitHubClient.ID(),
					ConfigSchema: vulnerabilityCollectSchema,
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
					Topic:       types.OperationTopic(DefinitionID.ID(), DirectorySyncOperation.Name()),
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
							Topic:  types.WebhookEventTopic(DefinitionID.ID(), PingWebhookEvent.Name()),
							Handle: PingWebhook{}.Handle,
						},
						{
							Name:   InstallationCreatedWebhookEvent.Name(),
							Topic:  types.WebhookEventTopic(DefinitionID.ID(), InstallationCreatedWebhookEvent.Name()),
							Handle: InstallationCreatedWebhook{}.Handle,
						},
						{
							Name:   InstallationDeletedWebhookEvent.Name(),
							Topic:  types.WebhookEventTopic(DefinitionID.ID(), InstallationDeletedWebhookEvent.Name()),
							Handle: InstallationDeletedWebhook{}.Handle,
						},
						{
							Name:  DependabotAlertWebhookEvent.Name(),
							Topic: types.WebhookEventTopic(DefinitionID.ID(), DependabotAlertWebhookEvent.Name()),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: DependabotAlertWebhook{}.Handle,
						},
						{
							Name:  CodeScanningAlertWebhookEvent.Name(),
							Topic: types.WebhookEventTopic(DefinitionID.ID(), CodeScanningAlertWebhookEvent.Name()),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: CodeScanningAlertWebhook{}.Handle,
						},
						{
							Name:  SecretScanningAlertWebhookEvent.Name(),
							Topic: types.WebhookEventTopic(DefinitionID.ID(), SecretScanningAlertWebhookEvent.Name()),
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
