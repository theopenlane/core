package objects

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"slices"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/pkg/objects/storage"
)

// FileSource represents any source that can provide file uploads.
// The tilde (~) allows for types that are identical or aliases to the specified types.
type FileSource interface {
	~map[string]any | ~*http.Request | ~*multipart.Form
}

// ParseFilesFromSource extracts files from any source using generics
func ParseFilesFromSource[T FileSource](source T, keys ...string) (map[string][]storage.File, error) {
	result := make(map[string][]storage.File)
	// Type switch on any(source) is required because Go does not allow type switches directly on generic type parameters
	switch s := any(source).(type) {
	case map[string]any:
		return parseVariablesMap(s, keys...)
	case *http.Request:
		if s.MultipartForm == nil {
			return result, nil
		}
		return parseMultipartForm(s.MultipartForm, keys...)
	case *multipart.Form:
		return parseMultipartForm(s, keys...)
	default:
		log.Warn().Any("type", reflect.TypeOf(source)).Msg("unsupported file source type")
		return result, nil
	}
}

// parseVariablesMap extracts files from a variables map (e.g., GraphQL variables)
func parseVariablesMap(variables map[string]any, keys ...string) (map[string][]storage.File, error) {
	result := make(map[string][]storage.File)

	for key, value := range variables {
		// Skip if this key isn't in our filter list (if provided)
		if len(keys) > 0 {
			found := slices.Contains(keys, key)
			if !found {
				continue
			}
		}

		uploads := extractUploads(value)
		if len(uploads) > 0 {
			var files []storage.File
			for _, upload := range uploads {
				files = append(files, storage.File{
					RawFile:      upload.File,
					OriginalName: upload.Filename,
					FileMetadata: storage.FileMetadata{
						Size:        upload.Size,
						ContentType: upload.ContentType,
						Key:         key,
					},
				})
			}
			result[key] = files
		}
	}

	return result, nil
}

// extractUploads extracts graphql.Upload from any value using type switches
func extractUploads(v any) []graphql.Upload {
	switch val := v.(type) {
	case []graphql.Upload:
		return val
	case graphql.Upload:
		return []graphql.Upload{val}
	case []any:
		var uploads []graphql.Upload
		for _, item := range val {
			if upload, ok := item.(graphql.Upload); ok {
				uploads = append(uploads, upload)
			}
		}
		return uploads
	case map[string]any:
		var uploads []graphql.Upload
		for _, value := range val {
			uploads = append(uploads, extractUploads(value)...)
		}
		return uploads
	default:
		return nil
	}
}

// ProcessFilesForMutation is a generic helper for ent hooks that:
// 1. Gets files from context using the provided key
// 2. Updates files with parent information from mutation
// 3. Updates context with modified files
// This replaces the pattern of individual checkXXXFile functions
func ProcessFilesForMutation[T Mutation](ctx context.Context, mutation T, key string, parentType ...string) (context.Context, error) {
	// Get files from context using the provided key
	files, err := FilesFromContextWithKey(ctx, key)
	if err != nil {
		return ctx, err
	}

	// Return early if no files
	if len(files) == 0 {
		return ctx, nil
	}

	// Get mutation ID and type
	mutationID, err := mutation.ID()
	if err != nil {
		return ctx, err
	}

	mutationType := mutation.Type()
	if len(parentType) > 0 {
		mutationType = parentType[0] // Allow override of parent type
	}

	// Update all files with parent information
	for i := range files {
		files[i].Parent.ID = mutationID
		files[i].Parent.Type = mutationType

		// Update each file in context
		ctx = UpdateFileInContextByKey(ctx, key, files[i])
	}

	return ctx, nil
}

