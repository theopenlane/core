package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaFileBackupListeners registers the listener that asynchronously replicates a file to
// its configured backup provider once the file's storage location has been written.
func RegisterGalaFileBackupListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeFile),
		Name:       "file.backup",
		Operations: []string{ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String()},
		Handle:     handleFileBackup,
	})
}

// EnqueueFileBackup dispatches the same direct File mutation event the upload path produces
func EnqueueFileBackup(ctx context.Context, g *gala.Gala, fileID string) error {
	payload := eventqueue.MutationGalaPayload{
		MutationType:  generated.TypeFile,
		Operation:     ent.OpUpdate.String(),
		EntityID:      fileID,
		ChangedFields: []string{file.FieldStorageProvider},
	}

	metadata := eventqueue.NewMutationGalaMetadata(fileID, payload)
	topic := string(eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, generated.TypeFile))

	return enqueueGalaMutation(ctx, g, topic, payload, metadata)
}

// handleFileBackup replicates a file to its backup provider
func handleFileBackup(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !eventqueue.MutationFieldChanged(payload, file.FieldStorageProvider) {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	svc := client.ObjectManager
	if svc == nil {
		return nil
	}

	fileID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || fileID == "" {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx.Context)

	f, err := client.File.Get(allowCtx, fileID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	source := storagetypes.ProviderType(f.StorageProvider)

	// nothing to do when this provider has no backup configured or the file is already backed up
	if _, ok := svc.BackupProviderFor(source); !ok {
		return nil
	}

	if f.BackupState.Status == enums.FileBackupStatusCompleted {
		return nil
	}

	storageFile := &storagetypes.File{
		ID:           f.ID,
		ProviderType: source,
		FileMetadata: storagetypes.FileMetadata{
			Key:         f.StoragePath,
			Bucket:      f.StorageVolume,
			Region:      f.StorageRegion,
			ContentType: f.DetectedContentType,
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: source,
			},
		},
	}

	result, backupErr := svc.Backup(allowCtx, storageFile)

	state := models.FileBackupState{
		Attempts: f.BackupState.Attempts + 1,
	}

	if backupErr != nil {
		state.Status = enums.FileBackupStatusFailed
		state.Error = backupErr.Error()

		if updateErr := persistBackupState(allowCtx, client, fileID, state); updateErr != nil {
			logx.FromContext(allowCtx).Err(updateErr).Str("file_id", fileID).Msg("failed to persist file backup failure state")
		}

		// return the error so gala retries the backup
		return backupErr
	}

	now := time.Now()
	state.Status = enums.FileBackupStatusCompleted
	state.Provider = string(result.Provider)
	state.Bucket = result.Bucket
	state.URI = result.URI
	state.CompletedAt = &now

	return persistBackupState(allowCtx, client, fileID, state)
}

// persistBackupState writes the backup replication state onto the file record
func persistBackupState(ctx context.Context, client *generated.Client, fileID string, state models.FileBackupState) error {
	return client.File.UpdateOneID(fileID).
		SetBackupState(state).
		Exec(ctx)
}
