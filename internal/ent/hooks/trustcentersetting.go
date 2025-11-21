package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
)

var ErrTooManyLogoFiles = errors.New("too many logo files uploaded, only one is allowed")
var ErrTooManyFaviconFiles = errors.New("too many favicon files uploaded, only one is allowed")

func HookTrustCenterSetting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("trust center setting hook")

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterFiles(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkTrustCenterFiles(ctx context.Context, m *generated.TrustCenterSettingMutation) (context.Context, error) {
	logoKey := "logoFile"
	faviconKey := "faviconFile"

	logoFiles, _ := objects.FilesFromContextWithKey(ctx, logoKey)
	if len(logoFiles) > 1 {
		return ctx, ErrTooManyLogoFiles
	}
	if len(logoFiles) == 1 {
		m.SetLogoLocalFileID(logoFiles[0].ID)

		adapter := objects.NewGenericMutationAdapter(m,
			func(mut *generated.TrustCenterSettingMutation) (string, bool) { return mut.ID() },
			func(mut *generated.TrustCenterSettingMutation) string { return mut.Type() },
		)

		ctx, _ = objects.ProcessFilesForMutation(ctx, adapter, logoKey, "trust_center_setting")
	}

	faviconFiles, _ := objects.FilesFromContextWithKey(ctx, faviconKey)
	if len(faviconFiles) > 1 {
		return ctx, ErrTooManyFaviconFiles
	}
	if len(faviconFiles) == 1 {
		m.SetFaviconLocalFileID(faviconFiles[0].ID)

		adapter := objects.NewGenericMutationAdapter(m,
			func(mut *generated.TrustCenterSettingMutation) (string, bool) { return mut.ID() },
			func(mut *generated.TrustCenterSettingMutation) string { return mut.Type() },
		)

		ctx, _ = objects.ProcessFilesForMutation(ctx, adapter, faviconKey, "trust_center_setting")
	}

	return ctx, nil
}
