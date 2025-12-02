package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/orgmembership"
	"github.com/theopenlane/ent/generated/privacy"
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

			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			program, err := m.Client().Program.Get(allowCtx, programID)
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
