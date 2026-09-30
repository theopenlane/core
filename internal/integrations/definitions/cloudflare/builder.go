package cloudflare

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Cloudflare definition builder with the supplied runtime config applied
func Builder(runtime *RuntimeConfig) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "Cloudflare",
				DisplayName: "Cloudflare",
				Description: "Perform directory sync and asset collection from Cloudflare.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/cloudflare",
				Tags:        []string{"directory", "assets"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[RuntimeConfig](),
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  cloudflareClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				cloudflareCredential.Registration(types.CredentialRegistration{
					Name:        "Cloudflare API Credential",
					Description: "API token used to access Cloudflare account and zone data.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				cloudflareConnection.Registration(types.ConnectionRegistration{
					Name:        "Cloudflare API Token",
					Description: "Configure Cloudflare access using an API token scoped to your account and zones.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Cloudflare dashboard.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				cloudflareClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Cloudflare REST API client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description:         "Collect account members as directory accounts",
					Policy:              types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Ingest:              providerkit.DirectoryIngestContracts(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"Account Settings Read", "Access: Users Read", "Access: Groups Read", "Access: Organizations, Identity Providers, and Groups Read"},
					Schedule:            gala.NewFullFetchSchedule(),
				}),
				findingsSyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description: "Collect Cloudflare Security Center insights as findings",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaFinding.Name,
						},
					},
					RequiredPermissions: []string{"Account Security Center Insights Read"},
				}),
				assetSyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description: "Collect Cloudflare domain registrations as assets",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaAsset.Name,
						},
					},
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"Registrar Domains Read"},
					Schedule: &gala.Schedule{
						MinInterval:        assetSyncMinIntervalHours * time.Hour,
						MaxInterval:        assetSyncMaxIntervalDays * assetSyncMinIntervalHours * time.Hour,
						HighDriftThreshold: gala.FullHighDriftThreshold,
					},
				}),
				DomainScanSubmitOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Submit domains to Cloudflare's URL Scanner for scanning",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanPollOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Poll a previously submitted Cloudflare URL Scanner result",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanEnrichmentOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Gather company profile, compliance, and DNS vendor data for a domain",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanBuildReportOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Build the onboarding domain scan report from a completed URL Scanner result and gathered enrichment",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanRequestOp.Registration(DefinitionID, types.OperationRegistration{
					Description:           "Request a domain scan for a single domain",
					Policy:                types.ExecutionPolicy{Inline: true, SkipRunRecord: true},
					DisabledForAll:        !runtime.Provisioned(),
					RateLimit:             &types.RateLimitPolicy{Window: time.Hour},
					CustomerSelectable:    lo.ToPtr(false),
					RequiresPaymentMethod: true,
				}),
				DomainScanImportOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Import a reviewer-accepted domain scan report into real records",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
			},
			GalaListeners: []types.GalaListenerRegistration{
				domainScanListeners(),
			},
			Mappings: append(providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
				providerkit.FindingMapping(mapExprFinding),
				types.MappingRegistration{
					Schema: entityops.SchemaAsset.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprAsset,
					},
				},
			),
		}

		if runtime.Provisioned() {
			config, err := jsonx.ToRawMessage(runtime)
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Schema: jsonx.SchemaFrom[RuntimeConfig](),
				Config: config,
				Build:  runtimeCloudflareClientBuilder(),
			}
		}

		return def, nil
	})
}
