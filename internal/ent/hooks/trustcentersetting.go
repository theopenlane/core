package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
)

var ErrTooManyLogoFiles = errors.New("too many logo files uploaded, only one is allowed")
var ErrTooManyFaviconFiles = errors.New("too many favicon files uploaded, only one is allowed")

func HookTrustCenterSetting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center setting hook")

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkTrustCenterFiles(ctx context.Context, m *generated.TrustCenterSettingMutation) (context.Context, error) {
	logoKey := "logoFile"
	faviconKey := "faviconFile"

	// get the file from the context, if it exists
	logoFile, _ := objects.FilesFromContextWithKey(ctx, logoKey)
	faviconFile, _ := objects.FilesFromContextWithKey(ctx, faviconKey)

	// this should always be true, but check just in case
	if logoFile != nil && logoFile[0].FieldName == logoKey {
		// we should only have one file
		if len(logoFile) > 1 {
			return ctx, ErrTooManyLogoFiles
		}

		m.SetLogoLocalFileID(logoFile[0].ID)

		logoFile[0].Parent.ID, _ = m.ID()
		logoFile[0].Parent.Type = "trust_center_setting"

		ctx = objects.UpdateFileInContextByKey(ctx, logoKey, logoFile[0])
	}

	if faviconFile != nil && faviconFile[0].FieldName == faviconKey {
		if len(faviconFile) > 1 {
			return ctx, ErrTooManyFaviconFiles
		}

		m.SetFaviconLocalFileID(faviconFile[0].ID)

		faviconFile[0].Parent.ID, _ = m.ID()
		faviconFile[0].Parent.Type = "trust_center_setting"

		ctx = objects.UpdateFileInContextByKey(ctx, faviconKey, faviconFile[0])
	}

	return ctx, nil
}
