package store

import (
	"context"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect/sql"

	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
)

// CreateFileRecord creates a file record in the database and returns the resulting ent.File entity.
func CreateFileRecord(ctx context.Context, f pkgobjects.File) (*ent.File, error) {
	return createFile(ctx, f)
}

// UpdateFileWithStorageMetadata updates a file entity with metadata returned from the storage provider.
func UpdateFileWithStorageMetadata(ctx context.Context, entFile *ent.File, fileData pkgobjects.File) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	update := txFileClientFromContext(ctx).
		UpdateOne(entFile).
		SetPersistedFileSize(fileData.Size).
		SetURI(fileData.FileMetadata.FullURI).
		SetStoragePath(fileData.Key).
		SetStorageVolume(fileData.Bucket).
		SetStorageProvider(string(fileData.ProviderType))

	if len(fileData.Metadata) > 0 {
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

	return nil
}

func createFile(ctx context.Context, f pkgobjects.File) (*ent.File, error) {
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
		StorageProvider:       lo.ToPtr(string(f.ProviderType)),
		StorageVolume:         &f.Bucket,
		StoragePath:           &f.Folder,
	}

	if orgID != "" {
		set.OrganizationIDs = []string{orgID}
	}

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

func getOrgOwnerID(ctx context.Context, f pkgobjects.File) (string, error) {
	if strings.EqualFold(f.CorrelatedObjectType, "user") {
		return "", nil
	}

	// If the actor is a system admin, prefer deriving the organization from the
	// correlated object rather than using the admin's org from context
	isAdmin := false
	if au, err := auth.GetAuthenticatedUserFromContext(ctx); err == nil && au.IsSystemAdmin {
		isAdmin = true
	}

	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	if isAdmin {
		orgID = ""
	}

	if orgID == "" {
		var rows sql.Rows

		// derive table name from correlated object type using snake_case to match DB naming
		objectType := lo.SnakeCase(f.CorrelatedObjectType)
		objectTable := pluralize.NewClient().Plural(objectType)

		// For schemas that do not include owner_id (e.g., trust_center_docs), resolve owner through the parent
		if objectType == "trust_center_doc" {
			// first get the trust_center_id from the doc
			var tcRows sql.Rows

			tcQuery := "SELECT trust_center_id FROM trust_center_docs WHERE id = $1"

			if err := txClientFromContext(ctx).Driver().Query(ctx, tcQuery, []any{f.CorrelatedObjectID}, &tcRows); err != nil {
				return "", err
			}
			if tcRows.Err() != nil {
				_ = tcRows.Close()
				return "", tcRows.Err()
			}

			var trustCenterID string

			if tcRows.Next() {
				if err := tcRows.Scan(&trustCenterID); err != nil {
					_ = tcRows.Close()
					return "", err
				}
			}

			_ = tcRows.Close()
			if trustCenterID == "" {
				return "", ErrMissingOrganizationID
			}

			// now resolve owner_id from trust_centers
			var ownerRows sql.Rows
			ownerQuery := "SELECT owner_id FROM trust_centers WHERE id = $1"
			if err := txClientFromContext(ctx).Driver().Query(ctx, ownerQuery, []any{trustCenterID}, &ownerRows); err != nil {
				return "", err
			}
			if ownerRows.Err() != nil {
				_ = ownerRows.Close()
				return "", ownerRows.Err()
			}
			if ownerRows.Next() {
				var ownerID string
				if err := ownerRows.Scan(&ownerID); err != nil {
					_ = ownerRows.Close()
					return "", err
				}
				orgID = ownerID
			}
			_ = ownerRows.Close()
		} else {
			// Generic case: attempt to read owner_id directly from the correlated table
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
	}

	if orgID == "" {
		return "", ErrMissingOrganizationID
	}

	return orgID, nil
}

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
