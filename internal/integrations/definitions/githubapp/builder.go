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
				Family:      "GitHub",
				DisplayName: "GitHub App",
				Description: "Install the Openlane GitHub App to collect repository metadata and security alerts",
				Category:    "source-control",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/github_app/overview",
				Tags:        []string{"vulnerabilities", "assets", "directory"},
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
					Description: "Integration credential managed by the GitHub App install flow.",
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       gitHubAppCredential.ID(),
					Name:                "GitHub App installation",
					Description:         "Install the Openlane GitHub App into your GitHub organization.",
					CredentialRefs:      []types.CredentialSlotID{gitHubAppCredential.ID()},
					ClientRefs:          []types.ClientID{gitHubClient.ID()},
					ValidationOperation: healthDefaultOperation.Name(),
					Integration:         installation.Registration(),
					Auth: &types.AuthRegistration{
						CredentialRef: gitHubAppCredential.ID(),
						Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
							return startAppInstall(cfg)
						},
						Complete: func(ctx context.Context, state json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
							return completeAppInstall(ctx, cfg, state, input)
						},
					},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: gitHubAppCredential.ID(),
						Description:   "Uninstall the Openlane GitHub App from your GitHub organization settings. Openlane will complete the removal after GitHub confirms the uninstall.",
						Disconnect: func(_ context.Context, req types.DisconnectRequest) (types.DisconnectResult, error) {
							integrationID, err := disconnectInstallationID(req)
							if err != nil {
								return types.DisconnectResult{}, err
							}

							details, err := jsonx.ToRawMessage(disconnectDetails{
								InstallationID: strconv.FormatInt(integrationID, 10),
							})
							if err != nil {
								return types.DisconnectResult{}, ErrInstallationMetadataEncode
							}

							return types.DisconnectResult{
								RedirectURL:      fmt.Sprintf("https://github.com/settings/installations/%d", integrationID),
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
					Build:          Client{AppConfig: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthDefaultOperation.Name(),
					Description:  "Validate the GitHub App installation is reachable",
					Topic:        DefinitionID.OperationTopic(healthDefaultOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         repositorySyncOperation.Name(),
					Description:  "Collect repository inventory from the installation as assets",
					Topic:        DefinitionID.OperationTopic(repositorySyncOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					ConfigSchema: repositorySyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Disabled:     providerkit.DisabledWhen(func(u UserInput) bool { return u.RepositorySync.Disable }),
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
					Topic:        DefinitionID.OperationTopic(vulnerabilityCollectOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					ConfigSchema: vulnerabilityCollectSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Disabled:     providerkit.DisabledWhen(func(u UserInput) bool { return u.VulnerabilitySync.Disable }),
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
					Topic:        DefinitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:    gitHubClient.ID(),
					ConfigSchema: directorySyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Disabled:     providerkit.DisabledWhen(func(u UserInput) bool { return u.DirectorySync.Disable }),
					Ingest: []types.IngestContract{
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
							ExhaustiveSync: true,
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
							Topic:  DefinitionID.WebhookEventTopic(pingWebhookEvent.Name()),
							Handle: PingWebhook{}.Handle,
						},
						{
							Name:   installationCreatedWebhookEvent.Name(),
							Topic:  DefinitionID.WebhookEventTopic(installationCreatedWebhookEvent.Name()),
							Handle: InstallationCreatedWebhook{}.Handle,
						},
						{
							Name:   installationDeletedWebhookEvent.Name(),
							Topic:  DefinitionID.WebhookEventTopic(installationDeletedWebhookEvent.Name()),
							Handle: InstallationDeletedWebhook{}.Handle,
						},
						{
							Name:  dependabotAlertWebhookEvent.Name(),
							Topic: DefinitionID.WebhookEventTopic(dependabotAlertWebhookEvent.Name()),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: DependabotAlertWebhook{}.Handle,
						},
						{
							Name:  codeScanningAlertWebhookEvent.Name(),
							Topic: DefinitionID.WebhookEventTopic(codeScanningAlertWebhookEvent.Name()),
							Ingest: []types.IngestContract{
								{
									Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
								},
							},
							Handle: CodeScanningAlertWebhook{}.Handle,
						},
						{
							Name:  secretScanningAlertWebhookEvent.Name(),
							Topic: DefinitionID.WebhookEventTopic(secretScanningAlertWebhookEvent.Name()),
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
