package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
)

func HookProgramMembers() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ProgramMembershipFunc(func(ctx context.Context, m *generated.ProgramMembershipMutation) (generated.Value, error) {
			// if userID is on the mutation then we need to check if the user is a member of the organization
			userID, ok := m.UserID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			programID, ok := m.ProgramID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			program, err := m.Client().Program.Get(ctx, programID)
			if err != nil {
				// program not found, let the default validation handle it
				return next.Mutate(ctx, m)
			}

			// ensure user is a member of the organization
			exists, err := m.Client().OrgMembership.Query().
				Where(orgmembership.UserID(userID)).
				Where(orgmembership.OrganizationID(program.OwnerID)).
				Exist(ctx)
			if err != nil {
				return nil, err
			}

			if !exists {
				return nil, ErrUserNotInOrg
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
