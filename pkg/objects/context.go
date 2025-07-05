package objects

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// FileContextKey is the context key for the files
// This is the key that is used to store the files in the context, which is then used to retrieve the files
// in subsequent parts of the request
// this is different than the `key` in the multipart form, which is the form field name that the file was uploaded with
type FileContextKey struct {
	Files Files
}

// WriteFilesToContext retrieves any existing files from the context, appends the new files to the existing files map
// based on the form field name, then returns a new context with the updated files map stored in it
func WriteFilesToContext(ctx context.Context, f Files) context.Context {
	fileCtx, ok := contextx.From[FileContextKey](ctx)
	if !ok {
		fileCtx = FileContextKey{Files: Files{}}
	}

	for _, v := range f {
		for _, fileObj := range v {
			fileCtx.Files[fileObj.FieldName] = append(fileCtx.Files[fileObj.FieldName], fileObj)
		}
	}

	return contextx.With(ctx, fileCtx)
}

// UpdateFileInContextByKey updates the file in the context based on the key and the file ID
func UpdateFileInContextByKey(ctx context.Context, key string, f File) context.Context {
	fileCtx, ok := contextx.From[FileContextKey](ctx)
	if !ok {
		fileCtx = FileContextKey{Files: Files{}}
	}

	for i, v := range fileCtx.Files[key] {
		if v.ID == f.ID {
			// update the file in the existing files
			fileCtx.Files[key][i] = f
		}
	}

	return contextx.With(ctx, fileCtx)
}

// RemoveFileFromContext removes the file from the context based on the file ID
func RemoveFileFromContext(ctx context.Context, f File) context.Context {
	fileCtx, ok := contextx.From[FileContextKey](ctx)
	if !ok {
		fileCtx = FileContextKey{Files: Files{}}
	}

	for key, files := range fileCtx.Files {
		for i, v := range files {
			if v.ID == f.ID {
				fileCtx.Files[key] = append(fileCtx.Files[key][:i], fileCtx.Files[key][i+1:]...)

				// if there are no files left in the key, remove the key from the map
				if len(fileCtx.Files[key]) == 0 {
					delete(fileCtx.Files, key)
				}
			}
		}
	}

	return contextx.With(ctx, fileCtx)
}

// FilesFromContext returns all files that have been uploaded during the request
func FilesFromContext(ctx context.Context) (Files, error) {
	fileCtx, ok := contextx.From[FileContextKey](ctx)
	if !ok || fileCtx.Files == nil {
		return nil, ErrNoFilesUploaded
	}

	return fileCtx.Files, nil
}

// FilesFromContextWithKey returns all files that have been uploaded during the request
// and sorts by the provided form field
func FilesFromContextWithKey(ctx context.Context, key string) ([]File, error) {
	fileCtx, ok := contextx.From[FileContextKey](ctx)
	if !ok || fileCtx.Files == nil {
		return nil, ErrNoFilesUploaded
	}

	return fileCtx.Files[key], nil
}

// GetFileIDsFromContext returns the file IDs from the context that are associated with the request
func GetFileIDsFromContext(ctx context.Context) []string {
	// ignore the error, if the files are not found in the context, we just skip the file processing
	files, _ := FilesFromContext(ctx) //nolint:errcheck

	if len(files) == 0 {
		return []string{}
	}

	fileIDs := make([]string, 0, len(files))

	for _, fileKeys := range files {
		for _, file := range fileKeys {
			fileIDs = append(fileIDs, file.ID)
		}
	}

	return fileIDs
}
