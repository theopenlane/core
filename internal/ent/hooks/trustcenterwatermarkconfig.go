package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
)

func HookTrustCenterWatermarkConfig() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterWatermarkConfigFunc(func(ctx context.Context, m *generated.TrustCenterWatermarkConfigMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center watermark config hook")

			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterWatermarkConfigFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.SetFileID(fileIDs[0])
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkTrustCenterWatermarkConfigFiles(ctx context.Context, m *generated.TrustCenterWatermarkConfigMutation) (context.Context, error) {
	key := "logoFile"
	// get the file from the context, if it exists
	file, _ := objects.FilesFromContextWithKey(ctx, key)

	// return early if no file is provided
	if file == nil {
		return ctx, nil
	}

	// we should only have one file
	if len(file) > 1 {
		return ctx, ErrNotSingularUpload
	}
	if len(file) == 0 {
		return ctx, nil
	}
	if file[0].FieldName != key {
		return ctx, nil
	}

	file[0].Parent.ID, _ = m.ID()
	file[0].Parent.Type = "trust_center_watermark_config"

	ctx = objects.UpdateFileInContextByKey(ctx, key, file[0])

	return ctx, nil
}