// parseMultipartForm extracts files from multipart.Form
func parseMultipartForm(form *multipart.Form, keys ...string) (map[string][]storage.File, error) {
	result := make(map[string][]storage.File)

	// If no keys specified, use all available keys
	if len(keys) == 0 {
		for key := range form.File {
			keys = append(keys, key)
		}
	}

	for _, key := range keys {
		fileHeaders, exists := form.File[key]
		if !exists {
			continue
		}

		var files []storage.File
		for _, header := range fileHeaders {
			file, err := header.Open()
			if err != nil {
				log.Error().Err(err).Str("file", header.Filename).Msg("failed to open file")

				continue
			}

			files = append(files, storage.File{
				RawFile:      file,
				OriginalName: header.Filename,
				FileMetadata: storage.FileMetadata{
					Size:        header.Size,
					ContentType: header.Header.Get("Content-Type"),
					Key:         key,
				},
			})
		}

		if len(files) > 0 {
			result[key] = files
		}
	}

	return result, nil
}

// WriteFilesToContext retrieves any existing files from the context, appends the new files to the existing files map
// based on the form field name, then returns a new context with the updated files map stored in it
func WriteFilesToContext(ctx context.Context, f storage.Files) context.Context {
	files, ok := contextx.From[storage.Files](ctx)
	if !ok {
		files = storage.Files{}
	}

	for _, v := range f {
		for _, fileObj := range v {
			files[fileObj.Key] = append(files[fileObj.Key], fileObj)
		}
	}

	return contextx.With(ctx, files)
}

// UpdateFileInContextByKey updates the file in the context based on the key and the file ID
func UpdateFileInContextByKey(ctx context.Context, key string, f storage.File) context.Context {
	files, ok := contextx.From[storage.Files](ctx)
	if !ok {
		files = storage.Files{}
	}

	for i, v := range files[key] {
		if v.ID == f.ID {
			files[key][i] = f
		}
	}

	return contextx.With(ctx, files)
}

// RemoveFileFromContext removes the file from the context based on the file ID
func RemoveFileFromContext(ctx context.Context, f storage.File) context.Context {
	files, ok := contextx.From[storage.Files](ctx)
	if !ok {
		files = storage.Files{}
	}

	for key, fileList := range files {
		filteredFiles := lo.Filter(fileList, func(file storage.File, _ int) bool {
			return file.ID != f.ID
		})

		if len(filteredFiles) == 0 {
			delete(files, key)
		} else {
			files[key] = filteredFiles
		}
	}

	return contextx.With(ctx, files)
}

// FilesFromContext returns all files that have been uploaded during the request
func FilesFromContext(ctx context.Context) (storage.Files, error) {
	files, ok := contextx.From[storage.Files](ctx)
	if !ok || files == nil {
		return nil, storage.ErrNoFilesUploaded
	}

	return files, nil
}

// FilesFromContextWithKey returns all files that have been uploaded during the request
// and sorts by the provided form field
func FilesFromContextWithKey(ctx context.Context, key string) ([]storage.File, error) {
	files, ok := contextx.From[storage.Files](ctx)
	if !ok || files == nil {
		return nil, storage.ErrNoFilesUploaded
	}

	return files[key], nil
}

// GetFileIDsFromContext returns the file IDs from the context that are associated with the request
func GetFileIDsFromContext(ctx context.Context) []string {
	files, _ := FilesFromContext(ctx)

	if len(files) == 0 {
		return []string{}
	}

	return lo.FlatMap(lo.Values(files), func(fileList []storage.File, _ int) []string {
		return lo.Map(fileList, func(file storage.File, _ int) string {
			return file.ID
		})
	})
}

// ReaderToSeeker function takes an io.Reader as input and returns an io.ReadSeeker which can be used to upload files to the object storage
func ReaderToSeeker(r io.Reader) (io.ReadSeeker, error) {
	if r == nil {
		return nil, nil
	}

	tmpfile, err := os.CreateTemp("", "upload-")
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(tmpfile, r); err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())

		return nil, err
	}

	if _, err = tmpfile.Seek(0, 0); err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())

		return nil, err
	}

	// Return the file, which implements io.ReadSeeker which you can now pass to the objects uploader
	return tmpfile, nil
}
