package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/pkg/logx"
)

// HookTrustCenterWatermarkConfig process files for trust center watermark config
func HookTrustCenterWatermarkConfig() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterWatermarkConfigFunc(func(ctx context.Context, m *generated.TrustCenterWatermarkConfigMutation) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("trust center watermark config hook")

			var err error

			ctx, err = checkTrustCenterWatermarkConfigFiles(ctx, m)
			if err != nil {
				return nil, err
			}

			if text, textSet := m.Text(); textSet && text != "" {
				m.ClearLogoID()
				m.ClearFile()
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			config, ok := v.(*generated.TrustCenterWatermarkConfig)
			if !ok {
				return v, nil
			}

			// trigger updates of all the TrustCenterDocs that have watermarking enabled
			docs, err := m.Client().TrustCenterDoc.Query().
				Where(trustcenterdoc.WatermarkingEnabled(true)).
				Where(trustcenterdoc.TrustCenterID(config.TrustCenterID)).All(ctx)
			if err != nil {
				return nil, err
			}

			for _, doc := range docs {
				if err := enqueueJob(ctx, m.Job, jobspec.WatermarkDocArgs{
					TrustCenterDocumentID: doc.ID,
				}, nil); err != nil {
					return nil, err
				}
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// checkTrustCenterWatermarkConfigFiles checks for watermark config files in the context
func checkTrustCenterWatermarkConfigFiles(ctx context.Context, m *generated.TrustCenterWatermarkConfigMutation) (context.Context, error) {
	key := "watermarkFile"
	ctx, err := processSingleMutationFile(ctx, m, key, "trust_center_watermark_config", ErrNotSingularUpload,
		func(mut *generated.TrustCenterWatermarkConfigMutation, id string) {
			mut.SetFileID(id)
			mut.ClearText()
		},
	)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
