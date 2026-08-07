package objects

import (
	"bytes"
	"context"
	"time"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// RestoreResult describes where a file was written back to its source provider during a restore
type RestoreResult struct {
	// Provider is the source provider the replica was written back to
	Provider storage.ProviderType
	// Bucket is the bucket the restored object was written to
	Bucket string
	// Key is the object key the restored object was written to, which is not guaranteed to match
	// the key the file previously had because the replacement storage derives its own
	Key string
	// Region is the region the restored object was written to, if applicable
	Region string
	// URI is the full URI of the restored object
	URI string
	// Bytes is the number of bytes restored
	Bytes int64
}

// Restore copies a file's replica at its backup provider back to its source provider, which is how
// a source recovers after its storage is lost and replaced. A file with no usable replica, or whose
// source provider has no backup configured, is a no-op
func (s *Service) Restore(ctx context.Context, file *storagetypes.File) (*RestoreResult, error) {
	if file == nil {
		return nil, ErrMissingFileID
	}

	if file.BackupLocation == nil || file.BackupLocation.Key == "" {
		return nil, nil
	}

	source := file.ProviderType

	target, ok := s.backups[source]
	if !ok {
		return nil, nil
	}

	destination := file.BackupLocation.ProviderType
	start := time.Now()

	// read the replica directly rather than through Download, so a restore does not depend on
	// whether reads are currently configured to be served from the backup
	downloaded, err := s.objectService.Download(ctx, target.Provider, replicaFile(file), &storage.DownloadOptions{})
	if err != nil {
		metrics.RecordStorageRestore(string(destination), string(source), 0, time.Since(start).Seconds(), err)

		return nil, err
	}

	// resolve the source provider itself, which after a recovery points at the replacement storage
	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		metrics.RecordStorageRestore(string(destination), string(source), 0, time.Since(start).Seconds(), err)

		return nil, err
	}

	if provider == nil {
		metrics.RecordStorageRestore(string(destination), string(source), 0, time.Since(start).Seconds(), ErrProviderResolutionFailed)

		return nil, ErrProviderResolutionFailed
	}

	uploaded, err := provider.Upload(ctx, bytes.NewReader(downloaded.File), &storage.UploadOptions{
		FileName:    file.Key,
		ContentType: file.ContentType,
		Bucket:      file.Bucket,
	})
	metrics.RecordStorageRestore(string(destination), string(source), downloaded.Size, time.Since(start).Seconds(), err)

	if err != nil {
		return nil, err
	}

	// providers do not all echo the key back, so fall back to the key the upload was requested with
	key := uploaded.Key
	if key == "" {
		key = file.Key
	}

	return &RestoreResult{
		Provider: source,
		Bucket:   uploaded.Bucket,
		Key:      key,
		Region:   uploaded.Region,
		URI:      uploaded.FullURI,
		Bytes:    downloaded.Size,
	}, nil
}
