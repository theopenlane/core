package objects

import (
	"context"
	"path/filepath"

	"entgo.io/ent/dialect/sql"

	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
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

		// allow the update, permissions are not yet set to allow the update
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

		// update the file with the size
		if _, err := txFileClientFromContext(ctx).
			UpdateOne(entFile).
			SetPersistedFileSize(metadata.Size).
			SetURI(objects.CreateURI(entFile.StorageScheme, metadata.FolderDestination, metadata.Key)).
			SetStorageVolume(metadata.FolderDestination).
			SetStoragePath(metadata.Key).
			Save(allowCtx); err != nil {
			log.Error().Err(err).Msg("failed to update file with size")
			return nil, err
		}

		log.Debug().Str("file", fileData.UploadedFileName).
			Str("id", fileData.FolderDestination).
			Str("mime_type", fileData.MimeType).
			Str("size", objects.FormatFileSize(fileData.Size)).
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

	orgID, err := getOrgOwnerID(ctx, f)
	if err != nil {
		return nil, err
	}

	set := ent.CreateFileInput{
		ProvidedFileName:      f.Filename,
		ProvidedFileExtension: filepath.Ext(f.Filename),
		ProvidedFileSize:      &f.Size,
		DetectedMimeType:      &f.ContentType,
		DetectedContentType:   contentType,
		StoreKey:              &f.Key,
		StorageScheme:         u.Storage.GetScheme(),
	}

	if orgID != "" {
		set.OrganizationIDs = []string{orgID}
	}

	// get file contents to store in the database
	contents, err := objects.StreamToByte(f.File)
	if err != nil {
		log.Error().Err(err).Str("file", f.Filename).Msg("failed to read file contents")

		return nil, err
	}

	entFile, err := txFileClientFromContext(ctx).Create().
		SetFileContents(contents).
		SetInput(set).
		Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create file")

		return nil, err
	}

	return entFile, nil
}

// getOrgOwnerID retrieves the organization ID from the context or input
func getOrgOwnerID(ctx context.Context, f objects.FileUpload) (string, error) {
	// get the organization ID from the context, ignore the error if it is not set
	// and instead check the parent object for the owner ID
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)

	if orgID == "" {
		// check the parent
		var rows sql.Rows
		objectTable := pluralize.NewClient().Plural(f.CorrelatedObjectType)
		query := "SELECT owner_id FROM " + objectTable + " WHERE id = $1"
		if err := txClientFromContext(ctx).Driver().Query(ctx, query, []any{f.CorrelatedObjectID}, &rows); err != nil {
			return "", err
		}

		if rows.Err() != nil {
			return "", rows.Err()
		}

		defer rows.Close()

		if rows.Next() {
			var ownerID string
			if err := rows.Scan(&ownerID); err != nil {
				return "", err
			}

			orgID = ownerID
		}
	}

	return orgID, nil
}

// txFileClientFromContext returns the file client from the context if it exists
// used for transactional mutations, if the client does not exist, it will return nil
func txFileClientFromContext(ctx context.Context) *ent.FileClient {
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

// txFileClientFromContext returns the file client from the context if it exists
// used for transactional mutations, if the client does not exist, it will return nil
func txClientFromContext(ctx context.Context) *ent.Client {
	client := ent.FromContext(ctx)
	if client != nil {
		return client
	}

	tx := transaction.FromContext(ctx)
	if tx != nil {
		return tx.Client()
	}

	return nil
}
