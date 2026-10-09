package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// HookProgramMembers is a hook that ensures that the user is a member of the organization
// before allowing them to be added to a program
// TODO (sfunk): can this be generic across all edges with users that are owned by an organization?
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

			internalCtx := auth.WithInternalReadContext(ctx)

			program, err := m.Client().Program.Get(internalCtx, programID)
			if err != nil {
				return nil, err
			}

			// ensure user is a member of the organization
			orgMemberID, err := m.Client().OrgMembership.Query().
				Where(orgmembership.UserID(userID)).
				Where(orgmembership.OrganizationID(program.OwnerID)).
				OnlyID(ctx)
			if err != nil || orgMemberID == "" {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get org membership, cannot add user to program")

				return nil, ErrUserNotInOrg
			}

			m.SetOrgMembershipID(orgMemberID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
