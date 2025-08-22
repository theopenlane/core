package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/permissioncache"
	"github.com/theopenlane/iam/fgax"
)

// HookOrgModule adds the feature tuples to fga as needed
func HookOrgModule() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgModuleFunc(func(ctx context.Context, omm *generated.OrgModuleMutation) (generated.Value, error) {

			if !omm.EntConfig.Modules.Enabled {
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

			if err := createFeatureTuples(ctx, omm.Authz, orgID, feats); err != nil {
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
			if !omm.EntConfig.Modules.Enabled {
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
	if err := createFeatureTuples(ctx, omm.Authz, orgModule.OwnerID, feats); err != nil {
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

	if err := deleteModuleTuple(ctx, omm.Authz, module); err != nil {
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

	if err := deleteModuleTuple(ctx, omm.Authz, moduleToDelete); err != nil {
		return nil, err
	}

	return v, nil
}

func deleteModuleTuple(ctx context.Context, authz fgax.Client, module *generated.OrgModule) error {
	deleteTuple := fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   module.OwnerID,
		SubjectType: generated.TypeOrganization,
		ObjectID:    module.Module.String(),
		ObjectType:  "feature",
		Relation:    "enabled",
	})

	_, err := authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{deleteTuple})
	return err
}

// createFeatureTuples writes default feature tuples to FGA and inserts them into
// the feature cache if available.
func createFeatureTuples(ctx context.Context, authz fgax.Client, orgID string, feats []models.OrgModule) error {
	tuples := make([]fgax.TupleKey, 0, len(feats))
	for _, f := range feats {
		tuples = append(tuples, fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   orgID,
			SubjectType: generated.TypeOrganization,
			ObjectID:    f.String(),
			ObjectType:  "feature",
			Relation:    "enabled",
		}))
	}

	if _, err := authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
		return err
	}

	if cache, ok := permissioncache.CacheFromContext(ctx); ok {
		return cache.SetFeatures(ctx, orgID, feats)
	}

	return nil
}
