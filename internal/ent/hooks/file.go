package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	errInvalidStoragePath = errors.New("invalid path when deleting file from object storage")
)

// HookFileDelete makes sure to clean up the file from external storage once deleted
func HookFileDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.FileFunc(
			func(ctx context.Context, m *generated.FileMutation) (generated.Value, error) {

				if m.ObjectManager == nil || !isDeleteOp(ctx, m) {
					return next.Mutate(ctx, m)
				}

				ids := getMutationIDs(ctx, m)
				if len(ids) == 0 {
					return nil, errInvalidStoragePath
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
						file.FieldStorageRegion,
					).All(ctx)
				if err != nil {
					return nil, err
				}

				logx.FromContext(ctx).Debug().Interface("files", files).Msg("deleting files from object storage")

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
								Bucket:        f.StorageVolume,
								Region:        f.StorageRegion,
								ProviderHints: &storagetypes.ProviderHints{},
							},
						}

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
							logx.FromContext(ctx).Error().Err(err).Str("fileID", f.ID).Msg("failed to delete file from object storage")

							return nil, err
						}
					}
				}

				return v, err
			})
	}, ent.OpDelete|ent.OpDeleteOne)
}
