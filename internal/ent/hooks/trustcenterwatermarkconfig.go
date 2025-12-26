package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/pkg/jobspec"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
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
				if _, err := m.Job.Insert(ctx, jobspec.WatermarkDocArgs{
					TrustCenterDocumentID: doc.ID,
				}, nil); err != nil {
					return nil, err
				}
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// checkTrustCenterWatermarkConfigFiles checks for logo files in the context
func checkTrustCenterWatermarkConfigFiles(ctx context.Context, m *generated.TrustCenterWatermarkConfigMutation) (context.Context, error) {
	key := "logoFile"

	logoFiles, _ := objects.FilesFromContextWithKey(ctx, key)
	if len(logoFiles) > 1 {
		return ctx, ErrNotSingularUpload
	}

	if len(logoFiles) == 1 {
		m.SetFileID(logoFiles[0].ID)
		m.ClearText()

		adapter := objects.NewGenericMutationAdapter(m,
			func(mut *generated.TrustCenterWatermarkConfigMutation) (string, bool) { return mut.ID() },
			func(mut *generated.TrustCenterWatermarkConfigMutation) string { return mut.Type() },
		)

		return objects.ProcessFilesForMutation(ctx, adapter, key, "trust_center_watermark_config")
	}

	return ctx, nil
}
