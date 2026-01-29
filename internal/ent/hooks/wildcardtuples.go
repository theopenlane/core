package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

type systemOwnedMutation interface {
	SystemOwned() (bool, bool)
	OldSystemOwned(context.Context) (bool, error)
}

// HookCreatePublicAccess adds public access (wildcard tuples) to the created object
// Deletion of tuples is handled by the global HookDeletePermissions hook
func HookPublicAccess() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if !auth.IsSystemAdminFromContext(ctx) {
				return retVal, nil
			}

			systemOwnedMutation, ok := m.(systemOwnedMutation)
			if !ok {
				logx.FromContext(ctx).Warn().Str("mutation", m.Type()).Str("hook", "HookPublicAccess").Msg("mutation does not implement SystemOwnedMutation, skipping")

				return retVal, nil
			}

			systemOwned, _ := systemOwnedMutation.SystemOwned()
			oldSystemOwned, _ := systemOwnedMutation.OldSystemOwned(ctx)

			if (!systemOwned) && !oldSystemOwned {
				logx.FromContext(ctx).Debug().Msg("object is not system owned, skipping public access hook")

				return retVal, nil
			}

			genericMut, ok := m.(utils.GenericMutation)
			if !ok {
				logx.FromContext(ctx).Warn().Str("mutation", m.Type()).Str("hook", "HookPublicAccess").Msg("mutation does not implement GenericMutation, skipping")
				return retVal, nil
			}

			objID, exists := genericMut.ID()
			if !exists {
				return nil, nil
			}

			objectType := strcase.SnakeCase(genericMut.Type())
			wildcardTuple := fgax.CreateWildcardViewerTuple(objID, objectType)
			logx.FromContext(ctx).Debug().Interface("request", wildcardTuple).
				Msg("creating public viewer relationship tuples")

			if _, err := genericMut.Client().Authz.WriteTupleKeys(ctx, wildcardTuple, nil); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to create public viewer relationship tuples")

				return nil, err
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}
