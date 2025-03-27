package objects

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// FileUpload is the object that holds the file information
type FileUpload struct {
	// File is the file to be uploaded
	File io.ReadSeeker
	// Filename is the name of the file provided in the multipart form
	Filename string
	// Size is the size of the file in bytes
	Size int64
	// ContentType is the content type of the file from the header
	ContentType string
	// Key is the field name from the graph input or multipart form
	Key string
	// CorrelatedObjectID is the key of the object that the file is associated with
	CorrelatedObjectID string
	// CorrelatedObjectType is the type of the object that the file is associated with
	CorrelatedObjectType string
}

// FileUpload uploads the files to the storage and returns the the context with the uploaded files
func (u *Objects) FileUpload(ctx context.Context, files []FileUpload) (context.Context, error) {
	// set up a wait group to wait for all the uploads to finish
	var wg errgroup.Group

	uploadedFiles := []File{}

	wg.Go(func() (err error) {
		uploadedFiles, err = u.Uploader(ctx, u, files)
		if err != nil {
			log.Error().Err(err).Msg("failed to upload files")

			return err
		}

		return nil
	})

	// wait for all the uploads to finish
	if err := wg.Wait(); err != nil {
		return ctx, err
	}

	// check if any files were uploaded, if not return early
	if len(uploadedFiles) == 0 {
		return ctx, nil
	}

	// write the uploaded files to the context
	ctx = WriteFilesToContext(ctx, Files{"upload": uploadedFiles})

	// return the response
	return ctx, nil
}

// FileUploadMiddleware is a middleware that handles the file upload process
// this can be added to the middleware chain to handle file uploads prior to the main handler
// Since gqlgen handles file uploads differently, this middleware is not used in the graphql handler
func FileUploadMiddleware(config *Objects) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(r) {
				next.ServeHTTP(w, r)

				return
			}

			ctx, err := config.multiformParseForm(w, r, config.Keys...)
			if err != nil {
				config.ErrorResponseHandler(err, http.StatusBadRequest).ServeHTTP(w, r)

				return
			}

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// multiformParseForm parses the multipart form and uploads the files to the storage and returns the context with the uploaded files
func (u *Objects) multiformParseForm(w http.ResponseWriter, r *http.Request, keys ...string) (context.Context, error) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, u.MaxSize)

	// skip if the content type is not multipart
	if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		return ctx, nil
	}

	if err := r.ParseMultipartForm(u.MaxSize); err != nil {
		return nil, err
	}

	var wg errgroup.Group

	for _, key := range keys {
		wg.Go(func() error {
			fileHeaders, err := u.getFileHeaders(r, key)
			if err != nil {
				// log the error and skip the key
				// do not return an error if the key is not found
				// this is to allow for optional keys
				log.Info().Err(err).Str("key", key).Msg("key not found, skipping")

				return nil
			}

			files, err := parse(fileHeaders, key)
			if err != nil {
				log.Error().Err(err).Str("key", key).Msg("failed to parse files from headers")
			}

			ctx, err = u.FileUpload(ctx, files)
			if err != nil {
				log.Error().Err(err).Str("key", key).Msg("failed to upload files")

				return err
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return ctx, nil
}

// getFileHeaders returns the file headers for a given key in the multipart form
func (u *Objects) getFileHeaders(r *http.Request, key string) ([]*multipart.FileHeader, error) {
	fileHeaders, ok := r.MultipartForm.File[key]
	if !ok {
		if u.IgnoreNonExistentKeys {
			return nil, nil
		}

		return nil, errors.New("file key not found") // nolint:err113
	}

	return fileHeaders, nil
}

// parse handles the parses the multipart form and returns the files to be uploaded
func parse(fileHeaders []*multipart.FileHeader, key string) ([]FileUpload, error) {
	files := []FileUpload{}

	for _, header := range fileHeaders {
		f, err := header.Open()
		if err != nil {
			log.Error().Err(err).Str("file", header.Filename).Msg("failed to open file")
			return nil, err
		}

		defer f.Close()

		fileUpload := FileUpload{
			File:        f,
			Filename:    header.Filename,
			Size:        header.Size,
			ContentType: header.Header.Get("Content-Type"),
			Key:         key,
		}

		files = append(files, fileUpload)
	}

	return files, nil
}

// FormatFileSize converts a file size in bytes to a human-readable string in MB/GB notation.
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

// CreateURI creates a URI for the file based on the scheme, destination and key
func CreateURI(scheme, destination, key string) string {
	return fmt.Sprintf("%s%s/%s", scheme, destination, key)
}
