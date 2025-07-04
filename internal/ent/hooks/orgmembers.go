package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
)

func HookOrgMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgMembershipFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) (generated.Value, error) {
			// check role, if its not set the default is member
			role, _ := m.Role()
			if role == enums.RoleOwner {
				return next.Mutate(ctx, m)
			}

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

			// get the organization
			org, err := m.Client().Organization.Query().WithSetting().Where(organization.ID(orgID)).Only(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("failed to get organization")

				return nil, err
			}

			// do not allow members to be added to personal orgs
			if org.PersonalOrg {
				return nil, ErrPersonalOrgsNoMembers
			}

			// ensure user email can be added to the org
			userID, _ := m.UserID()

			// allow the request, which is for a user other than the authenticated user
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			user, err := m.Client().User.Get(allowCtx, userID)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("failed to get user")

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
				zerolog.Ctx(ctx).Warn().Msg("skipping org membership delete hook")

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
					zerolog.Ctx(ctx).Error().Err(err).Interface("delete_tuple", tuple).Msg("failed to delete relationship tuple")
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
