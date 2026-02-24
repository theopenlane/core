package store

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect/sql"

	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/consts"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
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
		SetStorageRegion(fileData.Region).
		SetStorageProvider(string(fileData.ProviderType))

	if len(fileData.Metadata) > 0 {
		metadata := make(map[string]any)
		for k, v := range fileData.Metadata {
			metadata[k] = v
		}

		update = update.SetMetadata(metadata)
	}

	if _, err := update.Save(allowCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to update file with storage metadata")

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
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create file")

		return nil, err
	}

	return entFile, nil
}

// mappedParent represents the table used to derive the organization ID for object types
// that do not have an owner_id field directly and the field in the child table that references the parent
type mappedParent struct {
	// parentTable is the table that contains the owner_id field, it should be the plural form, e.g. for trust_center it would be trust_centers
	parentTable string
	// idField is the field in the child table that references the parent table, e.g. for trust_center_docs it would be trust_center_id
	idField string
}

// nonOwnedSchemas is a map of object types that do not have an owner_id field
// to their parent table and the field in that table that references the object
// This is used to derive the organization ID for files correlated to these object types
// by joining through the parent table
var nonOwnedSchemas = map[string]mappedParent{
	"trust_center_doc": {
		parentTable: "trust_centers",
		idField:     "trust_center_id",
	},
}

func getOrgOwnerID(ctx context.Context, f pkgobjects.File) (string, error) {
	if strings.EqualFold(f.CorrelatedObjectType, "user") {
		return "", nil
	}

	// If the actor is a system admin, prefer deriving the organization from the
	// correlated object rather than using the admin's org from context
	persistCaller, persistOk := auth.CallerFromContext(ctx)
	if !persistOk || persistCaller == nil {
		return "", ErrMissingOrganizationID
	}

	if !persistCaller.Has(auth.CapSystemAdmin) {
		if persistCaller.OrganizationID != "" {
			return persistCaller.OrganizationID, nil
		}

		orgIDs := persistCaller.OrgIDs()
		if len(orgIDs) == 1 {
			return orgIDs[0], nil
		}
	}

	// derive table name from correlated object type using snake_case to match DB naming
	objectType := lo.SnakeCase(f.CorrelatedObjectType)
	objectTable := pluralize.NewClient().Plural(objectType)

	var rows sql.Rows

	// For schemas that do not include owner_id (e.g., trust_center_docs), resolve owner through the parent
	if mappedParent, ok := nonOwnedSchemas[objectType]; ok {
		tcQuery := fmt.Sprintf("SELECT a.owner_id FROM %s as a JOIN %s as b ON a.id = b.%s WHERE b.id = $1", mappedParent.parentTable, objectTable, mappedParent.idField)

		if err := txClientFromContext(ctx).Driver().Query(ctx, tcQuery, []any{f.CorrelatedObjectID}, &rows); err != nil {
			return "", err
		}
	} else {
		// Generic case: attempt to read owner_id directly from the correlated table
		query := "SELECT owner_id FROM " + objectTable + " WHERE id = $1"
		if err := txClientFromContext(ctx).Driver().Query(ctx, query, []any{f.CorrelatedObjectID}, &rows); err != nil {
			return "", err
		}
	}

	defer rows.Close()

	if rows.Err() != nil {
		return "", rows.Err()
	}

	if rows.Next() {
		var ownerID sql.NullString
		if err := rows.Scan(&ownerID); err != nil {
			return "", err
		}

		if ownerID.Valid {
			return ownerID.String, nil
		}
	}

	// use system admin org if the user is a system admin and we got to here
	if persistCaller.Has(auth.CapSystemAdmin) {
		return consts.SystemAdminOrgID, nil
	}

	return "", ErrMissingOrganizationID
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
