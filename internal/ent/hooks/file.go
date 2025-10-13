package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

var (
	errInvalidStoragePath = errors.New("invalid path when deleting file from object storage")
)

// HookFileDelete makes sure to clean up the file from external storage once deleted
func HookFileDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.FileFunc(
			func(ctx context.Context, m *generated.FileMutation) (generated.Value, error) {

				if m.ObjectManager == nil && !isDeleteOp(ctx, m) {
					return next.Mutate(ctx, m)
				}

				var ids []string

				switch m.Op() {
				case ent.OpDelete:
					dbIDs, err := m.IDs(ctx)
					if err != nil {
						return nil, err
					}

					ids = append(ids, dbIDs...)

				case ent.OpDeleteOne:

					id, ok := m.ID()
					if !ok {
						return nil, errInvalidStoragePath
					}

					ids = append(ids, id)
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				files, err := m.Client().File.Query().Where(file.IDIn(ids...)).
					Select(
						file.FieldID,
						file.FieldStoragePath,
						file.FieldStorageProvider,
						file.FieldDetectedContentType,
						file.FieldPersistedFileSize,
						file.FieldMetadata,
						file.FieldStorageVolume,
					).
					WithIntegrations().
					WithSecrets().
					All(ctx)
				if err != nil {
					return nil, err
				}

				log.Debug().Interface("files", files).Msg("deleting files from object storage")

				for _, f := range files {
					if f.StoragePath != "" && m.ObjectManager != nil {
						// Convert ent File to storagetypes.File
						storageFile := &storagetypes.File{
							ID:           f.ID,
							OriginalName: f.ProvidedFileName,
							FileMetadata: storagetypes.FileMetadata{
								Key:           f.StoragePath,
								ContentType:   f.DetectedContentType,
								Size:          f.PersistedFileSize,
								ProviderHints: &storagetypes.ProviderHints{},
							},
						}

						// Set integration ID from edges
						for _, integration := range f.Edges.Integrations {
							storageFile.ProviderHints.IntegrationID = integration.ID
						}

						// Set hush ID from edges
						for _, secret := range f.Edges.Secrets {
							storageFile.ProviderHints.HushID = secret.ID
						}

						// Convert metadata from map[string]interface{} to map[string]string
						if f.Metadata != nil {
							metadata := make(map[string]string)
							for k, v := range f.Metadata {
								if str, ok := v.(string); ok {
									metadata[k] = str
								}
							}
							storageFile.Metadata = metadata
						}

						// Set provider-specific fields if available
						if f.StorageProvider != "" {
							storageFile.ProviderType = storagetypes.ProviderType(f.StorageProvider)
						}

						if err := m.ObjectManager.Delete(ctx, storageFile, nil); err != nil {
							log.Error().Err(err).Str("file_id", f.ID).Msg("failed to delete file from storage")
							// Continue with other files rather than failing the entire operation
						}
					}
				}

				return v, err
			})
	}, ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne)
}
