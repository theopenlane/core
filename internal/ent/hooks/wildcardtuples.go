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

type trustCenterChildMutation interface {
	TrustCenterID() (string, bool)
	OldTrustCenterID(context.Context) (string, error)
}

// HookCreatePublicAccess adds public access (wildcard tuples) to the created object for
// system owned objects.
// Deletion of tuples is handled by the global HookDeletePermissions hook
func HookPublicAccess() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// // if it is a trust center child, create public access based on trust center
			// if _, ok := m.(trustCenterChildMutation); ok {
			// 	if err := createTrustCenterPublicAccess(ctx, m); err != nil {
			// 		return nil, err
			// 	}
			// }

			// all other mutations, only create public access for system admins
			if !auth.IsSystemAdminFromContext(ctx) {
				return retVal, nil
			}

			// add wildcard tuples for public access if the object is system owned
			if _, ok := m.(systemOwnedMutation); ok {
				if err := createSystemOwnedPublicAccess(ctx, m); err != nil {
					return nil, err
				}
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}

// createSystemOwnedPublicAccess creates wildcard viewer tuples for system owned objects
func createSystemOwnedPublicAccess(ctx context.Context, m ent.Mutation) error {
	systemOwnedMutation, ok := m.(systemOwnedMutation)
	if !ok {
		logx.FromContext(ctx).Warn().Str("mutation", m.Type()).Str("hook", "HookPublicAccess").Msg("mutation does not implement SystemOwnedMutation, skipping")

		return nil
	}

	systemOwned, _ := systemOwnedMutation.SystemOwned()
	oldSystemOwned, _ := systemOwnedMutation.OldSystemOwned(ctx)

	if (!systemOwned) && !oldSystemOwned {
		logx.FromContext(ctx).Debug().Msg("object is not system owned, skipping public access hook")

		return nil
	}

	return createWildcardTuple(ctx, m)
}

// createTrustCenterPublicAccess creates wildcard viewer tuples for trust center child objects
func createTrustCenterPublicAccess(ctx context.Context, m ent.Mutation) (err error) {
	mut, _ := m.(trustCenterChildMutation)
	trustCenterID, ok := mut.TrustCenterID()

	if !ok || trustCenterID == "" {
		// check old trust center id for updates
		trustCenterID, err = mut.OldTrustCenterID(ctx)
		if err != nil || trustCenterID == "" {
			logx.FromContext(ctx).Debug().Msg("object does not have a trust center ID, skipping public access hook")

			return nil
		}
	}

	return createWildcardTuple(ctx, m)
}

// createWildcardTuple creates a wildcard viewer tuple for the object created in the mutation
func createWildcardTuple(ctx context.Context, m ent.Mutation) error {
	genericMut, ok := m.(utils.GenericMutation)
	if !ok {
		logx.FromContext(ctx).Warn().Str("mutation", m.Type()).Str("hook", "HookPublicAccess").Msg("mutation does not implement GenericMutation, skipping")
		return nil
	}

	objID, exists := genericMut.ID()
	if !exists {
		return nil
	}

	objectType := strcase.SnakeCase(genericMut.Type())
	wildcardTuple := fgax.CreateWildcardViewerTuple(objID, objectType)

	logx.FromContext(ctx).Debug().Interface("request", wildcardTuple).
		Msg("creating public viewer relationship tuples")

	if _, err := genericMut.Client().Authz.WriteTupleKeys(ctx, wildcardTuple, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create public viewer relationship tuples")

		return err
	}

	return nil
}
