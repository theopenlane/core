package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks/contextx"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaOrganizationCleanupListeners registers the organization cascade delete on Gala.
// This is deliberately independent of the entitlement listeners, the cascade has to run whether or
// not billing is configured, otherwise the records an organization owns are left behind entirely
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
// The records are hard deleted and their history rows purged, leaving them soft deleted would keep
// the rows, the uploaded objects and the full field values in the history tables indefinitely
func handleOrganizationCascadeDelete(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	handlerCtx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	orgID, ok := eventqueue.MutationEntityID(payload, handlerCtx.Envelope.Headers.Properties)
	if !ok || orgID == "" {
		return nil
	}

	cleanupCtx := entgen.NewContext(organizationCleanupContext(handlerCtx.Context), client)

	if err := entgen.OrganizationEdgeCleanup(cleanupCtx, orgID); err != nil {
		logx.FromContext(cleanupCtx).Error().Err(err).Str("organization_id", orgID).
			Msg("failed to cascade delete organization edges")

		return err
	}

	// this has to run before the organization row is removed
	if err := entgen.PurgeOrganizationHistory(cleanupCtx, organization.ID(orgID)); err != nil {
		logx.FromContext(cleanupCtx).Error().Err(err).Str("organization_id", orgID).
			Msg("failed to purge organization history")

		return err
	}

	// the organization row goes last, once everything it owned is gone
	if _, err := client.Organization.Delete().Where(organization.ID(orgID)).Exec(cleanupCtx); err != nil {
		logx.FromContext(cleanupCtx).Error().Err(err).Str("organization_id", orgID).
			Msg("failed to delete organization")

		return err
	}

	logx.FromContext(cleanupCtx).Info().Str("organization_id", orgID).Msg("organization cascade delete completed")

	return nil
}

// organizationCleanupContext builds the context the cascade runs under, it bypasses privacy rules,
// turns the cascaded deletes into hard deletes and opts the cascade into purging history rows
func organizationCleanupContext(ctx context.Context) context.Context {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, auth.NewWebhookCaller(""))

	// without this the cascaded deletes are rewritten into soft deletes by the mixin, which also
	// means the file hook never fires and the uploaded objects are orphaned in object storage
	allowCtx = entx.SkipSoftDelete(allowCtx)

	return contextx.WithPurgeHistory(allowCtx)
}
