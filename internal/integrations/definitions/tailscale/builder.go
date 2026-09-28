package tailscale

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				tailscaleCredential.Registration(types.CredentialRegistration{
					Name:        "Tailscale OAuth Client",
					Description: "OAuth client credentials used to read users and devices from your tailnet.",
					Schema:      tailscaleCredential.Schema(),
				}),
			},
			Connections: []types.ConnectionRegistration{
				tailscaleConnection.Registration(types.ConnectionRegistration{
					Name:           "Tailscale OAuth",
					Description:    "Configure Tailscale access using an OAuth client scoped to your tailnet.",
					CredentialRefs: []types.CredentialSlotID{tailscaleCredential.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: tailscaleClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: tailscaleCredential.ID(),
						Description:   "Removes the stored OAuth credentials from Openlane. If the client is no longer needed, revoke it in the Tailscale admin console.",
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
					Description:    "Sync Tailscale users and role-based groups as directory accounts",
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
					RequiredPermissions: []string{"users:read", "policy_file:read"},
					Schedule:            gala.NewFullFetchSchedule(),
				}),
				assetSyncOperation.Registration(definitionID, types.OperationRegistration{
					Description:    "Sync Tailscale devices as assets",
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.AssetSync.Disable }),
					ConfigResolver: assetSyncOperation.ConfigFrom(func(u UserInput) AssetSync { return u.AssetSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaAsset.Name,
						},
					},
					IngestHandle:        AssetSync{}.IngestHandle(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"devices:core:read", "devices:posture_attributes:read", "devices:routes:read"},
					Schedule:            gala.NewFullFetchSchedule(),
				}),
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
					Schema:  entityops.SchemaAsset.Name,
					Variant: deviceAssetVariant,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprAsset,
					},
				},
			},
		}, nil
	})
}
