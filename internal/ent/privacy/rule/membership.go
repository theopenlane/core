package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// AllowSelfOrgMembershipDelete allows users to delete only their own org membership
func AllowSelfOrgMembershipDelete() privacy.OrgMembershipMutationRuleFunc {
	return privacy.OrgMembershipMutationRuleFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) error {
		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil || caller.SubjectID == "" {
			return privacy.Skipf("unable to get user ID from context")
		}

		userID := caller.SubjectID

		id, ok := m.ID()
		if !ok {
			return privacy.Skip
		}

		orgMembership, err := m.Client().OrgMembership.Get(ctx, id)
		if err != nil {
			return privacy.Skipf("unable to get org membership: %v", err)
		}

		if orgMembership.UserID == userID {
			return privacy.Allow
		}

		return privacy.Skipf("user can only delete their own membership")
	})
}

// AllowOrgMemberRoleUpdate allows role updates using the same ceiling as invites.
func AllowOrgMemberRoleUpdate() privacy.OrgMembershipMutationRuleFunc {
	return privacy.OrgMembershipMutationRuleFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) error {
		newRole, ok := m.Role()
		if !ok {
			return privacy.Skip
		}

		var ids []string
		var err error

		switch {
		case m.Op().Is(ent.OpUpdateOne):

			if id, ok := m.ID(); ok {
				ids = []string{id}
			}

		case m.Op().Is(ent.OpUpdate):

			ids, err = m.IDs(ctx)
			if err != nil {
				return privacy.Skipf("unable to get org membership ids: %v", err)
			}
		}

		if len(ids) == 0 {
			ids, err = m.IDs(ctx)
			if err != nil {
				return privacy.Skipf("unable to get org membership ids: %v", err)
			}
		}

		members, err := m.Client().OrgMembership.Query().
			Where(orgmembership.IDIn(ids...)).
			Select(orgmembership.FieldOrganizationID, orgmembership.FieldRole).
			All(ctx)
		if err != nil {
			return privacy.Skipf("unable to get org membership: %v", err)
		}

		if len(members) == 0 {
			return privacy.Allow
		}

		if newRole == enums.RoleOwner {
			return privacy.Skip
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		for _, member := range members {
			if member.Role == enums.RoleOwner {
				return privacy.Skip
			}

			check := fgax.AccessCheck{
				SubjectID:   caller.SubjectID,
				SubjectType: caller.SubjectType(),
				ObjectID:    member.OrganizationID,
				Relation:    InviteRelationForRole(member.Role),
				Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
			}

			access, err := m.Authz.CheckOrgAccess(ctx, check)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Interface("tuple", check).Msg("unable to check role assignment access")
				return privacy.Skipf("unable to check access: %v", err)
			}

			if !access {
				return generated.ErrPermissionDenied
			}

			newRoleAccess := check
			newRoleAccess.Relation = InviteRelationForRole(newRole)

			access, err = m.Authz.CheckOrgAccess(ctx, newRoleAccess)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Interface("tuple", newRoleAccess).Msg("unable to check role assignment access")
				return privacy.Skipf("unable to check access: %v", err)
			}

			if !access {
				return generated.ErrPermissionDenied
			}
		}

		return privacy.Allow
	})
}
