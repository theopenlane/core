package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/models"
)

// HookOrgModule adds the feature tuples to fga as needed
func HookOrgModule() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgModuleFunc(func(ctx context.Context, omm *generated.OrgModuleMutation) (generated.Value, error) {
			if !omm.EntitlementManager.Config.IsEnabled() {
				return next.Mutate(ctx, omm)
			}

			v, err := next.Mutate(ctx, omm)
			if err != nil {
				return nil, err
			}

			orgID, ok := omm.OwnerID()
			if !ok || orgID == "" {
				return v, fmt.Errorf("%w: owner_id", ErrFieldRequired)
			}

			orgModule, ok := v.(*generated.OrgModule)
			if !ok {
				return v, fmt.Errorf("%w: owner_id", ErrFieldRequired)
			}

			feats := []models.OrgModule{orgModule.Module}

			if err := entitlements.CreateFeatureTuples(ctx, &omm.Authz, orgID, feats); err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("error creating feature tuples")
				return nil, err
			}

			return v, nil
		})
	}, ent.OpCreate)
}

// HookOrgModuleUpdate updates the feature tuple in fga based off the module status in the database
func HookOrgModuleUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgModuleFunc(func(ctx context.Context, omm *generated.OrgModuleMutation) (generated.Value, error) {
			if !omm.EntitlementManager.Config.IsEnabled() {
				return next.Mutate(ctx, omm)
			}

			switch omm.Op() {
			case ent.OpUpdateOne:
				return handleOrgModuleUpdate(ctx, omm, next)
			case ent.OpDeleteOne:
				return handleOrgModuleDelete(ctx, omm, next)
			default:
				return next.Mutate(ctx, omm)
			}
		})
	}, ent.OpUpdateOne|ent.OpDeleteOne)
}

func handleOrgModuleUpdate(ctx context.Context, omm *generated.OrgModuleMutation, next ent.Mutator) (generated.Value, error) {
	newActive, newActiveExists := omm.Active()
	if !newActiveExists {
		return next.Mutate(ctx, omm)
	}

	id, exists := omm.ID()
	if !exists {
		return next.Mutate(ctx, omm)
	}

	oldActive, err := omm.OldActive(ctx)
	if err != nil {
		return nil, err
	}

	v, err := next.Mutate(ctx, omm)
	if err != nil {
		return nil, err
	}

	switch {
	case !oldActive && newActive:
		return handleActivation(ctx, omm, v)
	case oldActive && !newActive:
		return handleDeactivation(ctx, omm, id)
	default:
		return v, nil
	}
}

func handleActivation(ctx context.Context, omm *generated.OrgModuleMutation, v generated.Value) (generated.Value, error) {
	orgModule, ok := v.(*generated.OrgModule)
	if !ok {
		// should not really happen
		return v, fmt.Errorf("%w: unable to cast to OrgModule", ErrFieldRequired)
	}

	feats := []models.OrgModule{orgModule.Module}
	if err := entitlements.CreateFeatureTuples(ctx, &omm.Authz, orgModule.OwnerID, feats); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating feature tuples on activating module again")
		return nil, err
	}

	return v, nil
}

func handleDeactivation(ctx context.Context, omm *generated.OrgModuleMutation, id string) (generated.Value, error) {
	module, err := omm.Client().OrgModule.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := entitlements.DeleteModuleTuple(ctx, &omm.Authz, module.OwnerID, string(module.Module.String())); err != nil {
		return nil, err
	}

	return module, nil
}

func handleOrgModuleDelete(ctx context.Context, omm *generated.OrgModuleMutation, next ent.Mutator) (generated.Value, error) {
	id, exists := omm.ID()
	if !exists {
		return next.Mutate(ctx, omm)
	}

	moduleToDelete, err := omm.Client().OrgModule.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	v, err := next.Mutate(ctx, omm)
	if err != nil {
		return nil, err
	}

	if err := entitlements.DeleteModuleTuple(ctx, &omm.Authz, moduleToDelete.OwnerID, string(moduleToDelete.Module.String())); err != nil {
		return nil, err
	}

	return v, nil
}
