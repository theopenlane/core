package objects

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
	"gotest.tools/v3/assert"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/metrics"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/eddy"
)

// resolverForProvider builds a resolver + client service that always resolves to the given provider,
// standing in for the source provider a backup downloads from
func resolverForProvider(source storage.Provider) (*eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], *eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]) {
	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](pool)

	builder := &eddy.BuilderFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Type: source.ProviderType().String(),
		Func: func(context.Context, storage.ProviderCredentials, *storage.ProviderOptions) (storage.Provider, error) {
			return source, nil
		},
	}

	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	resolver.AddRule(&eddy.RuleFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		EvaluateFunc: func(context.Context) mo.Option[eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]] {
			return mo.Some(eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: builder,
				Output:  storage.ProviderCredentials{},
				Config:  storage.NewProviderOptions(),
			})
		},
	})

	return resolver, clientService
}

func TestServiceBackupProviderForAndSources(t *testing.T) {
	dest := &fakeProvider{id: "s3"}
	svc := NewService(Config{
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	got, ok := svc.BackupProviderFor(storage.R2Provider)
	assert.Assert(t, ok)
	assert.Equal(t, got, storage.Provider(dest))

	_, ok = svc.BackupProviderFor(storage.S3Provider)
	assert.Assert(t, !ok)

	sources := svc.BackupSources()
	assert.Equal(t, len(sources), 1)
	assert.Equal(t, sources[0], storage.R2Provider)
}

func TestServiceBackupNoBackupConfigured(t *testing.T) {
	svc := NewService(Config{})

	result, err := svc.Backup(context.Background(), &storagetypes.File{ProviderType: storage.R2Provider})
	assert.NilError(t, err)
	assert.Assert(t, result == nil)
}

func TestServiceBackupReplicatesToDestination(t *testing.T) {
	data := []byte("hello world")

	source := &fakeProvider{
		id: string(storage.R2Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{
			File: data,
			Size: int64(len(data)),
		},
	}
	dest := &fakeProvider{
		id: string(storage.S3Provider),
		uploadMetadata: &storagetypes.UploadedFileMetadata{
			FileMetadata: storagetypes.FileMetadata{
				Bucket:  "backup-bucket",
				FullURI: "s3://backup-bucket/org/file/name",
			},
		},
	}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	file := &storagetypes.File{
		ID:           "file",
		ProviderType: storage.R2Provider,
		FileMetadata: storagetypes.FileMetadata{
			Key:         "org/file/name",
			ContentType: "text/plain",
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: storage.R2Provider,
			},
		},
	}

	result, err := svc.Backup(context.Background(), file)
	assert.NilError(t, err)
	assert.Assert(t, result != nil)
	assert.Equal(t, result.Provider, storage.S3Provider)
	assert.Equal(t, result.Bucket, "backup-bucket")
	assert.Equal(t, result.URI, "s3://backup-bucket/org/file/name")
	assert.Equal(t, result.Bytes, int64(len(data)))

	// the destination did not echo a key back, so the source key is recorded instead
	assert.Equal(t, result.Key, "org/file/name")

	// the backup writes to the destination once, under the same object key as the source
	assert.Equal(t, dest.uploadCallCount, 1)
	assert.Equal(t, dest.lastUploadOpts.FileName, "org/file/name")
}

func TestServiceBackupRecordsDestinationKeyAndRegion(t *testing.T) {
	data := []byte("hello world")

	source := &fakeProvider{
		id: string(storage.R2Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{
			File: data,
			Size: int64(len(data)),
		},
	}

	// the destination derives its own key, which is what must be recorded so a restore can find it
	dest := &fakeProvider{
		id: string(storage.S3Provider),
		uploadMetadata: &storagetypes.UploadedFileMetadata{
			FileMetadata: storagetypes.FileMetadata{
				Bucket:  "backup-bucket",
				Key:     "backup/prefix/org/file/name",
				Region:  "us-west-2",
				FullURI: "s3://backup-bucket/backup/prefix/org/file/name",
			},
		},
	}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	file := &storagetypes.File{
		ID:           "file",
		ProviderType: storage.R2Provider,
		FileMetadata: storagetypes.FileMetadata{
			Key: "org/file/name",
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: storage.R2Provider,
			},
		},
	}

	result, err := svc.Backup(context.Background(), file)
	assert.NilError(t, err)
	assert.Assert(t, result != nil)
	assert.Equal(t, result.Key, "backup/prefix/org/file/name")
	assert.Equal(t, result.Region, "us-west-2")
}

func TestServiceReadFromBackupFor(t *testing.T) {
	dest := &fakeProvider{id: "s3"}
	svc := NewService(Config{
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider:   {Provider: dest, ReadFromBackup: true},
			storage.DiskProvider: {Provider: dest},
		},
	})

	got, ok := svc.ReadFromBackupFor(storage.R2Provider)
	assert.Assert(t, ok)
	assert.Equal(t, got, storage.Provider(dest))

	// a backup target without the flag is still a backup destination, but reads do not use it
	_, ok = svc.ReadFromBackupFor(storage.DiskProvider)
	assert.Assert(t, !ok)

	_, ok = svc.BackupProviderFor(storage.DiskProvider)
	assert.Assert(t, ok)

	_, ok = svc.ReadFromBackupFor(storage.S3Provider)
	assert.Assert(t, !ok)
}

func TestServiceBackupDownloadErrorSkipsUpload(t *testing.T) {
	downloadErr := errors.New("download failed")

	source := &fakeProvider{id: string(storage.R2Provider), downloadErr: downloadErr}
	dest := &fakeProvider{id: string(storage.S3Provider)}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	file := &storagetypes.File{
		ID:           "file",
		ProviderType: storage.R2Provider,
		FileMetadata: storagetypes.FileMetadata{
			Key: "org/file/name",
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: storage.R2Provider,
			},
		},
	}

	result, err := svc.Backup(context.Background(), file)
	assert.ErrorIs(t, err, downloadErr)
	assert.Assert(t, result == nil)
	assert.Equal(t, dest.uploadCallCount, 0)
}

