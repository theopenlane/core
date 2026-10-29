package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/report"
	"github.com/theopenlane/core/v2/internal/ent/generated/reportmembership"
)

// HookReportMembers makes sure the visibility rules are respected. members can be added to
// reports except it is a private one which only the owner can be a member of
func HookReportMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ReportMembershipFunc(func(ctx context.Context, m *generated.ReportMembershipMutation) (ent.Value, error) {
			role, ok := m.Role()

			// only owner and member roles allowed here
			if ok && role != enums.RoleOwner && role != enums.RoleMember {
				return nil, ErrUnsupportedFGARole
			}

			if auth.IsInternalRequest(ctx) {
				return next.Mutate(ctx, m)
			}

			if m.Op().Is(ent.OpCreate) {
				if err := createReportMember(ctx, m); err != nil {
					return nil, err
				}

				return next.Mutate(ctx, m)
			}

			ids, err := getMutationIDs(ctx, m)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get report membership IDs for update")

				return nil, err
			}

			// if at least one report in a bulk update has a private visibility, we still
			// want to block it getting members
			exists, err := m.Client().ReportMembership.Query().
				Where(
					reportmembership.IDIn(ids...),
					reportmembership.HasReportWith(
						report.VisibilityEQ(enums.ReportVisibilityPrivate),
					),
				).
				Exist(ctx)
			if err != nil {
				return nil, err
			}

			if exists {
				return nil, ErrReportIsPrivate
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

func createReportMember(ctx context.Context, m *generated.ReportMembershipMutation) error {
	userID, ok := m.UserID()
	if !ok {
		return nil
	}

	id, ok := m.ReportID()
	if !ok {
		return nil
	}

	report, err := m.Client().Report.Query().
		Where(report.ID(id)).
		Select(report.FieldVisibility, report.FieldOwnerID).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("report_id", id).Msg("failed to get report for membership creation")

		return err
	}

	// cannot add members to a private report
	if report.Visibility == enums.ReportVisibilityPrivate {
		return ErrReportIsPrivate
	}

	memberID, err := m.Client().OrgMembership.Query().
		Where(orgmembership.UserID(userID)).
		Where(orgmembership.OrganizationID(report.OwnerID)).
		OnlyID(ctx)
	if err != nil || memberID == "" {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get org membership, cannot add user to report")

		return ErrUserNotInOrg
	}

	m.SetOrgMembershipID(memberID)
	return nil
}
