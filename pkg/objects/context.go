package objects

import (
	"context"
	"net/http"
)

type contextKey string

const (
	fileKey contextKey = "files"
)

// WriteFilesToContext retrieves any existing files from the context, appends the new files to the existing files map
// based on the form field name, then returns a new context with the updated files map stored in it
func WriteFilesToContext(ctx context.Context, f Files) context.Context {
	existingFiles, ok := ctx.Value(fileKey).(Files)
	if !ok {
		existingFiles = Files{}
	}

	for _, v := range f {
		// all the files should have the same form field so safe to use any index
		existingFiles[v[0].FieldName] = append(existingFiles[v[0].FieldName], v...)
	}

	return context.WithValue(ctx, fileKey, existingFiles)
}

// FilesFromContext returns all files that have been uploaded during the request
func FilesFromContext(r *http.Request) (Files, error) {
	files, ok := r.Context().Value(fileKey).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files, nil
}

// FilesFromContextWithKey returns  all files that have been uploaded during the request
// and sorts by the provided form field
func FilesFromContextWithKey(r *http.Request, key string) ([]File, error) {
	files, ok := r.Context().Value(fileKey).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files[key], nil
}
