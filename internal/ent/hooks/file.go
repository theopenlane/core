package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/logx"
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

				ids, err := getMutationIDs(ctx, m)
				if err != nil {
					return nil, err
				}
				if len(ids) == 0 {
					return nil, errInvalidStoragePath
				}

				// the storage metadata has to be read before the mutation runs, a hard delete
				// removes the rows entirely so there is nothing left to look up afterwards
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
						file.FieldBackupState,
					).All(ctx)
				if err != nil {
					return nil, err
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				logx.FromContext(ctx).Debug().Interface("files", files).Msg("deleting files from object storage")

				for _, f := range files {
					if f.StoragePath != "" && m.ObjectManager != nil {
						// Convert ent File to storagetypes.File
						storageFile := &storagetypes.File{
							ID:            f.ID,
							OriginalName:  f.ProvidedFileName,
							Key:           f.StoragePath,
							ContentType:   f.DetectedContentType,
							Size:          f.PersistedFileSize,
							Bucket:        f.StorageVolume,
							Region:        f.StorageRegion,
							ProviderHints: &storagetypes.ProviderHints{},
						}

						if f.Metadata != nil {
							metadata := make(map[string]any)
							for k, v := range f.Metadata {
								metadata[k] = v
							}
							storageFile.Metadata = metadata
						}

						// Set provider-specific fields if available
						if f.StorageProvider != "" {
							storageFile.ProviderType = storagetypes.ProviderType(f.StorageProvider)
						}

						storageFile.BackupLocation = f.BackupState.Location()

						// during a disaster recovery event the source storage is presumed lost, so the
						// replica is the only remaining copy and deleting the source would only fail
						_, readFromBackup := m.ObjectManager.ReadFromBackupFor(storageFile.ProviderType)

						if !readFromBackup {
							if err := m.ObjectManager.Delete(ctx, storageFile, nil); err != nil {
								logx.FromContext(ctx).Error().Err(err).Str("fileID", f.ID).Msg("failed to delete file from object storage")

								return nil, err
							}
						}

						// the replica shares the lifetime of the object it backs up, so it is deleted
						// with the same semantics as the source
						if err := m.ObjectManager.DeleteBackup(ctx, storageFile, nil); err != nil {
							logx.FromContext(ctx).Error().Err(err).Str("fileID", f.ID).Msg("failed to delete file from backup storage")

							return nil, err
						}
					}
				}

				return v, err
			})
	}, ent.OpDelete|ent.OpDeleteOne)
}

// fileEdgeMutation is a mutation on a schema with a files edge
type fileEdgeMutation interface {
	utils.GenericMutation
	AddFileIDs(ids ...string)
}

// HookFiles runs on mutations with a files edge to check for uploaded files under key
func HookFiles(key string) ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(fileEdgeMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			ctx, err := checkMutationFiles(ctx, mut, key)
			if err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookSingleFile runs on mutations to check for a single uploaded file under key and set it on field
func HookSingleFile(key, field string) ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(utils.GenericMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
			if len(files) == 0 {
				return next.Mutate(ctx, m)
			}

			if len(files) > 1 {
				return nil, fmt.Errorf("%w: %s", ErrTooManyFiles, key)
			}

			if err := m.SetField(field, files[0].ID); err != nil {
				return nil, err
			}

			ctx, err := pkgobjects.ProcessFilesForMutation(ctx, mut, key)
			if err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookLogoFile runs on mutations to check for an uploaded logo file
func HookLogoFile() ent.Hook {
	return HookSingleFile("logoFile", "logo_file_id")
}

// checkMutationFiles checks if files are provided under key and sets the local file ID(s)
func checkMutationFiles(ctx context.Context, m fileEdgeMutation, key string) (context.Context, error) {
	files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	fileIDs := make([]string, len(files))
	for i, f := range files {
		fileIDs[i] = f.ID
	}

	m.AddFileIDs(fileIDs...)

	return pkgobjects.ProcessFilesForMutation(ctx, m, key)
}
