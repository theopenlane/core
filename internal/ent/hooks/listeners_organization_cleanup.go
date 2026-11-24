package hooks

import (
	"context"

	openfga "github.com/openfga/go-sdk"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/hooks/contextx"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// init registers the organization cleanup listeners so gala setup picks them up automatically
func init() { registerListeners(OrganizationCleanupListeners) }

// OrganizationCleanupListeners cascades an organization soft delete by hard-deleting everything it owns
func OrganizationCleanupListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			// the cascade removes the organization row other delete listeners read, so it runs last
			Priority:   listenerPriorityLast,
			Schema:     entityops.SchemaOrganization,
			Operations: []string{entityops.OpSoftDelete},
			Caller: func(_ *auth.Caller, payload entityops.MutationPayload) *auth.Caller {
				return auth.NewOrgInternalCaller(payload.EntityID)
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

	if err := purgeOrgMembersAuditing(inv.Context, inv.EntityID); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to purge organization audit_org tuples")

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

func purgeOrgMembersAuditing(ctx context.Context, orgID string) error {
	client := utils.AuthzClientFromContext(ctx)

	resp, err := client.ListObjectsRequestWithConsistency(ctx, fgax.ListRequest{
		SubjectID:   orgID,
		SubjectType: generated.TypeOrganization,
		ObjectType:  generated.TypeUser,
		Relation:    auditOrgRelation,
	}, openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.Objects) == 0 {
		return nil
	}

	deletes := make([]fgax.TupleKey, 0, len(resp.Objects))
	for _, obj := range resp.Objects {
		entity, err := fgax.ParseEntity(obj)
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("object", obj).Msg("failed to parse entity for audit_org tuple cleanup")
			continue
		}

		deletes = append(deletes, fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   orgID,
			SubjectType: generated.TypeOrganization,
			ObjectID:    entity.Identifier,
			ObjectType:  generated.TypeUser,
			Relation:    auditOrgRelation,
		}))
	}

	_, err = client.WriteTupleKeys(ctx, nil, deletes, fgax.WithIgnoreDuplicateKeyError(true))

	return err
}
