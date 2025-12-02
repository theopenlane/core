package objects

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"slices"
	"sync"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/contextx"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/shared/objects/storage"
)

// FileContextKey is the context key for the files
// This is the key that is used to store the files in the context, which is then used to retrieve the files
// in subsequent parts of the request
// this is different than the `key` in the multipart form, which is the form field name that the file was uploaded with
type FileContextKey struct {
	Files Files
}

// FileSource represents any source that can provide file uploads.
// The tilde (~) allows for types that are identical or aliases to the specified types.
type FileSource interface {
	~map[string]any | ~*http.Request | ~*multipart.Form
}

// ParseFilesFromSource extracts files from any source using generics
func ParseFilesFromSource[T FileSource](source T, keys ...string) (map[string][]File, error) {
	result := make(map[string][]File)
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
func parseVariablesMap(variables map[string]any, keys ...string) (map[string][]File, error) {
	result := make(map[string][]File)

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
			var files []File
			for _, upload := range uploads {
				files = append(files, File{
					RawFile:      upload.File,
					OriginalName: upload.Filename,
					FieldName:    key,
					FileMetadata: FileMetadata{
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
	files, _ := FilesFromContextWithKey(ctx, key)

	// Return early if no files
	if files == nil {
		return ctx, nil
	}

	mutationType := mutation.Type()
	if len(parentType) > 0 {
		mutationType = parentType[0] // Allow override of parent type
	}

	// Set the parent ID and type for the file(s)
	for i, f := range files {
		// this should always be true, but check just in case
		if f.FieldName == key {
			id, _ := mutation.ID()
			// Set parent information used by tuple writer
			files[i].Parent.ID = id
			files[i].Parent.Type = mutationType
			// Also set correlated object information used by persistence to derive org owner
			files[i].CorrelatedObjectID = id
			files[i].CorrelatedObjectType = mutation.Type()

			ctx = UpdateFileInContextByKey(ctx, key, files[i])
		}
	}

	return ctx, nil
}

// parseMultipartForm extracts files from multipart.Form
func parseMultipartForm(form *multipart.Form, keys ...string) (map[string][]File, error) {
	result := make(map[string][]File)
	var mu sync.Mutex
	var eg errgroup.Group

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

		eg.Go(func() error {
			var files []File
			for _, header := range fileHeaders {
				file, err := header.Open()
				if err != nil {
					log.Error().Err(err).Str("file", header.Filename).Msg("failed to open file")
					// TODO: update behavior + wrap AroundResponses which can put failed files
					// into a context and the AroundResponses which checks for it and adds errors to return
					return err
				}

				defer file.Close()

				files = append(files, File{
					RawFile:      file,
					OriginalName: header.Filename,
					FieldName:    key,
					FileMetadata: FileMetadata{
						Size:        header.Size,
						ContentType: header.Header.Get("Content-Type"),
						Key:         key,
					},
				})
			}

			if len(files) > 0 {
				mu.Lock()
				result[key] = files
				mu.Unlock()
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return result, err
	}

	return result, nil
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

	for key, fileList := range fileCtx.Files {
		filteredFiles := lo.Filter(fileList, func(file File, _ int) bool {
			return file.ID != f.ID
		})

		if len(filteredFiles) == 0 {
			delete(fileCtx.Files, key)
		} else {
			fileCtx.Files[key] = filteredFiles
		}
	}

	return contextx.With(ctx, fileCtx)
}

// FilesFromContext returns all files that have been uploaded during the request
func FilesFromContext(ctx context.Context) (Files, error) {
	fileCtx, ok := contextx.From[FileContextKey](ctx)
	if !ok || fileCtx.Files == nil {
		return nil, storage.ErrNoFilesUploaded
	}

	return fileCtx.Files, nil
}

// FilesFromContextWithKey returns all files that have been uploaded during the request
// and sorts by the provided form field
func FilesFromContextWithKey(ctx context.Context, key string) ([]File, error) {
	fileCtx, ok := contextx.From[FileContextKey](ctx)
	if !ok || fileCtx.Files == nil {
		return nil, storage.ErrNoFilesUploaded
	}

	return fileCtx.Files[key], nil
}

// GetFileIDsFromContext returns the file IDs from the context that are associated with the request
func GetFileIDsFromContext(ctx context.Context) []string {
	files, _ := FilesFromContext(ctx)

	if len(files) == 0 {
		return []string{}
	}

	return lo.FlatMap(lo.Values(files), func(fileList []File, _ int) []string {
		return lo.Map(fileList, func(file File, _ int) string {
			return file.ID
		})
	})
}

// ReaderToSeeker function takes an io.Reader as input and returns an io.ReadSeeker which can be used to upload files to the object storage
// If the reader is already a ReadSeeker (e.g., BufferedReader from injectFileUploader), it returns it directly.
// For files under MaxInMemorySize (10MB), it uses in-memory buffering for efficiency.
// For larger files, it falls back to temporary file storage.
func ReaderToSeeker(r io.Reader) (io.ReadSeeker, error) {
	if r == nil {
		return nil, nil
	}

	// If already a ReadSeeker, return it directly
	if seeker, ok := r.(io.ReadSeeker); ok {
		return seeker, nil
	}

	// Try to use in-memory buffering for small files
	br, err := NewBufferedReaderFromReader(r)
	if err == nil {
		return br, nil
	}

	// If file is too large or buffering fails, fall back to temp file
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
