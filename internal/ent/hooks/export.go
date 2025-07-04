package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
)

func HookExport() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ExportFunc(
			func(ctx context.Context, m *generated.ExportMutation) (generated.Value, error) {

				// validate the id exists on the type
				exportType, ok := m.ExportType()
				if !ok {
					return nil, errors.New("provide export type")
				}

				itemID, ok := m.ItemID()
				if !ok {
					return nil, errors.New("provide item id")
				}

				switch exportType {
				case enums.ExportTypeControl:
					_, err := m.Client().Control.Get(ctx, itemID)
					if err != nil {
						return nil, err
					}

				default:
					return nil, errors.New("export id and type invalid")
				}

				if m.Op().Is(ent.OpCreate) {
					requestorID, err := auth.GetSubjectIDFromContext(ctx)
					if err != nil {
						return nil, err
					}

					m.SetRequestorID(requestorID)

					return next.Mutate(ctx, m)
				}

				// Handle file uploads on update path only
				if m.Op().Is(ent.OpUpdate | ent.OpUpdateOne) {
					fileIDs := objects.GetFileIDsFromContext(ctx)
					if len(fileIDs) > 0 {
						var err error

						ctx, err = checkExportFiles(ctx, m)
						if err != nil {
							return nil, err
						}

						m.AddFileIDs(fileIDs...)
					}
				}

				return next.Mutate(ctx, m)
			})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// checkExportFiles checks if export files are provided and sets the local file ID(s)
func checkExportFiles(ctx context.Context, m *generated.ExportMutation) (context.Context, error) {
	key := "exportFiles"

	// get the file from the context, if it exists
	file, err := objects.FilesFromContextWithKey(ctx, key)
	if err != nil {
		return ctx, err
	}

	if file == nil {
		return ctx, nil
	}

	for i, f := range file {
		if f.FieldName == key {
			file[i].Parent.ID, _ = m.ID()
			file[i].Parent.Type = m.Type()

			ctx = objects.UpdateFileInContextByKey(ctx, key, file[i])
		}
	}

	return ctx, nil
}
