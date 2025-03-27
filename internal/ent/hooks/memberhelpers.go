package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// MutationMember is an interface that can be implemented by a member mutation to get IDs
type MutationMember interface {
	UserIDs() []string
	UserID() (string, bool)
	ID() (string, bool)
	IDs(ctx context.Context) ([]string, error)
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
			userID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			if m.Op().Is(ent.OpCreate) {
				// Only run this hook on membership mutations
				// you can create a group/program/etc with yourself as a member
				if !checkMutation(ctx) {
					return next.Mutate(ctx, m)
				}

				if err := createMembershipCheck(mutationMember, userID); err != nil {
					zerolog.Ctx(ctx).Error().Msg("cannot create membership")

					return nil, err
				}

				return next.Mutate(ctx, m)
			}

			// if its not a create, check for updates with the user instead. This includes
			// update and delete operations
			if err := updateMembershipCheck(ctx, mutationMember, table, userID); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}
}

// createMembershipCheck is a helper function to check if a user is trying to add themselves to a membership
func createMembershipCheck(m MutationMember, actorID string) error {
	userIds := m.UserIDs()
	if len(userIds) == 0 {
		userID, ok := m.UserID()
		if !ok {
			return nil
		}

		userIds = append(userIds, userID)
	}

	for _, userID := range userIds {
		if userID == actorID {
			return generated.ErrPermissionDenied
		}
	}

	return nil
}

// updateMembershipCheck is a helper function to check if a user is trying to update themselves in a membership
func updateMembershipCheck(ctx context.Context, m MutationMember, table string, actorID string) error {
	memberIDs := getMutationMemberIDs(ctx, m)
	if len(memberIDs) == 0 {
		return nil
	}

	query := "SELECT user_id FROM " + table + " WHERE id in ($1)"

	var rows sql.Rows
	if err := generated.FromContext(ctx).Driver().Query(ctx, query, []any{strings.Join(memberIDs, ",")}, &rows); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to get user ID from membership")

		return err
	}

	defer rows.Close()

	if rows.Next() {
		var userID string

		if err := rows.Scan(&userID); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to scan user ID from membership")

			return err
		}

		if userID == actorID {
			zerolog.Ctx(ctx).Error().Msg("user cannot update their own membership")

			return generated.ErrPermissionDenied
		}
	}

	return nil
}

// getMutationMemberIDs is a helper function to get the member IDs from a mutation
// this can be used for group, program, and org membership mutations because
// they all implement the MutationMember interface
func getMutationMemberIDs(ctx context.Context, m MutationMember) []string {
	id, ok := m.ID()
	if ok {
		return []string{id}
	}

	ids, err := m.IDs(ctx)
	if err == nil && len(ids) > 0 {
		return ids
	}

	return ids
}

func checkMutation(ctx context.Context) bool {
	rootFieldCtx := graphql.GetRootFieldContext(ctx)

	// skip if not a graphql mutation
	if rootFieldCtx == nil {
		return false
	}

	// Check if the mutation is a membership mutation
	if strings.Contains(rootFieldCtx.Object, "Membership") {
		return true
	}

	return false
}
