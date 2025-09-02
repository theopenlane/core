package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
)

var (
	errExportTypeNotProvided = errors.New("provide export type")
	errFieldsNotProvided     = errors.New("at least one field must be provided for the schema to export")
)

func HookExport() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ExportFunc(
			func(ctx context.Context, m *generated.ExportMutation) (generated.Value, error) {
				if m.Op().Is(ent.OpCreate) {
					return handleExportCreate(ctx, m, next)
				}

				return handleExportUpdate(ctx, m, next)
			})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

func handleExportCreate(ctx context.Context, m *generated.ExportMutation, next ent.Mutator) (generated.Value, error) {
	exportType, ok := m.ExportType()
	if !ok {
		return nil, errExportTypeNotProvided
	}

	if err := ValidateExportType(exportType.String()); err != nil {
		return nil, err
	}

	values := m.Fields()
	if len(values) == 0 {
		return nil, errFieldsNotProvided
	}

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	requestorID := au.SubjectID

	m.SetRequestorID(requestorID)

	v, err := next.Mutate(ctx, m)
	if err != nil {
		return v, err
	}

	id, err := GetObjectIDFromEntValue(v)
	if err != nil {
		return v, err
	}

	_, err = m.Job.Insert(ctx, corejobs.ExportContentArgs{
		ExportID:       id,
		UserID:         requestorID,
		OrganizationID: au.OrganizationID,
	}, nil)

	return v, err
}

func handleExportUpdate(ctx context.Context, m *generated.ExportMutation, next ent.Mutator) (generated.Value, error) {
	fileIDs := objects.GetFileIDsFromContext(ctx)
	if len(fileIDs) > 0 {
		var err error

		ctx, err = checkExportFiles(ctx, m)
		if err != nil {
			return nil, err
		}

		m.AddFileIDs(fileIDs...)
	}

	return next.Mutate(ctx, m)
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
