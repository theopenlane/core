package hooks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/common/storagetypes"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/metrics"
)

// init registers the file backup listeners so gala setup picks them up automatically
func init() { registerListeners(FileBackupListeners) }

// FileBackupRequest asks for a single file to be replicated to its backup provider
type FileBackupRequest struct {
	FileID string `json:"file_id"`
}

// FileBackupTopic carries explicit replication requests to enable backfill of backups
var FileBackupTopic = gala.NamespacedTopic(gala.System, "file.backup.requested",
	gala.WithUniqueKey(func(req FileBackupRequest) string {
		return "file-backup-" + req.FileID
	}),
)

// MaxFileBackupAttempts caps replication retries before a file is marked exhausted
const MaxFileBackupAttempts = 5

// ErrFileBackupExhausted marks a replication that hit the attempt cap and will not be retried
var ErrFileBackupExhausted = errors.New("file backup exhausted max attempts")

// FileBackupListeners replicates a file to its configured backup provider once the file's storage
// location has been written
func FileBackupListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaFile,
			Operations: []string{entityops.OpCreate},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleFileBackup,
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaFile,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{file.FieldStorageProvider}, // only check when the provider is changed
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleFileBackup,
		},
		gala.Definition[FileBackupRequest]{
			Topic: FileBackupTopic,
			Caller: func(restored *auth.Caller, _ FileBackupRequest) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleFileBackupRequest,
		},
	}
}

// handleFileBackupRequest replicates the file named by an explicit backup request; the request
// topic is not a mutation topic, so the invocation the entityops preamble would build is
// assembled from the handler context here
func handleFileBackupRequest(hctx gala.HandlerContext, req FileBackupRequest) error {
	// client is set in context by the caller
	client := generated.FromContext(hctx.Context)
	if client == nil {
		return fmt.Errorf("%w: unable to handle file backup request", ErrClientResolveFailed)
	}

	return replicateFileToBackup(hctx.Context, client, req.FileID)
}

// handleFileBackup replicates a file to its backup provider
func handleFileBackup(inv entityops.Invocation, _ entityops.MutationPayload) error {
	return replicateFileToBackup(inv.Context, inv.Client, inv.EntityID)
}

// replicateFileToBackup replicates a file to the backup provider configured for its source provider and records the outcome
// A file with no backup configured, reading from its backup, or already replicated is a no-op
func replicateFileToBackup(ctx context.Context, client *generated.Client, fileID string) error {
	svc := client.ObjectManager
	if svc == nil {
		return nil
	}

	ctx = logx.WithFields(ctx, map[string]any{"file_id": fileID})

	f, ok, err := entityops.LoadEntity(ctx, fileID, client.File.Get)
	if err != nil || !ok {
		return err
	}

	source := storagetypes.ProviderType(f.StorageProvider)

	if _, hasBackup := svc.BackupProviderFor(source); !hasBackup {
		return nil
	}

	if _, readFromBackup := svc.ReadFromBackupFor(source); readFromBackup {
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

	destination := ""
	if dest, hasDest := svc.BackupProviderFor(source); hasDest {
		destination = dest.ProviderType().String()
	}

	result, backupErr := svc.Backup(ctx, storageFile)

	state := models.FileBackupState{
		Attempts: f.BackupState.Attempts + 1,
	}

	if backupErr != nil {
		state.Status = enums.FileBackupStatusFailed
		state.Error = backupErr.Error()

		exhausted := state.Attempts >= MaxFileBackupAttempts
		if exhausted {
			state.Status = enums.FileBackupStatusExhausted
		}

		metrics.RecordStorageBackupState(string(source), destination, string(state.Status))

		if updateErr := persistBackupState(ctx, client, fileID, state); updateErr != nil {
			logx.FromContext(ctx).Err(updateErr).Msg("failed to persist file backup failure state")
		}

		if exhausted {
			logx.FromContext(ctx).Warn().Err(backupErr).Int("attempts", state.Attempts).Msg("file backup exhausted max attempts; giving up")

			// cancelled with the cause instead of retried or reported as a success
			return river.JobCancel(fmt.Errorf("%w: %w", ErrFileBackupExhausted, backupErr))
		}

		// retried by the queue
		return backupErr
	}

	now := time.Now()
	state.Status = enums.FileBackupStatusCompleted
	state.Provider = string(result.Provider)
	state.Bucket = result.Bucket
	state.Key = result.Key
	state.Region = result.Region
	state.URI = result.URI
	state.CompletedAt = &now

	metrics.RecordStorageBackupState(string(source), destination, string(state.Status))

	return persistBackupState(ctx, client, fileID, state)
}

// persistBackupState writes the backup replication state onto the file record
func persistBackupState(ctx context.Context, client *generated.Client, fileID string, state models.FileBackupState) error {
	return client.File.UpdateOneID(fileID).
		SetBackupState(state).
		Exec(ctx)
}