// fileWithReplica builds a source file carrying a completed replica location at the backup provider
func fileWithReplica() *storagetypes.File {
	return &storagetypes.File{
		ID:           "file",
		ProviderType: storage.R2Provider,
		BackupLocation: &storagetypes.BackupLocation{
			ProviderType: storage.S3Provider,
			Bucket:       "backup-bucket",
			Key:          "backup/prefix/org/file/name",
			Region:       "us-west-2",
			FullURI:      "s3://backup-bucket/backup/prefix/org/file/name",
		},
		FileMetadata: storagetypes.FileMetadata{
			Key:    "org/file/name",
			Bucket: "source-bucket",
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: storage.R2Provider,
			},
		},
	}
}

func TestServiceDownloadReadsFromBackup(t *testing.T) {
	data := []byte("replica contents")

	source := &fakeProvider{id: string(storage.R2Provider)}
	dest := &fakeProvider{
		id: string(storage.S3Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{
			File: data,
			Size: int64(len(data)),
		},
	}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest, ReadFromBackup: true},
		},
	})

	downloaded, err := svc.Download(context.Background(), nil, fileWithReplica(), &storage.DownloadOptions{})
	assert.NilError(t, err)
	assert.DeepEqual(t, downloaded.File, data)

	// the source provider is never contacted, this is routing rather than a retry
	assert.Assert(t, source.lastDownloadFile == nil)

	// the replica location replaces the source location so the backup provider can find the object
	assert.Assert(t, dest.lastDownloadFile != nil)
	assert.Equal(t, dest.lastDownloadFile.Key, "backup/prefix/org/file/name")
	assert.Equal(t, dest.lastDownloadFile.Bucket, "backup-bucket")
	assert.Equal(t, dest.lastDownloadFile.Region, "us-west-2")
	assert.Equal(t, dest.lastDownloadFile.ProviderType, storage.S3Provider)
	assert.Equal(t, dest.lastDownloadFile.ProviderHints.KnownProvider, storage.S3Provider)
}

func TestServiceDownloadUsesSourceWhenReadFromBackupDisabled(t *testing.T) {
	data := []byte("source contents")

	source := &fakeProvider{
		id: string(storage.R2Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{
			File: data,
			Size: int64(len(data)),
		},
	}
	dest := &fakeProvider{id: string(storage.S3Provider)}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	downloaded, err := svc.Download(context.Background(), nil, fileWithReplica(), &storage.DownloadOptions{})
	assert.NilError(t, err)
	assert.DeepEqual(t, downloaded.File, data)

	assert.Assert(t, dest.lastDownloadFile == nil)
	assert.Assert(t, source.lastDownloadFile != nil)
	assert.Equal(t, source.lastDownloadFile.Key, "org/file/name")
}

func TestServiceDownloadUsesSourceWithoutReplica(t *testing.T) {
	data := []byte("source contents")

	source := &fakeProvider{
		id: string(storage.R2Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{
			File: data,
			Size: int64(len(data)),
		},
	}
	dest := &fakeProvider{id: string(storage.S3Provider)}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest, ReadFromBackup: true},
		},
	})

	// a file that was never replicated falls through to the source provider
	file := fileWithReplica()
	file.BackupLocation = nil

	downloaded, err := svc.Download(context.Background(), nil, file, &storage.DownloadOptions{})
	assert.NilError(t, err)
	assert.DeepEqual(t, downloaded.File, data)

	assert.Assert(t, dest.lastDownloadFile == nil)
	assert.Assert(t, source.lastDownloadFile != nil)
}

func TestServiceGetPresignedURLReadsFromBackup(t *testing.T) {
	source := &fakeProvider{id: string(storage.R2Provider), presignURL: "https://source/url"}
	dest := &fakeProvider{id: string(storage.S3Provider), presignURL: "https://backup/url"}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest, ReadFromBackup: true},
		},
	})

	url, err := svc.GetPresignedURL(context.Background(), fileWithReplica(), time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, url, "https://backup/url")

	assert.Equal(t, source.presignCallCount, 0)
	assert.Equal(t, dest.presignCallCount, 1)
	assert.Equal(t, dest.lastPresignFile.Key, "backup/prefix/org/file/name")
}

