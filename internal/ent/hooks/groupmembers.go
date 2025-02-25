package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
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
			orgMemberID, err := getOrgMemberID(ctx, m, userID, group.OwnerID)
			if err != nil {
				return nil, err
			}

			m.SetOrgmembershipID(orgMemberID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
