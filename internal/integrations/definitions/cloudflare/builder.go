package cloudflare

import (
	"fmt"
	"time"

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
		domainScanRequestOp := DomainScanRequestOp.DisabledForAll(!runtime.Provisioned())

		def := types.Definition{
			ID:          DefinitionID.ID(),
			Family:      "Cloudflare",
			DisplayName: "Cloudflare",
			Description: "Perform directory sync and asset collection from Cloudflare.",
			Category:    "security-posture",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/cloudflare",
			Tags:        []string{"directory", "assets"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[RuntimeConfig](),
			},
			HealthCheck:  cloudflareClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				cloudflareCredential.Registration(types.CredentialRegistration{
					Name:        "Cloudflare API Credential",
					Description: "API token used to access Cloudflare account and zone data.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: cloudflareCredential.ID(),
					Name:          "Cloudflare API Token",
					Description:   "Configure Cloudflare access using an API token scoped to your account and zones.",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: cloudflareCredential.ID(),
						Description:   "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Cloudflare dashboard.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				cloudflareClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Cloudflare REST API client",
				}),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[DirectorySync]().
					Ingests(cloudflareClient, runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("Account Settings Read", "Access: Users Read", "Access: Groups Read", "Access: Organizations, Identity Providers, and Groups Read").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Registration(DefinitionID, types.OperationRegistration{
						Description: "Collect account members as directory accounts",
					}),
				types.OperationRefOf[FindingsSync]().
					Ingests(cloudflareClient, runFindingsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaFinding.Name}).
					Permissions("Account Security Center Insights Read").
					Registration(DefinitionID, types.OperationRegistration{
						Description: "Collect Cloudflare Security Center insights as findings",
					}),
				types.OperationRefOf[AssetSync]().
					Ingests(cloudflareClient, runAssetCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					Permissions("Registrar Domains Read").
					Schedule(&gala.Schedule{
						MinInterval:        assetSyncMinIntervalHours * time.Hour,
						MaxInterval:        assetSyncMaxIntervalDays * assetSyncMinIntervalHours * time.Hour,
						HighDriftThreshold: gala.FullHighDriftThreshold,
					}).
					SkipDefaultLookback().
					Registration(DefinitionID, types.OperationRegistration{
						Description: "Collect Cloudflare domain registrations as assets",
					}),
				DomainScanSubmitOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Submit domains to Cloudflare's URL Scanner for scanning",
				}),
				DomainScanPollOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Poll a previously submitted Cloudflare URL Scanner result",
				}),
				DomainScanEnrichmentOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Gather company profile, compliance, and DNS vendor data for a domain",
				}),
				DomainScanBuildReportOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Build the onboarding domain scan report from a completed URL Scanner result and gathered enrichment",
				}),
				domainScanRequestOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Request a domain scan for a single domain",
				}),
				DomainScanImportOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Import a reviewer-accepted domain scan report into real records",
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
