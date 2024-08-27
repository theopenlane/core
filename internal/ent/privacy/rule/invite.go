package rule

import (
	"context"
	"strings"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
)

const (
	inviteMemberRelation = "can_invite_members"
	inviteAdminRelation  = "can_invite_admins"
)

// CanInviteUsers is a rule that returns allow decision if user has access to invite members or admins to the organization
func CanInviteUsers() privacy.InviteMutationRuleFunc {
	return privacy.InviteMutationRuleFunc(func(ctx context.Context, m *generated.InviteMutation) error {
		oID, err := getInviteOwnerID(ctx, m)
		if err != nil || oID == "" {
			return privacy.Skipf("no owner set on request, cannot check access")
		}

		userID, err := auth.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}

		relation, err := getRelationToCheck(ctx, m)
		if err != nil {
			m.Logger.Errorw("unable to determine relation to check", "error", err)

			return err
		}

		m.Logger.Debugw("checking relationship tuples", "relation", relation, "organization_id", oID, "user_id", userID)

		ac := fgax.AccessCheck{
			SubjectID:   userID,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			ObjectID:    oID,
			Relation:    relation,
		}

		access, err := m.Authz.CheckOrgAccess(ctx, ac)
		if err != nil {
			return privacy.Skipf("unable to check access, %s", err.Error())
		}

		if access {
			m.Logger.Debugw("access allowed", "relation", relation, "organization_id", oID)

			return privacy.Allow
		}

		// proceed to next rule
		return nil
	})
}

// getInviteOwnerID returns the owner id from the mutation or the context
func getInviteOwnerID(ctx context.Context, m *generated.InviteMutation) (string, error) {
	oID, ok := m.OwnerID()
	if ok && oID != "" {
		return oID, nil
	}

	return auth.GetOrganizationIDFromContext(ctx)
}

// getRelationToCheck returns the relation to check based on the role on the mutation
func getRelationToCheck(ctx context.Context, m *generated.InviteMutation) (string, error) {
	role, ok := m.Role()
	if !ok {
		// if it is not a create operation, we need to to check the existing invite for the role
		if m.Op() != generated.OpCreate {
			id, ok := m.ID()
			if !ok {
				return "", privacy.Skipf("unable to determine invite, cannot check access")
			}

			// get the role from the existing invite
			invite, err := generated.FromContext(ctx).Invite.Get(ctx, id)
			if err != nil {
				return "", err
			}

			role = invite.Role
		}
	}

	// allow member to invite members
	if strings.EqualFold(role.String(), fgax.MemberRelation) {
		return inviteMemberRelation, nil
	}

	// default to admin
	return inviteAdminRelation, nil
}
