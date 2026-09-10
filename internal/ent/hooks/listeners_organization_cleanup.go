package hooks

import (
	"context"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/hooks/contextx"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// OrganizationCleanupListeners cascades an organization soft delete by hard-deleting everything it owns
func OrganizationCleanupListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaOrganization,
			Operations: []string{entityops.OpSoftDelete},
			Caller: func(_ *auth.Caller, payload entityops.MutationPayload) *auth.Caller {
				return &auth.Caller{
					OrganizationID: payload.EntityID,
					Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
				}
			},
			ContextKeys: []func(context.Context) context.Context{
				entx.SkipSoftDelete,
				contextx.WithTupleCleanup,
				contextx.WithPurgeHistory,
				entityops.WithEmissionVetoed,
			},
			Handle: handleOrganizationCascadeDelete,
		},
	}
}

// handleOrganizationCascadeDelete removes everything an organization owns once it is deleted.
// The records are hard deleted and their history rows purged along with files stored in object storage
func handleOrganizationCascadeDelete(inv entityops.Invocation, _ entityops.MutationPayload) error {
	purgeOrganizationIntegrationJobs(inv)

	if err := organizationEdgeCleanup(inv.Context, inv.EntityID); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to cascade delete organization edges")

		return err
	}

	// this has to run before the organization row is removed
	if err := purgeOrganizationHistory(inv.Context, inv.EntityID); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to purge organization history")

		return err
	}

	// the organization row goes last, once everything it owned is gone
	if _, err := inv.Client.Organization.Delete().Where(organization.ID(inv.EntityID)).Exec(inv.Context); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to delete organization")

		return err
	}

	logx.FromContext(inv.Context).Info().Msg("organization cascade delete completed")

	return nil
}
