package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
)

func HookTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center doc hook")

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterDocFile(ctx, m)
				if err != nil {
					return nil, err
				}
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

func checkTrustCenterDocFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, error) {
	dockey := "trustCenterDocFile"

	// get the file from the context, if it exists
	docFile, _ := objects.FilesFromContextWithKey(ctx, dockey)
	if docFile == nil {
		return ctx, nil
	}

	// this should always be true, but check just in case
	if docFile[0].FieldName == dockey {
		// we should only have one file
		if len(docFile) > 1 {
			return ctx, ErrNotSingularUpload
		}
		m.SetFileID(docFile[0].ID)

		docFile[0].Parent.ID, _ = m.ID()
		docFile[0].Parent.Type = "trust_center_doc"

		ctx = objects.UpdateFileInContextByKey(ctx, dockey, docFile[0])
	}

	return ctx, nil
}
