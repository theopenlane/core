package hooks

import (
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// init registers the catalog listeners so gala setup picks them up automatically
func init() { registerListeners(CatalogListeners) }

// CatalogListeners refreshes adopted entities when a system-owned catalogue entity changes
func CatalogListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaEntity,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     entityops.SchemaEntity.Catalog.Fields,
			Caller:     catalogRefreshCaller,
			Handle:     handleCatalogEntityMutation,
		},
	}
}

// catalogRefreshCaller lets the refresh update adopted rows in every organization
func catalogRefreshCaller(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
	return restored.WithCapabilities(auth.CapInternalOperation | auth.CapBypassOrgFilter | auth.CapBypassFGA)
}

// handleCatalogEntityMutation refreshes the adopted copies of a system-owned entity
func handleCatalogEntityMutation(inv entityops.Invocation, _ entityops.MutationPayload) error {
	row, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Entity.Get)
	if err != nil || !ok {
		return err
	}

	if !row.SystemOwned {
		return nil
	}

	updated, err := entityops.SchemaEntity.RefreshAdopted(inv.Context, inv.Client, inv.EntityID)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Int("updated", updated).Msg("failed to refresh adopted entities from catalog")

		return err
	}

	logx.FromContext(inv.Context).Info().Int("updated", updated).Msg("refreshed adopted entities from catalog")

	return nil
}
