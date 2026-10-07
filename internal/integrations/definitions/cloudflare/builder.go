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
			Installation: installation.Registration(),
			Connections: []types.Connector{
				apiToken.
					Name("Cloudflare API Token").
					Description("Configure Cloudflare access using an API token scoped to your account and zones.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Cloudflare dashboard.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("Account Settings Read", "Access: Users Read", "Access: Groups Read", "Access: Organizations, Identity Providers, and Groups Read").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Description("Collect account members as directory accounts").
					Registration(),
				types.OperationRefOf[FindingsSync]().
					Ingests(runFindingsCollect).
					// TODO: remove with providerkit.UpgradeFromSection once every installation has been upgraded off main's client config
					Upgraded(providerkit.UpgradeFromSection[FindingsSync](mainFindingSyncKey)).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaFinding.Name}).
					Permissions("Account Security Center Insights Read").
					Description("Collect Cloudflare Security Center insights as findings").
					Registration(),
				types.OperationRefOf[AssetSync]().
					Ingests(runAssetCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					Permissions("Registrar Domains Read").
					Schedule(&gala.Schedule{
						MinInterval:        assetSyncMinIntervalHours * time.Hour,
						MaxInterval:        assetSyncMaxIntervalDays * assetSyncMinIntervalHours * time.Hour,
						HighDriftThreshold: gala.FullHighDriftThreshold,
					}).
					SkipDefaultLookback().
					Description("Collect Cloudflare domain registrations as assets").
					Registration(),
				DomainScanSubmitOp.Description("Submit domains to Cloudflare's URL Scanner for scanning").Registration(),
				DomainScanPollOp.Description("Poll a previously submitted Cloudflare URL Scanner result").Registration(),
				DomainScanEnrichmentOp.Description("Gather company profile, compliance, and DNS vendor data for a domain").Registration(),
				DomainScanBuildReportOp.Description("Build the onboarding domain scan report from a completed URL Scanner result and gathered enrichment").Registration(),
				domainScanRequestOp.Description("Request a domain scan for a single domain").Registration(),
				DomainScanImportOp.Description("Import a reviewer-accepted domain scan report into real records").Registration(),
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
