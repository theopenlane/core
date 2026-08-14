package hooks

import (
	"context"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/hooks/contextx"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// OrganizationCleanupListeners returns the organization cascade delete listener
func OrganizationCleanupListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaOrganization,
			Label:      "cascade_delete",
			Operations: []string{entityops.OpSoftDelete},
			Caller: func(_ *auth.Caller, payload entityops.MutationPayload) *auth.Caller {
				return newOrganizationCleanupCaller(payload.EntityID)
			},
			// the cascade hard-deletes everything the organization owns; veto emission so
			// the cascade itself produces no mutation events
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
	orgID := inv.EntityID
	cleanupCtx := inv.Context

	// queued integration jobs go first so no worker races the hard deletes below
	cancelOrganizationIntegrationJobs(inv)

	if err := organizationEdgeCleanup(cleanupCtx, orgID); err != nil {
		logx.FromContext(cleanupCtx).Error().Err(err).
			Msg("failed to cascade delete organization edges")

		return err
	}

	// this has to run before the organization row is removed
	if err := purgeOrganizationHistory(cleanupCtx, orgID); err != nil {
		logx.FromContext(cleanupCtx).Error().Err(err).
			Msg("failed to purge organization history")

		return err
	}

	// the organization row goes last, once everything it owned is gone
	if _, err := inv.Client.Organization.Delete().Where(organization.ID(orgID)).Exec(cleanupCtx); err != nil {
		logx.FromContext(cleanupCtx).Error().Err(err).Msg("failed to delete organization")

		return err
	}

	logx.FromContext(cleanupCtx).Info().Msg("organization cascade delete completed")

	return nil
}

// newOrganizationCleanupCaller returns the caller the cascade runs as. It needs to reach every
// record the organization owns regardless of who is deleting it, so it bypasses FGA checks and identifies itself as an internal operation
func newOrganizationCleanupCaller(orgID string) *auth.Caller {
	return &auth.Caller{
		OrganizationID: orgID,
		Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
	}
}
