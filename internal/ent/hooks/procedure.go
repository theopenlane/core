package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/microcosm-cc/bluemonday"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

type importSchemaMutation interface {
	SetName(string)
	SetDetails(string)
}

// importedFileToSchemaImport handles the common logic for processing uploaded files
// and setting the name and details from the file content
func importedFileToSchema(ctx context.Context, m importSchemaMutation, objectManager *objects.Objects, fileKey string) error {
	file, _ := objects.FilesFromContextWithKey(ctx, fileKey)

	if len(file) == 0 {
		return nil
	}

	content, err := objectManager.Storage.Download(ctx, &objects.DownloadFileOptions{
		FileName: file[0].UploadedFileName,
	})
	if err != nil {
		return err
	}

	parsedContent, err := objects.ParseDocument(content.File, file[0].MimeType)
	if err != nil {
		return err
	}

	p := bluemonday.UGCPolicy()

	m.SetDetails(p.Sanitize(parsedContent))
	m.SetName(file[0].OriginalName)

	return nil
}

// HookProcedure checks to see if we have an uploaded file.
// If we do, use that as the details of the procedure. and also
// use the name of the file
func HookProcedure() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ProcedureFunc(func(ctx context.Context, m *generated.ProcedureMutation) (generated.Value, error) {

			fileIDs := objects.GetFileIDsFromContext(ctx)

			// no uploaded file. continue as expected
			if len(fileIDs) == 0 {
				return next.Mutate(ctx, m)
			}

			ctx, err := checkProcedureFile(ctx, m)
			if err != nil {
				return nil, err
			}

			m.AddFileIDs(fileIDs...)

			if err := importedFileToSchema(ctx, m, m.ObjectManager, "procedureFile"); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

func checkProcedureFile[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "procedureFile"

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

	// this should always be true, but check just in case
	if file[0].FieldName == key {

		file[0].Parent.ID, _ = m.ID()
		file[0].Parent.Type = m.Type()

		ctx = objects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
}