func TestServiceDeleteBackupRemovesReplica(t *testing.T) {
	source := &fakeProvider{id: string(storage.R2Provider)}
	dest := &fakeProvider{id: string(storage.S3Provider)}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	err := svc.DeleteBackup(context.Background(), fileWithReplica(), nil)
	assert.NilError(t, err)

	// the replica is deleted at its own location, leaving the source untouched
	assert.Assert(t, dest.lastDeleteFile != nil)
	assert.Equal(t, dest.lastDeleteFile.Key, "backup/prefix/org/file/name")
	assert.Equal(t, dest.lastDeleteFile.Bucket, "backup-bucket")
	assert.Assert(t, source.lastDeleteFile == nil)
}

func TestServiceDeleteBackupWithoutReplicaIsNoop(t *testing.T) {
	dest := &fakeProvider{id: string(storage.S3Provider)}

	svc := NewService(Config{
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	file := fileWithReplica()
	file.BackupLocation = nil

	assert.NilError(t, svc.DeleteBackup(context.Background(), file, nil))
	assert.Assert(t, dest.lastDeleteFile == nil)
}

func TestServiceDeleteBackupPropagatesError(t *testing.T) {
	deleteErr := errors.New("backup delete failed")
	dest := &fakeProvider{id: string(storage.S3Provider), deleteErr: deleteErr}

	svc := NewService(Config{
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	// the replica delete fails the same way a source delete does
	assert.ErrorIs(t, svc.DeleteBackup(context.Background(), fileWithReplica(), nil), deleteErr)
}

func TestServiceBackupReadMetrics(t *testing.T) {
	source := &fakeProvider{
		id: string(storage.R2Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{
			File: []byte("contents"),
		},
	}
	dest := &fakeProvider{
		id: string(storage.S3Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{
			File: []byte("contents"),
		},
	}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest, ReadFromBackup: true},
		},
	})

	reads := metrics.StorageBackupReads.WithLabelValues(
		string(storage.R2Provider), string(storage.S3Provider), backupOpDownload)
	missing := metrics.StorageBackupReadsMissingReplica.WithLabelValues(
		string(storage.R2Provider), backupOpDownload)

	readsBefore := testutil.ToFloat64(reads)
	missingBefore := testutil.ToFloat64(missing)

	_, err := svc.Download(context.Background(), nil, fileWithReplica(), &storage.DownloadOptions{})
	assert.NilError(t, err)

	assert.Equal(t, testutil.ToFloat64(reads), readsBefore+1)
	assert.Equal(t, testutil.ToFloat64(missing), missingBefore)

	// a file with no replica falls through to the source and is counted as a gap instead
	withoutReplica := fileWithReplica()
	withoutReplica.BackupLocation = nil

	_, err = svc.Download(context.Background(), nil, withoutReplica, &storage.DownloadOptions{})
	assert.NilError(t, err)

	assert.Equal(t, testutil.ToFloat64(reads), readsBefore+1)
	assert.Equal(t, testutil.ToFloat64(missing), missingBefore+1)
}

func TestServiceDeleteBackupMetrics(t *testing.T) {
	dest := &fakeProvider{id: string(storage.S3Provider)}

	svc := NewService(Config{
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	deletes := metrics.StorageBackupDeletes.WithLabelValues(
		string(storage.R2Provider), string(storage.S3Provider), "success")

	before := testutil.ToFloat64(deletes)

	assert.NilError(t, svc.DeleteBackup(context.Background(), fileWithReplica(), nil))
	assert.Equal(t, testutil.ToFloat64(deletes), before+1)
}

func TestServiceExistsReadsFromBackup(t *testing.T) {
	source := &fakeProvider{id: string(storage.R2Provider)}
	dest := &fakeProvider{id: string(storage.S3Provider), existsResult: true}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest, ReadFromBackup: true},
		},
	})

	exists, err := svc.Exists(context.Background(), fileWithReplica())
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	// existence is answered by the replica, the source is never contacted
	assert.Assert(t, source.lastExistsFile == nil)
	assert.Assert(t, dest.lastExistsFile != nil)
	assert.Equal(t, dest.lastExistsFile.Key, "backup/prefix/org/file/name")
	assert.Equal(t, dest.lastExistsFile.Bucket, "backup-bucket")
	assert.Equal(t, dest.lastExistsFile.ProviderType, storage.S3Provider)
}

func TestServiceExistsUsesSourceWhenReadFromBackupDisabled(t *testing.T) {
	source := &fakeProvider{id: string(storage.R2Provider), existsResult: true}
	dest := &fakeProvider{id: string(storage.S3Provider)}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest, ReadFromBackup: false},
		},
	})

	exists, err := svc.Exists(context.Background(), fileWithReplica())
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	assert.Assert(t, dest.lastExistsFile == nil)
	assert.Assert(t, source.lastExistsFile != nil)
}
