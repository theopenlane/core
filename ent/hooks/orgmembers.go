package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/group"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/organization"
	"github.com/theopenlane/ent/generated/privacy"
)

func HookOrgMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgMembershipFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) (generated.Value, error) {
			userID, _ := m.UserID()

			orgID, exists := m.OrganizationID()
			if !exists || orgID == "" {
				var err error
				// get the organization based on authorized context if its not set
				orgID, err = auth.GetOrganizationIDFromContext(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get organization id from context: %w", err)
				}

				// set organization id in mutation
				m.SetOrganizationID(orgID)
			}

			orgMember := OrgMember{
				UserID: userID,
				OrgID:  orgID,
			}

			// check role, if its not set the default is member
			role, _ := m.Role()
			if role == enums.RoleOwner {
				val, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				ctxWithAuth := ctx
				if _, err := auth.GetAuthenticatedUserFromContext(ctx); err != nil {
					// set up authenticated user context for internal calls if not already present
					// this is needed for clis and other test contexts that may not have proper auth context
					ctxWithAuth = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
						SubjectID:       userID,
						OrganizationID:  orgID,
						OrganizationIDs: []string{orgID},
					})
				}

				return val, createUserManagedGroup(ctxWithAuth, m, orgMember)
			}

			// get the organization
			org, err := m.Client().Organization.Query().WithSetting().Where(organization.ID(orgID)).Only(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get organization")

				return nil, err
			}

			// do not allow members to be added to personal orgs
			if org.PersonalOrg {
				return nil, ErrPersonalOrgsNoMembers
			}

			// allow the request, which is for a user other than the authenticated user
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			// ensure user email can be added to the org
			user, err := m.Client().User.Get(allowCtx, userID)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get user")

				if generated.IsNotFound(err) {
					// use a different error message for user not found
					// so our error parsing can differentiate between the two
					return nil, ErrUserNotFound
				}

				return nil, err
			}

			if err := checkAllowedEmailDomain(user.Email, org.Edges.Setting); err != nil {
				return nil, err
			}

			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// update the managed group members when members are added
			// after the mutation has been executed
			if err := updateManagedGroupMembers(ctx, m); err != nil {
				return nil, err
			}

			if err := createUserManagedGroup(ctx, m, orgMember); err != nil {
				return nil, err
			}

			// check to see if the default org needs to be updated for the user
			if err := updateOrgMemberDefaultOrgOnCreate(ctx, m, orgID); err != nil {
				return nil, err
			}

			return retValue, err
		})
	}, ent.OpCreate)
}

// HookUpdateManagedGroups runs when org members are added to add the users to the system managed groups
func HookUpdateManagedGroups() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgMembershipFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				// update the managed group members when members are added
				// before the mutation has been executed
				if err := updateManagedGroupMembers(ctx, m); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne) // handle soft deletes as well as hard deletes
}

// HookOrgMembersDelete is a hook that runs during the delete operation of an org membership
func HookOrgMembersDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgMembershipFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) (generated.Value, error) {
			// we only want to do this on direct deleteOrgMembership operations
			// deleteOrganization will be handled by the organization hook
			rootFieldCtx := graphql.GetRootFieldContext(ctx)
			if rootFieldCtx == nil || rootFieldCtx.Object != "deleteOrgMembership" {
				logx.FromContext(ctx).Info().Msg("skipping org membership delete hook")

				return next.Mutate(ctx, m)
			}

			// get the existing org membership
			id, ok := m.ID()
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrInvalidInput, "id is required")
			}

			// get the org membership
			orgMembership, err := m.Client().OrgMembership.Get(ctx, id)
			if err != nil {
				return nil, err
			}

			if orgMembership.Role == enums.RoleOwner {
				return nil, ErrOrgOwnerCannotBeDeleted
			}

			// execute the delete operation
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// check to see if the default org needs to be updated for the user
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
			if _, err = checkAndUpdateDefaultOrg(allowCtx, orgMembership.UserID, orgMembership.OrganizationID, m.Client()); err != nil {
				return nil, err
			}

			if err := deleteSystemManagedUserGroup(allowCtx, m, orgMembership.UserID, orgMembership.OrganizationID); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error deleting user's system managed group from organization")
				return nil, err
			}

			if m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
				req := fgax.TupleRequest{
					SubjectID:   orgMembership.UserID,
					SubjectType: generated.TypeUser,
					ObjectID:    orgMembership.OrganizationID,
					ObjectType:  generated.TypeOrganization,
					Relation:    orgMembership.Role.String(),
				}

				tuple := fgax.GetTupleKey(req)

				if _, err := m.Client().Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{tuple}); err != nil {
					logx.FromContext(ctx).Error().Err(err).Interface("delete_tuple", tuple).Msg("failed to delete relationship tuple")
					return nil, err
				}
			}

			return retValue, err
		})
	}, ent.OpDeleteOne|ent.OpDelete|ent.OpUpdate|ent.OpUpdateOne) // handle soft deletes as well as hard deletes
}

