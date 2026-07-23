package tailscale

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
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
				{
					Ref:         tailscaleCredential.ID(),
					Name:        "Tailscale OAuth Client",
					Description: "OAuth client credentials used to read users and devices from your tailnet.",
					Schema:      tailscaleSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       tailscaleCredential.ID(),
					Name:                "Tailscale OAuth",
					Description:         "Configure Tailscale access using an OAuth client scoped to your tailnet.",
					CredentialRefs:      []types.CredentialSlotID{tailscaleCredential.ID()},
					ClientRefs:          []types.ClientID{tailscaleClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: tailscaleCredential.ID(),
						Description:   "Removes the stored OAuth credentials from Openlane. If the client is no longer needed, revoke it in the Tailscale admin console.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            tailscaleClient.ID(),
					CredentialRefs: []types.CredentialSlotID{tailscaleCredential.ID()},
					Description:    "Tailscale HTTP API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:                healthCheckOperation.Name(),
					Description:         "Verify Tailscale OAuth credentials by listing tailnet users",
					Topic:               definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:           tailscaleClient.ID(),
					Policy:              types.ExecutionPolicy{Inline: true},
					Handle:              HealthCheck{}.Handle(),
					ConfigSchema:        healthCheckSchema,
					RequiredPermissions: []string{"users:read"},
				},
				{
					Name:           directorySyncOperation.Name(),
					Description:    "Sync Tailscale users and role-based groups as directory accounts",
					Topic:          definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:      tailscaleClient.ID(),
					ConfigSchema:   directorySyncSchema,
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.DirectorySync.Disable }),
					ConfigResolver: providerkit.ConfigFrom(func(u UserInput) DirectorySync { return u.DirectorySync }),
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
				},
				{
					Name:           assetSyncOperation.Name(),
					Description:    "Sync Tailscale devices as assets",
					Topic:          definitionID.OperationTopic(assetSyncOperation.Name()),
					ClientRef:      tailscaleClient.ID(),
					ConfigSchema:   assetSyncSchema,
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.AssetSync.Disable }),
					ConfigResolver: providerkit.ConfigFrom(func(u UserInput) AssetSync { return u.AssetSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaAsset.Name,
						},
					},
					IngestHandle:        AssetSync{}.IngestHandle(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"devices:core:read", "devices:posture_attributes:read", "devices:routes:read"},
					Schedule:            gala.NewFullFetchSchedule(),
				},
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
