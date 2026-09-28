package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	inviteMemberRelation     = "can_invite_members"
	inviteAdminRelation      = "can_invite_admins"
	inviteSuperAdminRelation = "can_invite_super_admins"
	inviteAuditors           = "can_invite_auditors"
)

// CanInviteUsers is a rule that returns allow decision if user has access to invite members or admins to the organization
func CanInviteUsers() privacy.InviteMutationRuleFunc {
	return privacy.InviteMutationRuleFunc(func(ctx context.Context, m *generated.InviteMutation) error {
		oID, err := getInviteOwnerID(ctx, m)
		if err != nil || oID == "" {
			return generated.ErrPermissionDenied
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		relation, err := getRelationToCheck(ctx, m)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("unable to determine relation to check")

			return err
		}

		// owner invites are only allowed as an ownership transfer by the current owner
		role, _ := m.Role()
		transfer, _ := m.OwnershipTransfer()

		if (role == enums.RoleOwner) != transfer {
			return generated.ErrPermissionDenied
		}

		if transfer {
			relation = fgax.OwnerRelation
		}

		ac := fgax.AccessCheck{
			SubjectID:   caller.SubjectID,
			SubjectType: caller.SubjectType(),
			ObjectID:    oID,
			Relation:    relation,
		}

		logx.FromContext(ctx).Debug().Interface("tuple", ac).Msg("checking relationship tuples")

		access, err := m.Authz.CheckOrgAccess(ctx, ac)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Interface("tuple", ac).Msg("unable to check invite access")

			return generated.ErrPermissionDenied
		}

		if access {
			logx.FromContext(ctx).Debug().Str("relation", relation).Str("organization_id", oID).Msg("access allowed")

			return privacy.Allow
		}

		return generated.ErrPermissionDenied
	})
}

// getInviteOwnerID returns the owner id from the mutation or the context
func getInviteOwnerID(ctx context.Context, m *generated.InviteMutation) (string, error) {
	oID, ok := m.OwnerID()
	if ok && oID != "" {
		return oID, nil
	}

	caller, callerOk := auth.CallerFromContext(ctx)
	if !callerOk || caller == nil {
		return "", auth.ErrNoAuthUser
	}

	return caller.OrganizationID, nil
}

// getRelationToCheck returns the relation to check based on the role on the mutation
func getRelationToCheck(ctx context.Context, m *generated.InviteMutation) (string, error) {
	role, ok := m.Role()
	if ok {
		return InviteRelationForRole(role), nil
	}

	if m.Op() == generated.OpCreate {
		return InviteRelationForRole(enums.RoleMember), nil
	}

	// if it is not a create operation, we need to to check the existing invite for the role
	id, ok := m.ID()
	if !ok {
		return "", generated.ErrPermissionDenied
	}

	// get the role from the existing invite
	invite, err := m.Client().Invite.Get(ctx, id)
	if err != nil {
		return "", err
	}

	return InviteRelationForRole(invite.Role), nil
}