// updateOrgMemberDefaultOrgOnCreate updates the user's default org if the user has no default org or
// the default org is their personal org
func updateOrgMemberDefaultOrgOnCreate(ctx context.Context, m *generated.OrgMembershipMutation, orgID string) error {
	// get the user id from the mutation, this is a required field
	userID, ok := m.UserID()
	if !ok {
		// this should never happen because the mutation should have already failed
		return fmt.Errorf("%w: %s", ErrInvalidInput, "user id is required")
	}

	// allow the request, which is for a user other than the authenticated user
	// to update the default org
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	return updateDefaultOrgIfPersonal(allowCtx, userID, orgID, m.Client())
}

func deleteSystemManagedUserGroup(ctx context.Context,
	m *generated.OrgMembershipMutation, userID, orgID string) error {
	user, err := m.Client().User.Get(ctx, userID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error getting user for managed group deletion")
		return err
	}

	_, err = m.Client().Group.Delete().
		Where(
			group.CreatedBy(userID),
			group.IsManaged(true),
			group.OwnerID(orgID),
			group.Name(getUserGroupName(user.DisplayName, user.ID)),
		).Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error deleting user's system managed group")
		return err
	}

	return nil
}

func getUserGroupName(displayName, id string) string {
	return fmt.Sprintf("%s - %s", displayName, id)
}

// createUserManagedGroup creates a personal managed group for the user accepting the invite
// this mirrors the behavior in organization creation where users get their own managed group
func createUserManagedGroup(ctx context.Context, m *generated.OrgMembershipMutation, member OrgMember) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	dbUser, err := m.Client().User.Get(allowCtx, member.UserID)
	if err != nil {
		logx.FromContext(allowCtx).Error().Err(err).Msg("error fetching user from the database")
		return err
	}

	org, err := m.Client().Organization.Get(allowCtx, member.OrgID)
	if err != nil {
		return err
	}

	if org.PersonalOrg {
		return nil
	}

	tags := []string{"managed"}

	desc := fmt.Sprintf("Group for %s", dbUser.DisplayName)

	groupInput := generated.CreateGroupInput{
		Name:        getUserGroupName(dbUser.DisplayName, dbUser.ID),
		DisplayName: &dbUser.DisplayName,
		Description: &desc,
		Tags:        tags,
	}

	group, err := m.Client().Group.Create().
		SetInput(groupInput).
		SetIsManaged(true).
		SetOwnerID(member.OrgID).
		Save(allowCtx)
	if err != nil {
		logx.FromContext(allowCtx).Error().Err(err).Msg("error creating user managed group")
		return err
	}

	input := generated.CreateGroupMembershipInput{
		Role:    &enums.RoleMember,
		UserID:  member.UserID,
		GroupID: group.ID,
	}

	if err := m.Client().GroupMembership.Create().SetInput(input).Exec(allowCtx); err != nil {
		logx.FromContext(allowCtx).Error().Err(err).Msg("error adding user to their managed group")
		return err
	}

	return nil
}
