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
			Installation: installation.Registration(),
			Connections: []types.Connector{
				tailscaleConnection.
					Name("Tailscale OAuth").
					Description("Configure Tailscale access using an OAuth client scoped to your tailnet.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored OAuth credentials from Openlane. If the client is no longer needed, revoke it in the Tailscale admin console.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[providerkit.DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("users:read", "policy_file:read").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Description("Sync Tailscale users and role-based groups as directory accounts").
					Registration(),
				types.OperationRefOf[AssetSync]().
					Ingests(runAssetSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					Permissions("devices:core:read", "devices:posture_attributes:read", "devices:routes:read").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Description("Sync Tailscale devices as assets").
					Registration(),
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
