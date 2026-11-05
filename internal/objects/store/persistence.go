package store

import (
	"context"
	"encoding/json"
	"maps"
	"path/filepath"
	"strings"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/consts"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/middleware/transaction"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/logx"
)

// CreateFileRecord creates a file record in the database and returns the resulting ent.File entity.
func CreateFileRecord(ctx context.Context, f pkgobjects.File) (*ent.File, error) {
	return createFile(ctx, f)
}

// UpdateFileWithStorageMetadata updates a file entity with metadata returned from the storage provider.
func UpdateFileWithStorageMetadata(ctx context.Context, entFile *ent.File, fileData pkgobjects.File) error {
	internalCtx := auth.WithInternalOperationContext(ctx)

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
		maps.Copy(metadata, fileData.Metadata)

		update = update.SetMetadata(metadata)
	}

	if _, err := update.Save(internalCtx); err != nil {
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
		Name:                  new(retrieveFileName(f)),
		ProvidedFileName:      f.OriginalName,
		ProvidedFileExtension: resolveProvidedExtension(f),
		ProvidedFileSize:      &f.Size,
		DetectedMimeType:      &f.ContentType,
		DetectedContentType:   contentType,
		StoreKey:              &f.Key,
		StorageProvider:       new(string(f.ProviderType)),
		StorageVolume:         &f.Bucket,
		StoragePath:           &f.Folder,
	}

	if len(f.MD5) > 0 {
		set.Md5Hash = new(string(f.MD5))
	}

	if name := getCategoryNameForSchema(f); name != "" {
		set.CategoryName = &name
	}

	if orgID != "" {
		set.OrganizationIDs = []string{orgID}
	}

	internalCtx := auth.WithInternalOperationContext(ctx)
	entFile, err := txFileClientFromContext(ctx).Create().
		SetInput(set).
		Save(internalCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create file")

		return nil, err
	}

	return entFile, nil
}

// resolveProvidedExtension returns the file extension to persist. It prefers an
// explicitly-set ProvidedExtension and falls back to deriving from OriginalName,
// which is the source populated by the multipart parser and handles filenames
// with spaces, multiple dots, or other special characters.
func resolveProvidedExtension(f pkgobjects.File) string {
	if f.ProvidedExtension != "" {
		return f.ProvidedExtension
	}

	return filepath.Ext(f.OriginalName)
}

func retrieveFileName(f pkgobjects.File) string {
	if name := strings.TrimSpace(f.Name); name != "" {
		return name
	}

	if f.Metadata != nil {
		if name, ok := f.Metadata["name"].(string); ok {
			name = strings.TrimSpace(name)
			if name != "" {
				return name
			}
		}
	}

	return f.OriginalName
}

func getCategoryNameForSchema(f pkgobjects.File) string {
	object := strings.TrimSpace(f.CorrelatedObjectType)

	for k, v := range FileCategoryDefaults {
		if strings.EqualFold(k, object) {
			return v
		}
	}

	return ""
}

// mappedParent represents the schema used to derive the organization ID for object types
// that do not have an owner_id field directly and the field in the child schema that references the parent
type mappedParent struct {
	// parentSchema is the schema that contains the owner field, e.g. TrustCenter for trust center docs
	parentSchema string
	// idField is the field in the child schema that references the parent, e.g. for trust_center_docs it would be trust_center_id
	idField string
}

// nonOwnedSchemas is a map of object types that do not have an owner_id field
// to their parent schema and the field in the child that references the parent
// This is used to derive the organization ID for files correlated to these object types
// by loading through the parent
var nonOwnedSchemas = map[string]mappedParent{
	ent.TypeTrustCenterDoc: {
		parentSchema: ent.TypeTrustCenter,
		idField:      "trust_center_id",
	},
}

func getOrgOwnerID(ctx context.Context, f pkgobjects.File) (string, error) {
	if strings.EqualFold(f.CorrelatedObjectType, "user") {
		return "", nil
	}

	// If the actor is a system admin, prefer deriving the organization from the
	// correlated object rather than using the admin's org from context
	persistCaller, persistOk := auth.CallerFromContext(ctx)
	if !persistOk {
		return "", auth.ErrNoAuthUser
	}

	if !persistCaller.Has(auth.CapSystemAdmin) {
		if persistCaller.OrganizationID != "" {
			return persistCaller.OrganizationID, nil
		}

		orgID, ok := persistCaller.ActiveOrg()
		if ok {
			return orgID, nil
		}
	}

	ownerID, err := loadOwnerID(ctx, f.CorrelatedObjectType, f.CorrelatedObjectID)
	if err != nil {
		return "", err
	}

	if ownerID != "" {
		return ownerID, nil
	}

	// use system admin org if the user is a system admin and we got to here
	if persistCaller.Has(auth.CapSystemAdmin) {
		return consts.SystemAdminOrgID, nil
	}

	return "", ErrMissingOrganizationID
}

// loadOwnerID returns the owner of the correlated object, resolving through the parent for schemas without an owner field
func loadOwnerID(ctx context.Context, objectType, objectID string) (string, error) {
	schema, ok := entityops.LookupSchema(objectType)
	if !ok || schema.Load == nil {
		return "", nil
	}

	if schema.OwnerField != "" {
		return loadStringField(ctx, schema, objectID, schema.OwnerField)
	}

	parent, ok := nonOwnedSchemas[schema.Name]
	if !ok {
		return "", nil
	}

	parentID, err := loadStringField(ctx, schema, objectID, parent.idField)
	if err != nil || parentID == "" {
		return "", err
	}

	return loadOwnerID(ctx, parent.parentSchema, parentID)
}

// loadStringField loads the object through the entityops registry and returns the named field, empty when the object does not exist
func loadStringField(ctx context.Context, schema *entityops.Schema, objectID, field string) (string, error) {
	data, err := schema.Load(ctx, txClientFromContext(ctx), objectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}

		return "", err
	}

	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return "", err
	}

	value, _ := fields[field].(string)

	return value, nil
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
