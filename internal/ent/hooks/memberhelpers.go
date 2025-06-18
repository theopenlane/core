package hooks

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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

			id, ok := mutationMember.ID()
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrInvalidInput, "id is required")
			}

			if err := updateMembershipCheck(ctx, mutationMember, table, userID); err != nil {
				return nil, err
			}

			orgMembership, err := mutationMember.Client().OrgMembership.Get(ctx, id)
			if err != nil {
				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if err := deleteOrgMembershipFGATuples(ctx, orgMembership, mutationMember.Client()); err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("failed to delete FGA tuples after successful DB deletion")
				return nil, err
			}

			return retVal, err
		})
	}
}

func deleteOrgMembershipFGATuples(ctx context.Context, orgMembership *generated.OrgMembership,
	client *generated.Client) error {

	req := fgax.TupleRequest{
		SubjectID:   orgMembership.UserID,
		SubjectType: generated.TypeUser,
		ObjectID:    orgMembership.OrganizationID,
		ObjectType:  generated.TypeOrganization,
		Relation:    orgMembership.Role.String(),
	}

	tuple := fgax.GetTupleKey(req)

	if _, err := client.Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{tuple}); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Interface("delete_tuple", tuple).Msg("failed to delete relationship tuple")
		return err
	}

	return nil
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
		return generated.ErrPermissionDenied
	}

	return nil
}

// updateMembershipCheck is a helper function to check if a user is trying to update themselves in a membership
func updateMembershipCheck(ctx context.Context, m MutationMember, table string, actorID string) error {
	memberIDs := getMutationMemberIDs(ctx, m)
	if len(memberIDs) == 0 {
		return nil
	}

	// only deletes allowed by a user on the org_memberships table
	if table == "org_memberships" && (m.Op().Is(ent.OpDelete | ent.OpDeleteOne)) {
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
