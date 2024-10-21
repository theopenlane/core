package objects

import (
	"context"
)

// FileContextKey is the context key for the files
// This is the key that is used to store the files in the context, which is then used to retrieve the files
// in subsequent parts of the request
// this is different than the `key` in the multipart form, which is the form field name that the file was uploaded with
var FileContextKey = &ContextKey{"files"}

// ContextKey is the key name for the additional context
type ContextKey struct {
	name string
}

// WriteFilesToContext retrieves any existing files from the context, appends the new files to the existing files map
// based on the form field name, then returns a new context with the updated files map stored in it
func WriteFilesToContext(ctx context.Context, f Files) context.Context {
	existingFiles, ok := ctx.Value(FileContextKey).(Files)
	if !ok {
		existingFiles = Files{}
	}

	for _, v := range f {
		// all the files should have the same form field so safe to use any index
		existingFiles[v[0].FieldName] = append(existingFiles[v[0].FieldName], v...)
	}

	return context.WithValue(ctx, FileContextKey, existingFiles)
}

// UpdateFileInContextByKey updates the file in the context based on the key and the file ID
func UpdateFileInContextByKey(ctx context.Context, key string, f File) context.Context {
	existingFiles, ok := ctx.Value(FileContextKey).(Files)
	if !ok {
		existingFiles = Files{}
	}

	for i, v := range existingFiles[key] {
		if v.ID == f.ID {
			// update the file in the existing files
			existingFiles[key][i] = f
		}
	}

	return context.WithValue(ctx, FileContextKey, existingFiles)
}

// RemoveFileFromContext removes the file from the context based on the file ID
func RemoveFileFromContext(ctx context.Context, f File) context.Context {
	existingFiles, ok := ctx.Value(FileContextKey).(Files)
	if !ok {
		return ctx
	}

	for key, files := range existingFiles {
		for i, v := range files {
			if v.ID == f.ID {
				existingFiles[key] = append(existingFiles[key][:i], existingFiles[key][i+1:]...)

				// if there are no files left in the key, remove the key from the map
				if len(existingFiles[key]) == 0 {
					delete(existingFiles, key)
				}
			}
		}
	}

	return context.WithValue(ctx, FileContextKey, existingFiles)
}

// FilesFromContext returns all files that have been uploaded during the request
func FilesFromContext(ctx context.Context) (Files, error) {
	files, ok := ctx.Value(FileContextKey).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files, nil
}

// FilesFromContextWithKey returns all files that have been uploaded during the request
// and sorts by the provided form field
func FilesFromContextWithKey(ctx context.Context, key string) ([]File, error) {
	files, ok := ctx.Value(FileContextKey).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files[key], nil
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
