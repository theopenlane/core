package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// HookGroupMembers checks the users role, ensures they are a member of the org, and prevents direct modifications to managed groups unless the caller has the bypass capability
func HookGroupMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupMembershipFunc(func(ctx context.Context, m *generated.GroupMembershipMutation) (generated.Value, error) {
			// skip if this is an internal operation (e.g. org creation)
			if caller, ok := auth.CallerFromContext(ctx); ok && caller.Has(auth.CapInternalOperation) {
				return next.Mutate(ctx, m)
			}

			// skip if this is an explicit managed group bypass
			if caller, ok := auth.CallerFromContext(ctx); ok && caller.Has(auth.CapBypassManagedGroup) {
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

			// allow general allow context or managed group bypass to modify managed groups
			_, allowCtx := privacy.DecisionFromContext(ctx)
			caller, _ := auth.CallerFromContext(ctx)
			allowManagedCtx := caller != nil && caller.Has(auth.CapBypassManagedGroup)

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
