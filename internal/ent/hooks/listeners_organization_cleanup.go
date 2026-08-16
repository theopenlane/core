package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/hooks/contextx"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaOrganizationCleanupListeners registers the organization cascade delete on Gala
func RegisterGalaOrganizationCleanupListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeOrganization),
			Name:  "organization.cascade_delete",
			Operations: []string{
				ent.OpDelete.String(),
				ent.OpDeleteOne.String(),
				eventqueue.SoftDeleteOne,
			},
			Handle: handleOrganizationCascadeDelete,
		},
	)
}

// handleOrganizationCascadeDelete removes everything an organization owns once it is deleted.
// The records are hard deleted and their history rows purged along with files stored in object storage
func handleOrganizationCascadeDelete(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	handlerCtx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	orgID, ok := eventqueue.MutationEntityID(payload, handlerCtx.Envelope.Headers.Properties)
	if !ok || orgID == "" {
		return nil
	}

	cleanupCtx := entgen.NewContext(organizationCleanupContext(handlerCtx.Context, orgID), client)

	cleanupCtx = logx.WithFields(cleanupCtx, logx.LogFields{
		"organization_id": orgID,
	})

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
	if _, err := client.Organization.Delete().Where(organization.ID(orgID)).Exec(cleanupCtx); err != nil {
		logx.FromContext(cleanupCtx).Error().Err(err).
			Msg("failed to delete organization")

		return err
	}

	logx.FromContext(cleanupCtx).Info().Msg("organization cascade delete completed")

	return nil
}

// organizationCleanupContext builds the context the cascade runs under, it bypasses privacy rules,
// turns the cascaded deletes into hard deletes and opts the cascade into purging history rows
func organizationCleanupContext(ctx context.Context, orgID string) context.Context {
	allowCtx := auth.WithCaller(ctx, newOrganizationCleanupCaller(orgID))

	allowCtx = entx.SkipSoftDelete(allowCtx)

	// explicitly cleanup tuples for every record
	allowCtx = contextx.WithTupleCleanup(allowCtx)

	return contextx.WithPurgeHistory(allowCtx)
}

// newOrganizationCleanupCaller returns the caller the cascade runs as. It needs to reach every
// record the organization owns regardless of who is deleting it, so it bypasses FGA checks and identifies itself as an internal operation
func newOrganizationCleanupCaller(orgID string) *auth.Caller {
	return &auth.Caller{
		OrganizationID: orgID,
		Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
	}
}
