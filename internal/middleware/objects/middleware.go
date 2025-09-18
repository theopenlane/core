package objects

import (
	"context"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect/sql"

	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
)

const (
	// MIMEDetectionBufferSize defines the buffer size for MIME type detection
	MIMEDetectionBufferSize = 512
)

// CreateFileRecord creates a file in the database and returns the file object
func CreateFileRecord(ctx context.Context, f storage.File) (*ent.File, error) {
	return createFile(ctx, f)
}

// UpdateFileWithStorageMetadata updates a file entity with storage metadata
func UpdateFileWithStorageMetadata(ctx context.Context, entFile *ent.File, fileData storage.File) error {
	// allow the update, permissions are not yet set to allow the update
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// update the file with the complete storage metadata
	update := txFileClientFromContext(ctx).
		UpdateOne(entFile).
		SetPersistedFileSize(fileData.Size).
		SetURI(fileData.FileMetadata.FullURI).
		SetStoragePath(fileData.Key)

	// Store additional metadata
	if len(fileData.Metadata) > 0 {
		// Convert to map[string]any for ent
		metadata := make(map[string]any)
		for k, v := range fileData.Metadata {
			metadata[k] = v
		}
		update = update.SetMetadata(metadata)
	}

	if _, err := update.Save(allowCtx); err != nil {
		log.Error().Err(err).Msg("failed to update file with storage metadata")
		return err
	}

	log.Debug().Str("file", fileData.Name).Str("id", fileData.ID).Str("key", fileData.Key).Int64("size", fileData.Size).Msg("file uploaded")

	return nil
}

// createFile creates a file in the database and returns the file object
func createFile(ctx context.Context, f storage.File) (*ent.File, error) {
	// Detect content type if not already provided
	contentType := f.ContentType
	if contentType == "" {
		if detectedType, err := storage.DetectContentType(f.RawFile); err == nil {
			contentType = detectedType
		}
	}

	orgID, err := getOrgOwnerID(ctx, f)
	if err != nil {
		return nil, err
	}

	set := ent.CreateFileInput{
		ProvidedFileName:      f.OriginalName,
		ProvidedFileExtension: filepath.Ext(f.ProvidedExtension),
		ProvidedFileSize:      &f.Size,
		DetectedMimeType:      &f.ContentType,
		DetectedContentType:   contentType,
		StoreKey:              &f.Key,
	}

	if orgID != "" {
		set.OrganizationIDs = []string{orgID}
	}

	// bypass further permissions checks and allow the file to be created
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	entFile, err := txFileClientFromContext(ctx).Create().
		SetInput(set).
		Save(allowCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create file")

		return nil, err
	}

	return entFile, nil
}

// getOrgOwnerID retrieves the organization ID from the context or input
func getOrgOwnerID(ctx context.Context, f storage.File) (string, error) {
	// skip if the file is a user file, they will not have an organization ID
	// as the owner and can be used across organizations
	if strings.EqualFold(f.CorrelatedObjectType, "user") {
		return "", nil
	}

	// get the organization ID from the context, ignore the error if it is not set
	// and instead check the parent object for the owner ID
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)

	if orgID == "" {
		// check the parent for the owner_id
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
