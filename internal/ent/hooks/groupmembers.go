package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
)

// HookGroupMembers checks the users role, ensures they are a member of the org, and prevents direct modifications to managed groups unless the caller has the bypass capability
func HookGroupMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupMembershipFunc(func(ctx context.Context, m *generated.GroupMembershipMutation) (generated.Value, error) {
			// skip when the caller has both caps, e.g. org creation adding the creator to managed groups before they are an org member
			if auth.HasInContextCaller(ctx, auth.CapInternalOperation|auth.CapBypassFGA) {
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

			// allow query, the permissions are not yet added to the group
			// and members get added during the create process
			group, err := m.Client().Group.Get(auth.WithInternalOperationContext(ctx), groupID)
			if err != nil {
				return nil, err
			}

			// if the group is managed, but isn't a managed group context
			// return error
			if group.IsManaged && !auth.HasInContextCaller(ctx, auth.CapBypassManagedGroup) {
				return nil, ErrManagedGroup
			}

			// the edges already ensure the user is a member of the organization now.
			// but we still need to fetch the org membership id to link correctly
			orgMemberID, err := getOrgMemberID(ctx, m, userID, group.OwnerID)
			if err != nil {
				return nil, err
			}

			m.SetOrgMembershipID(orgMemberID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
