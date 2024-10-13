package objects

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/objects"
)

// Upload is the object that handles the file upload process
type Upload struct {
	// ObjectStorage is the object storage configuration
	ObjectStorage *objects.Objects
	// Storage is the storage type to use, this can be S3 or Disk
	Storage objects.Storage
}

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
}

// Config defines the config for Mime middleware
type Config struct {
	// Keys is a list of keys to look for in the multipart form
	Keys []string `yaml:"keys"`
	// Skipper defines a function to skip middleware.
	Skipper func(r *http.Request) bool `json:"-" koanf:"-"`
	// Upload is the upload object that handles the file upload process
	Upload *Upload
}

// FileUpload uploads the files to the storage and returns the the context with the uploaded files
func (u *Upload) FileUpload(ctx context.Context, files []FileUpload) (context.Context, error) {
	// set up a wait group to wait for all the uploads to finish
	var wg errgroup.Group

	uploadedFiles := []objects.File{}

	wg.Go(func() (err error) {
		uploadedFiles, err = u.upload(ctx, files)
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
	ctx = objects.WriteFilesToContext(ctx, objects.Files{"upload": uploadedFiles})

	// return the response
	return ctx, nil
}

// FileUploadMiddleware is a middleware that handles the file upload process
// this can be added to the middleware chain to handle file uploads prior to the main handler
// Since gqlgen handles file uploads differently, this middleware is not used in the graphql handler
func FileUploadMiddleware(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper != nil && config.Skipper(r) {
				next.ServeHTTP(w, r)

				return
			}

			ctx, err := config.Upload.multiformParseForm(w, r, config.Keys...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// multiformParseForm parses the multipart form and uploads the files to the storage and returns the context with the uploaded files
func (u *Upload) multiformParseForm(w http.ResponseWriter, r *http.Request, keys ...string) (context.Context, error) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, u.ObjectStorage.MaxSize)

	// skip if the content type is not multipart
	if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		return ctx, nil
	}

	if err := r.ParseMultipartForm(u.ObjectStorage.MaxSize); err != nil {
		u.ObjectStorage.ErrorResponseHandler(err).ServeHTTP(w, r)

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

// upload handles the file upload process per key in the multipart form and returns the uploaded files
func (u *Upload) upload(ctx context.Context, files []FileUpload) ([]objects.File, error) {
	uploadedFiles := make([]objects.File, 0, len(files))

	for _, f := range files {
		// create the file in the database
		entFile, err := u.createFile(ctx, f)
		if err != nil {
			log.Error().Err(err).Str("file", f.Filename).Msg("failed to create file")

			return nil, err
		}

		// generate the uploaded file name
		uploadedFileName := u.ObjectStorage.NameFuncGenerator(entFile.ID + "_" + f.Filename)
		fileData := objects.File{
			ID:               entFile.ID,
			FieldName:        f.Key,
			OriginalName:     f.Filename,
			UploadedFileName: uploadedFileName,
			MimeType:         entFile.DetectedMimeType,
		}

		// validate the file
		if err := u.ObjectStorage.ValidationFunc(fileData); err != nil {
			log.Error().Err(err).Str("file", f.Filename).Msg("failed to validate file")

			return nil, err
		}

		// Upload the file to the storage and get the metadata
		metadata, err := u.Storage.Upload(ctx, f.File, &objects.UploadFileOptions{
			FileName:    uploadedFileName,
			ContentType: entFile.DetectedContentType,
			Metadata: map[string]string{
				"file_id": entFile.ID,
			},
		})
		if err != nil {
			log.Error().Err(err).Str("file", f.Filename).Msg("failed to upload file")

			return nil, err
		}

		// add metadata to file information
		fileData.Size = metadata.Size
		fileData.FolderDestination = metadata.FolderDestination
		fileData.StorageKey = metadata.Key

		// generate a presigned URL that is valid for 15 minutes
		fileData.PresignedURL, err = u.Storage.GetPresignedURL(ctx, uploadedFileName)
		if err != nil {
			log.Error().Err(err).Str("file", f.Filename).Msg("failed to get presigned URL")

			return nil, err
		}

		// update the file with the size
		if _, err := txClientFromContext(ctx).
			UpdateOne(entFile).
			SetPersistedFileSize(metadata.Size).
			SetURI(createURI(entFile.StorageScheme, metadata.FolderDestination, metadata.Key)).
			SetStorageVolume(metadata.FolderDestination).
			SetStoragePath(metadata.Key).
			Save(ctx); err != nil {
			log.Error().Err(err).Msg("failed to update file with size")
			return nil, err
		}

		log.Info().Str("file", fileData.UploadedFileName).
			Str("id", fileData.FolderDestination).
			Str("mime_type", fileData.MimeType).
			Str("size", formatFileSize(fileData.Size)).
			Str("presigned_url", fileData.PresignedURL).
			Msg("file uploaded")

		uploadedFiles = append(uploadedFiles, fileData)
	}

	return uploadedFiles, nil
}

// getFileHeaders returns the file headers for a given key in the multipart form
func (u *Upload) getFileHeaders(r *http.Request, key string) ([]*multipart.FileHeader, error) {
	fileHeaders, ok := r.MultipartForm.File[key]
	if !ok {
		if u.ObjectStorage.IgnoreNonExistentKeys {
			return nil, nil
		}

		return nil, errors.New("file key not found") // nolint:goerr113
	}

	return fileHeaders, nil
}

// formatFileSize converts a file size in bytes to a human-readable string in MB/GB notation.
func formatFileSize(size int64) string {
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

// createURI creates a URI for the file
func createURI(scheme, destination, key string) string {
	return fmt.Sprintf("%s%s/%s", scheme, destination, key)
}

// createFile creates a file in the database and returns the file object
func (u *Upload) createFile(ctx context.Context, f FileUpload) (*ent.File, error) {
	contentType, err := objects.DetectContentType(f.File)
	if err != nil {
		log.Error().Err(err).Str("file", f.Filename).Msg("failed to fetch content type")

		return nil, err
	}

	md5Hash, err := objects.ComputeChecksum(f.File)
	if err != nil {
		log.Error().Err(err).Str("file", f.Filename).Msg("failed to compute checksum")

		return nil, err
	}

	set := ent.CreateFileInput{
		ProvidedFileName:      f.Filename,
		ProvidedFileExtension: filepath.Ext(f.Filename),
		ProvidedFileSize:      &f.Size,
		DetectedMimeType:      &f.ContentType,
		DetectedContentType:   contentType,
		Md5Hash:               &md5Hash,
		StoreKey:              &f.Key,
		StorageScheme:         u.Storage.GetScheme(),
	}

	// get file contents to store in the database
	contents, err := objects.StreamToByte(f.File)
	if err != nil {
		log.Error().Err(err).Str("file", f.Filename).Msg("failed to read file contents")

		return nil, err
	}

	entFile, err := txClientFromContext(ctx).Create().
		SetFileContents(contents).
		SetInput(set).
		Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create file")

		return nil, err
	}

	return entFile, nil
}

// txClientFromContext returns the file client from the context if it exists
// used for transactional mutations, if the client does not exist, it will return nil
func txClientFromContext(ctx context.Context) *ent.FileClient {
	client := ent.FromContext(ctx)
	if client != nil {
		return client.File
	}

	tx := transaction.FromContext(ctx)
	if tx != nil {
		return tx.File
	}

	return nil
}
