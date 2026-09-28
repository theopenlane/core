package githubapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/github_app",
				Tags:        []string{"vulnerabilities", "assets", "directory"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  gitHubClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				gitHubAppCredential.Registration(types.CredentialRegistration{
					Name:        "GitHub App Credential",
					Description: "Integration credential managed by the GitHub App install flow.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				gitHubAppConnection.Registration(types.ConnectionRegistration{
					Name:        "GitHub App installation",
					Description: "Install the Openlane GitHub App into your GitHub organization.",
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
						Description: "Uninstall the Openlane GitHub App from your GitHub organization settings. Openlane will complete the removal after GitHub confirms the uninstall.",
						Disconnect: func(ctx context.Context, req types.DisconnectRequest) (types.DisconnectResult, error) {
							integrationID, name, err := disconnectInstallationID(ctx, req)
							if err != nil {
								return types.DisconnectResult{}, err
							}

							details, err := jsonx.ToRawMessage(disconnectDetails{
								InstallationID:   strconv.FormatInt(integrationID, 10),
								OrganizationName: name,
							})
							if err != nil {
								return types.DisconnectResult{}, ErrInstallationMetadataEncode
							}

							url := fmt.Sprintf("https://github.com/settings/installations/%d", integrationID)
							if name != "" {
								url = fmt.Sprintf("https://github.com/organizations/%s/settings/installations/%d", name, integrationID)
							}

							return types.DisconnectResult{
								RedirectURL: url,
								Message:     "Uninstall the Openlane GitHub App in GitHub to finish disconnecting this integration.",
								Details:     details,
							}, nil
						},
					},
				}),
			},
			Clients: []types.ClientRegistration{
				gitHubClient.Registration(Client{AppConfig: cfg}.Build, types.ClientRegistration{
					Description: "GitHub GraphQL client",
				}),
			},
			Operations: []types.OperationRegistration{
				repositorySyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description: "Collect repository inventory from the installation as assets",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaAsset.Name,
						},
					},
					SkipDefaultLookback: true,
				}),
				vulnerabilityCollectOperation.Registration(DefinitionID, types.OperationRegistration{
					Description: "Collect vulnerability alerts from the installation",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaVulnerability.Name,
						},
					},
				}),
				directorySyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description: "Collect organization members, teams, and team memberships",
					Policy:      types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaDirectoryAccount.Name,
						},
						{
							Schema: entityops.SchemaDirectoryGroup.Name,
						},
						{
							Schema: entityops.SchemaDirectoryMembership.Name,
						},
					},
					SkipDefaultLookback: true,
					Schedule:            gala.NewFullFetchSchedule(),
				}),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema:  entityops.SchemaVulnerability.Name,
					Variant: githubAlertTypeDependabot,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDependabot,
					},
				},
				{
					Schema:  entityops.SchemaVulnerability.Name,
					Variant: githubAlertTypeDependabotPoll,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDependabotPoll,
					},
				},
				{
					Schema:  entityops.SchemaVulnerability.Name,
					Variant: githubAlertTypeCodeScanning,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprCodeScanning,
					},
				},
				{
					Schema:  entityops.SchemaVulnerability.Name,
					Variant: githubAlertTypeSecretScan,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprSecretScanning,
					},
				},
				{
					Schema:  entityops.SchemaAsset.Name,
					Variant: repositoryAssetVariant,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprRepositoryAsset,
					},
				},
				{
					Schema: entityops.SchemaDirectoryAccount.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryAccount,
					},
				},
				{
					Schema: entityops.SchemaDirectoryGroup.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryGroup,
					},
				},
				{
					Schema: entityops.SchemaDirectoryMembership.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryMembership,
						Links: []types.LinkRule{
							{
								TargetSchema: entityops.SchemaDirectoryAccount.Name,
								TargetField:  directoryaccount.FieldExternalID,
								SourceField:  entityops.DirectoryMembershipFields.DirectoryAccountID.InputKey,
							},
							{
								TargetSchema: entityops.SchemaDirectoryGroup.Name,
								TargetField:  directorygroup.FieldExternalID,
								SourceField:  entityops.DirectoryMembershipFields.DirectoryGroupID.InputKey,
							},
						},
					},
				},
			},
			Webhooks: []types.WebhookRegistration{
				InstallationEventsWebhook.Registration(types.WebhookRegistration{
					StaticRoute:        "/github/app/webhook",
					SecretSource:       func() string { return cfg.WebhookSecret },
					ResolveIntegration: ResolveWebhookIntegration,
					Verify:             app.Verify,
					Event:              app.Event,
					Events: []types.WebhookEventRegistration{
						pingWebhookEvent.Registration(DefinitionID, types.WebhookEventRegistration{
							Handle: PingWebhook{}.Handle,
						}),
						installationCreatedWebhookEvent.Registration(DefinitionID, types.WebhookEventRegistration{
							Handle: InstallationCreatedWebhook{}.Handle,
						}),
						installationDeletedWebhookEvent.Registration(DefinitionID, types.WebhookEventRegistration{
							Handle: InstallationDeletedWebhook{}.Handle,
						}),
						dependabotAlertWebhookEvent.Registration(DefinitionID, types.WebhookEventRegistration{
							Ingest: []types.IngestContract{
								{
									Schema: entityops.SchemaVulnerability.Name,
								},
							},
							Handle: DependabotAlertWebhook{}.Handle,
						}),
						codeScanningAlertWebhookEvent.Registration(DefinitionID, types.WebhookEventRegistration{
							Ingest: []types.IngestContract{
								{
									Schema: entityops.SchemaVulnerability.Name,
								},
							},
							Handle: CodeScanningAlertWebhook{}.Handle,
						}),
						secretScanningAlertWebhookEvent.Registration(DefinitionID, types.WebhookEventRegistration{
							Ingest: []types.IngestContract{
								{
									Schema: entityops.SchemaVulnerability.Name,
								},
							},
							Handle: SecretScanningAlertWebhook{}.Handle,
						}),
					},
				}),
			},
		}, nil
	})
}
