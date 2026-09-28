package cloudflare

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Cloudflare definition builder with the supplied runtime config applied.
// runtime.DomainScan configures vendor/technology classification for onboarding domain scan reports.
// When devMode is true or runtime.Provisioned() is true, a RuntimeIntegration is included so
// system-initiated calls (e.g. onboarding domain scans) can use the operator-owned account
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
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				cloudflareCredential.Registration(types.CredentialRegistration{
					Name:        "Cloudflare API Credential",
					Description: "API token used to access Cloudflare account and zone data.",
					Schema:      cloudflareCredential.Schema(),
				}),
			},
			Connections: []types.ConnectionRegistration{
				cloudflareConnection.Registration(types.ConnectionRegistration{
					Name:           "Cloudflare API Token",
					Description:    "Configure Cloudflare access using an API token scoped to your account and zones.",
					CredentialRefs: []types.CredentialSlotID{cloudflareCredential.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: cloudflareClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: cloudflareCredential.ID(),
						Description:   "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Cloudflare dashboard.",
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
					Description:    "Collect account members as directory accounts",
					Policy:         types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.DirectorySync.Disable }),
					ConfigResolver: directorySyncOperation.ConfigFrom(func(u UserInput) DirectorySync { return u.DirectorySync }),
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
					IngestHandle:        DirectorySync{}.IngestHandle(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"Account Settings Read", "Access: Users Read", "Access: Groups Read", "Access: Organizations, Identity Providers, and Groups Read"},
					Schedule:            gala.NewFullFetchSchedule(),
				}),
				findingsSyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description:    "Collect Cloudflare Security Center insights as findings",
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.FindingsSync.Disable }),
					ConfigResolver: findingsSyncOperation.ConfigFrom(func(u UserInput) FindingsSync { return u.FindingsSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaFinding.Name,
						},
					},
					IngestHandle:        FindingsCollect{}.IngestHandle(),
					RequiredPermissions: []string{"Account Security Center Insights Read"},
				}),
				assetSyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description:    "Collect Cloudflare domain registrations as assets",
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.AssetSync.Disable }),
					ConfigResolver: assetSyncOperation.ConfigFrom(func(u UserInput) AssetSync { return u.AssetSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaAsset.Name,
						},
					},
					IngestHandle:        AssetCollect{}.IngestHandle(),
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
					Handle:             DomainScanSubmit{}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanPollOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Poll a previously submitted Cloudflare URL Scanner result",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					Handle:             DomainScanPoll{}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanEnrichmentOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Gather company profile, compliance, and DNS vendor data for a domain",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					Handle:             DomainScanGatherEnrichment{}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanBuildReportOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Build the onboarding domain scan report from a completed URL Scanner result and gathered enrichment",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					Handle:             DomainScanBuildReport{}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
				DomainScanRequestOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Request a domain scan for a single domain",
					Policy:      types.ExecutionPolicy{Inline: true, SkipRunRecord: true},
					// Disable if the runtime is not provisioned
					DisabledForAll: !runtime.Provisioned(),
					// only applied to user created scans, not onboarding scans
					RateLimit:             &types.RateLimitPolicy{Window: time.Hour},
					Handle:                DomainScanRequest{}.Handle(),
					CustomerSelectable:    lo.ToPtr(false),
					RequiresPaymentMethod: true,
				}),
				DomainScanImportOp.Registration(DefinitionID, types.OperationRegistration{
					Description:        "Import a reviewer-accepted domain scan report into real records",
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					Handle:             DomainScanImport{}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				}),
			},
			GalaListeners: []types.GalaListenerRegistration{
				domainScanListeners(),
			},
			Mappings: []types.MappingRegistration{
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
				{
					Schema: entityops.SchemaFinding.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprFinding,
						Links: []types.LinkRule{
							{
								TargetSchema: entityops.SchemaControl.Name,
								TargetField:  control.FieldRefCode,
								SourceField:  entityops.FindingFields.Category.InputKey,
								SourceList:   entityops.FindingFields.Categories.InputKey,
							},
						},
					},
				},
				{
					Schema: entityops.SchemaAsset.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprAsset,
					},
				},
			},
		}

		if runtime.Provisioned() {
			runtimeCloudflareRef.SetConfig(runtime)

			marshaledConfig, err := runtimeCloudflareRef.MarshalConfig()
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
			}

			def.RuntimeIntegration = lo.ToPtr(runtimeCloudflareRef.Registration(types.RuntimeIntegrationRegistration{
				Config: marshaledConfig,
				Build:  runtimeCloudflareClientBuilder(),
			}))
		}

		return def, nil
	})
}
