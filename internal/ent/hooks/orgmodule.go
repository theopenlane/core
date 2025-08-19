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

func HookOrgModuleUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgModuleFunc(func(ctx context.Context, omm *generated.OrgModuleMutation) (generated.Value, error) {
			if !omm.EntConfig.Modules.Enabled {
				return next.Mutate(ctx, omm)
			}

			op := omm.Op()

			// Handle update operations
			if op == ent.OpUpdateOne {
				// Check if active field is being updated
				newActive, activeBeingSet := omm.Active()
				if !activeBeingSet {
					// Active field is not being updated, proceed normally
					return next.Mutate(ctx, omm)
				}

				id, exists := omm.ID()
				if !exists {
					return next.Mutate(ctx, omm)
				}

				// Get old active status to detect transitions
				oldActive, err := omm.OldActive(ctx)
				if err != nil {
					return nil, err
				}

				// Execute the mutation first
				v, err := next.Mutate(ctx, omm)
				if err != nil {
					return nil, err
				}

				// Handle activation (inactive -> active)
				if !oldActive && newActive {
					// Get the module info after update
					orgModule, ok := v.(*generated.OrgModule)
					if !ok {
						return v, fmt.Errorf("%w: unable to cast to OrgModule", ErrFieldRequired)
					}

					feats := []models.OrgModule{orgModule.Module}

					if err := createFeatureTuples(ctx, omm.Authz, orgModule.OwnerID, feats); err != nil {
						zerolog.Ctx(ctx).Error().Err(err).Msg("error creating feature tuples on activation")
						return nil, err
					}
				}

				// Handle deactivation (active -> inactive)
				if oldActive && !newActive {
					// Get the module info for deletion
					moduleToDeactivate, err := omm.Client().OrgModule.Get(ctx, id)
					if err != nil {
						return nil, err
					}

					deleteTuple := fgax.GetTupleKey(fgax.TupleRequest{
						SubjectID:   moduleToDeactivate.OwnerID,
						SubjectType: generated.TypeOrganization,
						ObjectID:    moduleToDeactivate.Module.String(),
						ObjectType:  "feature",
						Relation:    "enabled",
					})

					_, err = omm.Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{deleteTuple})
					if err != nil {
						return nil, err
					}
				}

				return v, nil
			}

			// Handle delete operations
			if op == ent.OpDeleteOne {
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

				deleteTuple := fgax.GetTupleKey(fgax.TupleRequest{
					SubjectID:   moduleToDelete.OwnerID,
					SubjectType: generated.TypeOrganization,
					ObjectID:    moduleToDelete.Module.String(),
					ObjectType:  "feature",
					Relation:    "enabled",
				})

				_, err = omm.Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{deleteTuple})
				if err != nil {
					return nil, err
				}

				return v, nil
			}

			// For other operations, proceed normally
			return next.Mutate(ctx, omm)
		})
	}, ent.OpUpdateOne|ent.OpDeleteOne)
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
