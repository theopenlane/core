package hooks

import (
	"context"
	"errors"
	"slices"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/ent/privacy/utils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/shared/logx"
)

// MutationMember is an interface that can be implemented by a member mutation to get IDs
type MutationMember interface {
	UserIDs() []string
	UserID() (string, bool)
	ID() (string, bool)
	IDs(ctx context.Context) ([]string, error)
	Op() ent.Op
	Client() *generated.Client
}

// HookMembershipSelf is a hook that runs on membership mutations
// to prevent users from updating their own membership
func HookMembershipSelf(table string) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// bypass privacy check if the context allows it
			if _, allow := privacy.DecisionFromContext(ctx); allow {
				return next.Mutate(ctx, m)
			}

			mutationMember, ok := m.(MutationMember)
			if !ok {
				return next.Mutate(ctx, m)
			}

			// check if group member is the authenticated user
			au, err := auth.GetAuthenticatedUserFromContext(ctx)
			if err != nil {
				return nil, err
			}

			// if the user is an org owner, skip the check
			if err := rule.CheckCurrentOrgAccess(ctx, nil, fgax.OwnerRelation); errors.Is(err, privacy.Allow) {
				// ensure this is not an org membership mutation, owners cannot update their own membership
				// in the organization, it must be done via a transfer
				if m.Type() != generated.TypeOrgMembership {
					return next.Mutate(ctx, m)
				}
			}

			if m.Op().Is(ent.OpCreate) {
				// Only run this hook on membership mutations
				if !checkMutation(ctx) {
					return next.Mutate(ctx, m)
				}

				if err := createMembershipCheck(mutationMember, au.SubjectID); err != nil {
					logx.FromContext(ctx).Error().Msg("cannot create membership")

					return nil, err
				}
			}

			if err := updateMembershipCheck(ctx, mutationMember, table, au.SubjectID); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}
}

// createMembershipCheck is a helper function to check if a user is trying to add themselves to a membership
func createMembershipCheck(m MutationMember, actorID string) error {
	userIDs := m.UserIDs()
	if len(userIDs) == 0 {
		userID, ok := m.UserID()
		if !ok {
			return nil
		}

		userIDs = append(userIDs, userID)
	}

	if slices.Contains(userIDs, actorID) {
		log.Warn().Str("user_id", actorID).Msg("user attempting to create membership for themselves")

		return generated.ErrPermissionDenied
	}

	return nil
}

// updateMembershipCheck is a helper function to check if a user is trying to update themselves in a membership
func updateMembershipCheck(ctx context.Context, m MutationMember, table string, actorID string) error {
	mut := m.(utils.GenericMutation)
	memberIDs := getMutationIDs(ctx, mut)
	if len(memberIDs) == 0 {
		return nil
	}

	query := "SELECT user_id FROM " + table + " WHERE id in ($1)"

	var rows sql.Rows
	if err := generated.FromContext(ctx).Driver().Query(ctx, query, []any{strings.Join(memberIDs, ",")}, &rows); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get user ID from membership")

		return err
	}

	defer rows.Close()

	if rows.Next() {
		var userID string

		if err := rows.Scan(&userID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to scan user ID from membership")

			return err
		}

		if userID == actorID {
			logx.FromContext(ctx).Error().Msg("user cannot update their own membership")

			return generated.ErrPermissionDenied
		}
	}

	return nil
}

func checkMutation(ctx context.Context) bool {
	rootFieldCtx := graphql.GetRootFieldContext(ctx)

	// skip if not a graphql mutation
	if rootFieldCtx == nil {
		return false
	}

	// Check if the mutation is a group creation with members
	if strings.Contains(rootFieldCtx.Object, "createGroupWithMembers") {
		return false
	}

	// Check if the mutation is a membership mutation
	if strings.Contains(rootFieldCtx.Object, "Membership") {
		return true
	}

	return false
}
