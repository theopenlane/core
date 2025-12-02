package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/privacy"
)

func HookGroupMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupMembershipFunc(func(ctx context.Context, m *generated.GroupMembershipMutation) (generated.Value, error) {
			// skip if we are creating an organization
			if _, ok := contextx.From[auth.OrganizationCreationContextKey](ctx); ok {
				return next.Mutate(ctx, m)
			}

			if _, ok := contextx.From[auth.ManagedGroupContextKey](ctx); ok {
				return next.Mutate(ctx, m)
			}

			// check role, if its not set the default is member
			userID, ok := m.UserID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			groupID, ok := m.GroupID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			allowQueryCtx := privacy.DecisionContext(ctx, privacy.Allow)

			group, err := m.Client().Group.Get(allowQueryCtx, groupID)
			if err != nil {
				return nil, err
			}

			// allow general allow context to bypass managed group check
			_, allowCtx := privacy.DecisionFromContext(ctx)
			_, allowManagedCtx := contextx.From[ManagedContextKey](ctx)

			if group.IsManaged && (!allowManagedCtx && !allowCtx) {
				return nil, ErrManagedGroup
			}

			// ensure user is a member of the organization
			orgMemberID, err := getOrgMemberID(ctx, m, userID, group.OwnerID)
			if err != nil {
				return nil, err
			}

			m.SetOrgMembershipID(orgMemberID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
