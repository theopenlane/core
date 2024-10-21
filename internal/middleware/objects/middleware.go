package objects

import (
	"context"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/objects"
)

// Upload handles the file Upload process per key in the multipart form and returns the uploaded files
// in addition to uploading the files to the storage, it also creates the file in the database
func Upload(ctx context.Context, u *objects.Objects, files []objects.FileUpload) ([]objects.File, error) {
	uploadedFiles := make([]objects.File, 0, len(files))

	for _, f := range files {
		// create the file in the database
		entFile, err := createFile(ctx, u, f)
		if err != nil {
			log.Error().Err(err).Str("file", f.Filename).Msg("failed to create file")

			return nil, err
		}

		// generate the uploaded file name
		uploadedFileName := u.NameFuncGenerator(entFile.ID + "_" + f.Filename)
		fileData := objects.File{
			ID:               entFile.ID,
			FieldName:        f.Key,
			OriginalName:     f.Filename,
			UploadedFileName: uploadedFileName,
			MimeType:         entFile.DetectedMimeType,
			ContentType:      entFile.DetectedContentType,
		}

		// validate the file
		if err := u.ValidationFunc(fileData); err != nil {
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
		fileData.PresignedURL, err = u.Storage.GetPresignedURL(ctx, uploadedFileName, 60*time.Minute) // nolint:mnd
		if err != nil {
			log.Error().Err(err).Str("file", f.Filename).Msg("failed to get presigned URL")

			return nil, err
		}

		// update the file with the size
		if _, err := txClientFromContext(ctx).
			UpdateOne(entFile).
			SetPersistedFileSize(metadata.Size).
			SetURI(objects.CreateURI(entFile.StorageScheme, metadata.FolderDestination, metadata.Key)).
			SetStorageVolume(metadata.FolderDestination).
			SetStoragePath(metadata.Key).
			Save(ctx); err != nil {
			log.Error().Err(err).Msg("failed to update file with size")
			return nil, err
		}

		log.Info().Str("file", fileData.UploadedFileName).
			Str("id", fileData.FolderDestination).
			Str("mime_type", fileData.MimeType).
			Str("size", objects.FormatFileSize(fileData.Size)).
			Str("presigned_url", fileData.PresignedURL).
			Msg("file uploaded")

		uploadedFiles = append(uploadedFiles, fileData)
	}

	return uploadedFiles, nil
}

// createFile creates a file in the database and returns the file object
func createFile(ctx context.Context, u *objects.Objects, f objects.FileUpload) (*ent.File, error) {
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
