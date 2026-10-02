package tailscale

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// Builder returns the Tailscale definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "Tailscale",
			DisplayName:  "Tailscale",
			Description:  "Sync Tailscale users and role groups as directory accounts, and Tailscale devices as assets.",
			Category:     "identity",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/tailscale/overview",
			Tags:         []string{"directory", "assets"},
			Active:       true,
			Visible:      true,
			HealthCheck:  tailscaleClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				tailscaleCredential.Registration(types.CredentialRegistration{
					Name:        "Tailscale OAuth Client",
					Description: "OAuth client credentials used to read users and devices from your tailnet.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: tailscaleCredential.ID(),
					Name:          "Tailscale OAuth",
					Description:   "Configure Tailscale access using an OAuth client scoped to your tailnet.",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: tailscaleCredential.ID(),
						Description:   "Removes the stored OAuth credentials from Openlane. If the client is no longer needed, revoke it in the Tailscale admin console.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				tailscaleClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Tailscale HTTP API client",
				}),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[DirectorySync]().
					Ingests(tailscaleClient, runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("users:read", "policy_file:read").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Registration(definitionID, types.OperationRegistration{
						Description: "Sync Tailscale users and role-based groups as directory accounts",
					}),
				types.OperationRefOf[AssetSync]().
					Ingests(tailscaleClient, runAssetSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					Permissions("devices:core:read", "devices:posture_attributes:read", "devices:routes:read").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Registration(definitionID, types.OperationRegistration{
						Description: "Sync Tailscale devices as assets",
					}),
			},
			Mappings: append(providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
				types.MappingRegistration{
					Schema:  entityops.SchemaAsset.Name,
					Variant: deviceAssetVariant,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprAsset,
					},
				},
			),
		}, nil
	})
}
