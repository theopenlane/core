package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks/contextx"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// HookDeletePermissions is an ent hook that deletes all relationship tuples associated with an object
// on either delete or soft-delete operations
func HookDeletePermissions() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			mut, ok := m.(utils.GenericMutation)
			if !ok {
				logx.FromContext(ctx).Warn().Msg("DeletePermissionsHook: mutation does not implement GenericMutation, skipping")
				return next.Mutate(ctx, m)
			}

			// the ids have to be resolved before the mutation runs, a bulk hard delete resolves its
			// predicate against rows that no longer exist once the delete has executed
			objIDs, err := getMutationIDs(ctx, mut)
			if err != nil {
				return nil, err
			}

			// run the mutation first
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return retVal, err
			}

			// then delete the permissions
			if err := deletePermissionsForIDs(ctx, mut, objIDs); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	},
		hook.HasOp(ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne),
	)
}

// DeletePermissionsHook deletes all relationship tuples associated with the object(s) in the mutation.
// The ids are resolved from the mutation, so this must be called before the records are hard deleted
func DeletePermissionsHook(ctx context.Context, m utils.GenericMutation) error {
	ids, err := getMutationIDs(ctx, m)
	if err != nil {
		return err
	}
	return deletePermissionsForIDs(ctx, m, ids)
}

// deletePermissionsForIDs deletes all relationship tuples for the given object ids
func deletePermissionsForIDs(ctx context.Context, m utils.GenericMutation, objIDs []string) error {
	client := utils.AuthzClientFromContext(ctx)
	if client == nil {
		logx.FromContext(ctx).Warn().Msg("Authz client not found in context, skipping deleting relationship tuples")
		return nil
	}

	if skipDeleteHook(ctx, m) {
		logx.FromContext(ctx).Debug().Msg("skipping delete permissions hook")

		return nil
	}

	if len(objIDs) == 0 {
		logx.FromContext(ctx).Debug().Msg("no object IDs found in mutation, skipping deleting relationship tuples")

		return nil
	}

	for _, objID := range objIDs {
		objType := strcase.SnakeCase(m.Type())
		object := fmt.Sprintf("%s:%s", objType, objID)

		logx.FromContext(ctx).Debug().Str("object", object).Msg("deleting relationship tuples")

		if err := client.DeleteAllObjectRelations(ctx, object, []string{}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to delete relationship tuples")

			return ErrInternalServerError
		}

		logx.FromContext(ctx).Debug().Str("object", object).Msg("deleted relationship tuples")
	}

	return nil
}

// skipDeleteHook checks if the delete hook should be skipped based on the context and mutation
func skipDeleteHook(ctx context.Context, m utils.GenericMutation) bool {
	// memberships go through the auth from mutation hooks as a special case
	if strings.Contains(m.Type(), "Membership") {
		return true
	}

	// the organization cascade delete runs as an internal request so it can bypass privacy rules,
	// but the records it removes still need their tuples cleaned up, so it opts back in explicitly
	if contextx.TupleCleanupEnabled(ctx) {
		return false
	}

	// skip if internal request
	if rule.IsInternalRequest(ctx) {
		return true
	}

	return false
}

// getTupleKeyFromRole creates a Tuple key with the provided subject, object, and role
func getTupleKeyFromRole(req fgax.TupleRequest, role enums.Role) (fgax.TupleKey, error) {
	fgaRelation, err := roleToRelation(role)
	if err != nil {
		return fgax.NewTupleKey(), err
	}

	req.Relation = fgaRelation

	return fgax.GetTupleKey(req), nil
}

func roleToRelation(r enums.Role) (string, error) {
	switch r {
	case enums.RoleOwner, enums.RoleAdmin, enums.RoleMember:
		return strings.ToLower(r.String()), nil
	case fgax.ParentRelation, fgax.ParentContextRelation:
		return r.String(), nil
	default:
		return "", ErrUnsupportedFGARole
	}
}
