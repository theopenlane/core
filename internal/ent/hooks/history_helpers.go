//go:build !codegen

// this file calls the generated edge and history cleanup functions, which are excluded from
// compilation during code generation because they reference history packages that may not be
// generated yet, so this file must be excluded during codegen
package hooks

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/program"
	"github.com/theopenlane/core/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/entfga"
)

// purgeOrganizationHistory is a wrapper around generated PurgeOrganizationHistory to prevent issues during codegen
func purgeOrganizationHistory(ctx context.Context, orgID string) error {
	return generated.PurgeOrganizationHistory(ctx, organization.ID(orgID))
}

// organizationEdgeCleanup is a wrapper around generated OrganizationEdgeCleanup to prevent issues during codegen
func organizationEdgeCleanup(ctx context.Context, orgID string) error {
	return generated.OrganizationEdgeCleanup(ctx, orgID)
}

// removeUserOrgScopedMemberships removes the user's group and program memberships that belong to
// the organization they are being removed from; memberships in other organizations are left intact
func removeUserOrgScopedMemberships(ctx context.Context, m *generated.OrgMembershipMutation, userID, orgID string) error {
	// delete the fga tuples for the memberships before the records are removed
	ctx = entfga.WithDeleteTuplesFirst(ctx)

	groupPreds := []predicate.GroupMembership{
		groupmembership.UserID(userID),
		groupmembership.HasGroupWith(group.OwnerID(orgID)),
	}

	// the history rows are matched by a sub-select on the records being removed, so this has to run first
	if err := generated.PurgeGroupMembershipHistory(ctx, groupPreds...); err != nil {
		return err
	}

	if _, err := m.Client().GroupMembership.Delete().Where(groupPreds...).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error deleting user's group memberships in organization")

		return err
	}

	programPreds := []predicate.ProgramMembership{
		programmembership.UserID(userID),
		programmembership.HasProgramWith(program.OwnerID(orgID)),
	}

	if err := generated.PurgeProgramMembershipHistory(ctx, programPreds...); err != nil {
		return err
	}

	if _, err := m.Client().ProgramMembership.Delete().Where(programPreds...).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error deleting user's program memberships in organization")

		return err
	}

	return nil
}
