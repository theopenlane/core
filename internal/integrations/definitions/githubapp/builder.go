package githubapp

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
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
			ID:          DefinitionID.ID(),
			Family:      "GitHub",
			DisplayName: "GitHub App",
			Description: "Install the Openlane GitHub App to collect repository metadata and security alerts",
			Category:    "source-control",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/github_app",
			Tags:        []string{"vulnerabilities", "assets", "directory"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			Installation: installation.Registration(),
			Connections: []types.Connector{
				appInstall.
					Name("GitHub App installation").
					Description("Install the Openlane GitHub App into your GitHub organization.").
					Authenticates(appInstallFlow(cfg)).
					Provides(appClient(cfg)).
					Verified(verify).
					Disconnects("Uninstall the Openlane GitHub App from your GitHub organization settings. Openlane will complete the removal after GitHub confirms the uninstall.", disconnectApp),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[RepositorySync]().
					Ingests(runRepositorySync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					SkipDefaultLookback().
					Description("Collect repository inventory from the installation as assets").
					Registration(),
				types.OperationRefOf[VulnerabilitySync]().
					Ingests(runVulnerabilityCollect).
					// TODO: remove with providerkit.UpgradeFromSection once every installation has been upgraded off main's client config
					Upgraded(providerkit.UpgradeFromSection[VulnerabilitySync](mainFindingSyncKey)).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaVulnerability.Name}).
					Description("Collect vulnerability alerts from the installation").
					Registration(),
				types.OperationRefOf[providerkit.DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Description("Collect organization members, teams, and team memberships").
					Registration(),
			},
			Mappings: append([]types.MappingRegistration{
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
			}, providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership)...),
			Webhooks: []types.WebhookRegistration{
				InstallationEventsWebhook.Registration(types.WebhookRegistration{
					StaticRoute:        "/github/app/webhook",
					SecretSource:       func() string { return cfg.WebhookSecret },
					ResolveIntegration: ResolveWebhookIntegration,
					Verify:             app.Verify,
					Event:              app.Event,
					Events: []types.WebhookEventRegistration{
						pingWebhookEvent.Registration(types.WebhookEventRegistration{
							Handle: PingWebhook{}.Handle,
						}),
						installationCreatedWebhookEvent.Registration(types.WebhookEventRegistration{
							Handle: InstallationCreatedWebhook{}.Handle,
						}),
						installationDeletedWebhookEvent.Registration(types.WebhookEventRegistration{
							Handle: InstallationDeletedWebhook{}.Handle,
						}),
						dependabotAlertWebhookEvent.Registration(types.WebhookEventRegistration{
							Ingest: []types.IngestContract{
								{
									Schema: entityops.SchemaVulnerability.Name,
								},
							},
							Handle: DependabotAlertWebhook{}.Handle,
						}),
						codeScanningAlertWebhookEvent.Registration(types.WebhookEventRegistration{
							Ingest: []types.IngestContract{
								{
									Schema: entityops.SchemaVulnerability.Name,
								},
							},
							Handle: CodeScanningAlertWebhook{}.Handle,
						}),
						secretScanningAlertWebhookEvent.Registration(types.WebhookEventRegistration{
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
