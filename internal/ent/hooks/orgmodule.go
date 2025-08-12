package hooks

import (
	"context"

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

			v, err := next.Mutate(ctx, omm)
			if err != nil {
				return nil, err
			}

			ids, err := omm.IDs(ctx)
			if err != nil {
				return nil, err
			}

			orgID, ok := omm.OwnerID()
			if !ok && orgID != "" {
				return nil, err
			}

			var feats = make([]models.OrgModule, 0, len(ids))

			for _, id := range ids {
				module, err := omm.Client().OrgModule.Get(ctx, id)
				if err != nil {
					return nil, err
				}

				feats = append(feats, module.Module)
			}

			if err := createFeatureTuples(ctx, omm.Authz, orgID, feats); err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("error creating feature tuples")
				return nil, err
			}

			return v, nil
		})
	}, ent.OpCreate)
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
