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

			// run the mutation first
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return retVal, err
			}

			// then delete the permissions
			if err := DeletePermissionsHook(ctx, mut); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	},
		hook.HasOp(ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne),
	)
}

// DeletePermissionsHook deletes all relationship tuples associated with the object(s) in the mutation
func DeletePermissionsHook(ctx context.Context, m utils.GenericMutation) error {
	client := utils.AuthzClientFromContext(ctx)
	if client == nil {
		logx.FromContext(ctx).Warn().Msg("Authz client not found in context, skipping deleting relationship tuples")
		return nil
	}

	if skipDeleteHook(ctx, m) {
		logx.FromContext(ctx).Debug().Msg("skipping delete permissions hook")

		return nil
	}

	objIDs := getMutationIDs(ctx, m)
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
	// skip if internal request
	if rule.IsInternalRequest(ctx) {
		return true
	}

	// memberships go through the auth from mutation hooks as a special case
	if strings.Contains(m.Type(), "Membership") {
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
	case fgax.ParentRelation:
		return r.String(), nil
	default:
		return "", ErrUnsupportedFGARole
	}
}
