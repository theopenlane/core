package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func HookGroupMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupMembershipFunc(func(ctx context.Context, m *generated.GroupMembershipMutation) (generated.Value, error) {
			// check role, if its not set the default is member
			userID, ok := m.UserID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			groupID, ok := m.GroupID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			group, err := m.Client().Group.Get(ctx, groupID)
			if err != nil {
				// group not found, let the default validation handle it
				return next.Mutate(ctx, m)
			}

			_, allowCtx := contextx.From[ManagedContextKey](ctx)
			if group.IsManaged && !allowCtx {
				return nil, ErrManagedGroup
			}

			// ensure user is a member of the organization
			orgMemberID, err := m.Client().OrgMembership.Query().
				Where(orgmembership.UserID(userID)).
				Where(orgmembership.OrganizationID(group.OwnerID)).
				OnlyID(ctx)
			if err != nil || orgMemberID == "" {
				log.Error().Err(err).Msg("failed to get org membership, cannot add user to group")

				return nil, ErrUserNotInOrg
			}

			m.SetOrgmembershipID(orgMemberID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// HookGroupDeleteMembers is a hook that runs on group membership deletions
// to ensure that a user cannot update or remove themselves from a group
func HookGroupDeleteMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupMembershipFunc(func(ctx context.Context, m *generated.GroupMembershipMutation) (generated.Value, error) {
			// bypass privacy check if the context allows it
			if _, allow := privacy.DecisionFromContext(ctx); allow {
				return next.Mutate(ctx, m)
			}

			gmID, ok := m.ID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			// check if group member is the authenticated user
			userID, err := auth.GetUserIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			groupMember, err := m.Client().GroupMembership.Query().Where(groupmembership.ID(gmID)).Only(ctx)
			if err != nil {
				return nil, err
			}

			if groupMember.UserID == userID {
				log.Debug().Msg("user cannot update or delete themselves in a group")

				return nil, generated.ErrPermissionDenied
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne|ent.OpUpdate|ent.OpDelete)
}
