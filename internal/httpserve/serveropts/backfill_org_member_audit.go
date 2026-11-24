package serveropts

import (
	"context"

	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
)

const userAuditOrgRelation = "audit_org"

func backfillOrgMembersForAuditing(ctx context.Context, deps backfillDeps) error {
	var lastKnownID string

	for {
		rows, err := deps.Client.OrgMembership.Query().
			Where(
				orgmembership.IDGT(lastKnownID),
				orgmembership.HasOrganizationWith(organization.PersonalOrg(false)),
			).
			Order(orgmembership.ByID()).
			Select(
				orgmembership.FieldID,
				orgmembership.FieldOrganizationID,
				orgmembership.FieldUserID,
			).
			Limit(backfillBatchSize).
			All(ctx)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			logx.FromContext(ctx).Info().Msg("backfill: org member auditing tuples complete")

			return nil
		}

		writes := make([]fgax.TupleKey, 0, len(rows))
		for _, row := range rows {
			if row.OrganizationID == "" || row.UserID == "" {
				continue
			}

			writes = append(writes, fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   row.OrganizationID,
				SubjectType: generated.TypeOrganization,
				ObjectID:    row.UserID,
				ObjectType:  generated.TypeUser,
				Relation:    userAuditOrgRelation,
			}))
		}

		if _, err := deps.Client.Authz.WriteTupleKeys(ctx, writes, nil, fgax.WithIgnoreDuplicateKeyError(true)); err != nil {
			return err
		}

		lastKnownID = rows[len(rows)-1].ID
	}
}
