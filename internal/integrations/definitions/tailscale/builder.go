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
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Tailscale",
				DisplayName: "Tailscale",
				Description: "Sync Tailscale users and role groups as directory accounts, and Tailscale devices as assets.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/tailscale/overview",
				Tags:        []string{"directory", "assets"},
				Active:      true,
				Visible:     true,
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  tailscaleClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				tailscaleCredential.Registration(types.CredentialRegistration{
					Name:        "Tailscale OAuth Client",
					Description: "OAuth client credentials used to read users and devices from your tailnet.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				tailscaleConnection.Registration(types.ConnectionRegistration{
					Name:        "Tailscale OAuth",
					Description: "Configure Tailscale access using an OAuth client scoped to your tailnet.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored OAuth credentials from Openlane. If the client is no longer needed, revoke it in the Tailscale admin console.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				tailscaleClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Tailscale HTTP API client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description:         "Sync Tailscale users and role-based groups as directory accounts",
					Policy:              types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Ingest:              providerkit.DirectoryIngestContracts(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"users:read", "policy_file:read"},
					Schedule:            gala.NewFullFetchSchedule(),
				}),
				assetSyncOperation.Registration(definitionID, types.OperationRegistration{
					Description: "Sync Tailscale devices as assets",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaAsset.Name,
						},
					},
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"devices:core:read", "devices:posture_attributes:read", "devices:routes:read"},
					Schedule:            gala.NewFullFetchSchedule(),
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
