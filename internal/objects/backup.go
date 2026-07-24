package objects

import (
	"bytes"
	"context"
	"time"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// BackupResult describes where a file was replicated during a backup operation
type BackupResult struct {
	// Provider is the destination provider the file was replicated to
	Provider storage.ProviderType
	// Bucket is the destination bucket the backup was written to
	Bucket string
	// URI is the full URI of the replicated object at the destination
	URI string
	// Bytes is the number of bytes replicated
	Bytes int64
}

// BackupSources returns the source provider types that have a backup destination configured
func (s *Service) BackupSources() []storage.ProviderType {
	sources := make([]storage.ProviderType, 0, len(s.backups))
	for source := range s.backups {
		sources = append(sources, source)
	}

	return sources
}

// Backup replicates a file from its source provider to the configured backup destination for that provider
// no backup provider defined is a no-op
func (s *Service) Backup(ctx context.Context, file *storagetypes.File) (*BackupResult, error) {
	if file == nil {
		return nil, ErrMissingFileID
	}

	source := file.ProviderType

	dest, ok := s.BackupProviderFor(source)
	if !ok {
		return nil, nil
	}

	destination := dest.ProviderType()
	start := time.Now()

	downloaded, err := s.Download(ctx, nil, file, &storage.DownloadOptions{})
	if err != nil {
		metrics.RecordStorageBackup(string(source), string(destination), 0, time.Since(start).Seconds(), err)
		return nil, err
	}

	uploaded, err := dest.Upload(ctx, bytes.NewReader(downloaded.File), &storage.UploadOptions{
		FileName:    file.Key,
		ContentType: file.ContentType,
	})
	metrics.RecordStorageBackup(string(source), string(destination), downloaded.Size, time.Since(start).Seconds(), err)

	if err != nil {
		return nil, err
	}

	return &BackupResult{
		Provider: destination,
		Bucket:   uploaded.Bucket,
		URI:      uploaded.FullURI,
		Bytes:    downloaded.Size,
	}, nil
}
