package githubapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
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
					Ref:         gitHubAppCredential.ID(),
					Name:        "GitHub App Credential",
					Description: "Auth-managed credential slot used by the GitHub App client in this definition.",
					Schema:      gitHubAppCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       gitHubAppCredential.ID(),
					Name:                "GitHub App installation",
					Description:         "Install the Openlane GitHub App into a GitHub organization or repository owner account.",
					CredentialRefs:      []types.CredentialSlotID{gitHubAppCredential.ID()},
					ClientRefs:          []types.ClientID{gitHubClient.ID()},
					ValidationOperation: healthDefaultOperation.Name(),
					Installation:        installation.Registration(),
					Auth: &types.AuthRegistration{
						CredentialRef: gitHubAppCredential.ID(),
						Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
							return startAppInstall(cfg)
						},
						Complete: func(ctx context.Context, state json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
							return completeAppInstall(ctx, cfg, state, input)
						},
						Refresh: func(ctx context.Context, credential types.CredentialSet) (types.CredentialSet, error) {
							return refreshAppInstall(ctx, cfg, credential)
						},
						TokenView: tokenViewAppInstall,
					},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: gitHubAppCredential.ID(),
						Name:          "Disconnect GitHub App installation",
						Description:   "Open the GitHub installation settings page and uninstall the Openlane GitHub App. Openlane will remove the installation after GitHub sends the uninstall webhook.",
						Disconnect: func(_ context.Context, req types.DisconnectRequest) (types.DisconnectResult, error) {
							installationID, err := disconnectInstallationID(req)
							if err != nil {
								return types.DisconnectResult{}, err
							}

							details, err := jsonx.ToRawMessage(disconnectDetails{
								InstallationID: strconv.FormatInt(installationID, 10),
							})
							if err != nil {
								return types.DisconnectResult{}, ErrInstallationMetadataEncode
							}

							return types.DisconnectResult{
								RedirectURL:      fmt.Sprintf("https://github.com/settings/installations/%d", installationID),
								Message:          "Uninstall the Openlane GitHub App in GitHub to finish disconnecting this integration.",
								Details:          details,
								SkipLocalCleanup: true,
							}, nil
						},
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            gitHubClient.ID(),
					CredentialRefs: []types.CredentialSlotID{gitHubAppCredential.ID()},
					Description:    "GitHub GraphQL client",
					Build:          Client{APIURL: cfg.APIURL}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthDefaultOperation.Name(),
					Description:  "Validate the GitHub App installation is reachable",
					Topic:        types.OperationTopic(DefinitionID.ID(), healthDefaultOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         repositorySyncOperation.Name(),
					Description:  "Collect repository inventory from the installation as assets",
					Topic:        types.OperationTopic(DefinitionID.ID(), repositorySyncOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					ConfigSchema: repositorySyncSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaAsset,
						},
					},
					IngestHandle: RepositorySync{}.IngestHandle(),
				},
				{
					Name:         vulnerabilityCollectOperation.Name(),
					Description:  "Collect vulnerability alerts from the installation",
					Topic:        types.OperationTopic(DefinitionID.ID(), vulnerabilityCollectOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					ConfigSchema: vulnerabilityCollectSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: VulnerabilityCollect{}.IngestHandle(),
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect organization members as directory accounts",
					Topic:        types.OperationTopic(DefinitionID.ID(), directorySyncOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					ConfigSchema: directorySyncSchema,
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
							Name:   pingWebhookEvent.Name(),
							Topic:  types.WebhookEventTopic(DefinitionID.ID(), pingWebhookEvent.Name()),
							Handle: PingWebhook{}.Handle,
						},
						{
							Name:   installationCreatedWebhookEvent.Name(),
							Topic:  types.WebhookEventTopic(DefinitionID.ID(), installationCreatedWebhookEvent.Name()),
							Handle: InstallationCreatedWebhook{}.Handle,
						},
						{
							Name:   installationDeletedWebhookEvent.Name(),
							Topic:  types.WebhookEventTopic(DefinitionID.ID(), installationDeletedWebhookEvent.Name()),
							Handle: InstallationDeletedWebhook{}.Handle,
						},
						{
							Name:  dependabotAlertWebhookEvent.Name(),
							Topic: types.WebhookEventTopic(DefinitionID.ID(), dependabotAlertWebhookEvent.Name()),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: DependabotAlertWebhook{}.Handle,
						},
						{
							Name:  codeScanningAlertWebhookEvent.Name(),
							Topic: types.WebhookEventTopic(DefinitionID.ID(), codeScanningAlertWebhookEvent.Name()),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: CodeScanningAlertWebhook{}.Handle,
						},
						{
							Name:  secretScanningAlertWebhookEvent.Name(),
							Topic: types.WebhookEventTopic(DefinitionID.ID(), secretScanningAlertWebhookEvent.Name()),
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
