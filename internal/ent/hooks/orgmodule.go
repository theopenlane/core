package hooks

import (
	"context"
	"errors"

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

			orgID, ok := omm.OwnerID()
			if !ok || orgID == "" {
				return v, errors.New("ownber id not exists")
			}

			orgModule, ok := v.(*generated.OrgModule)
			if !ok {
				return v, errors.New("ownber id not exists")
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
