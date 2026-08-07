package objects

import (
	"context"
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestServiceRestoreWritesReplicaBackToSource(t *testing.T) {
	data := []byte("replica contents")

	source := &fakeProvider{
		id: string(storage.R2Provider),
		uploadMetadata: &storagetypes.UploadedFileMetadata{
			FileMetadata: storagetypes.FileMetadata{
				Bucket:  "restored-bucket",
				Key:     "org/file/name",
				Region:  "us-east-1",
				FullURI: "r2://restored-bucket/org/file/name",
			},
		},
	}
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
			storage.R2Provider: {Provider: dest},
		},
	})

	result, err := svc.Restore(context.Background(), fileWithReplica())
	assert.NilError(t, err)
	assert.Assert(t, result != nil)
	assert.Equal(t, result.Provider, storage.R2Provider)
	assert.Equal(t, result.Key, "org/file/name")
	assert.Equal(t, result.Bucket, "restored-bucket")
	assert.Equal(t, result.Region, "us-east-1")
	assert.Equal(t, result.Bytes, int64(len(data)))

	// the replica is read at its own location and written back under the source key
	assert.Equal(t, dest.lastDownloadFile.Key, "backup/prefix/org/file/name")
	assert.Equal(t, source.uploadCallCount, 1)
	assert.Equal(t, source.lastUploadOpts.FileName, "org/file/name")
}

func TestServiceRestoreReadsBackupRegardlessOfReadFromBackup(t *testing.T) {
	// a restore must read the replica whether or not reads are currently served from the backup,
	// otherwise recovery could never write anything back to the source
	for _, readFromBackup := range []bool{true, false} {
		source := &fakeProvider{
			id:             string(storage.R2Provider),
			uploadMetadata: &storagetypes.UploadedFileMetadata{},
		}
		dest := &fakeProvider{
			id:               string(storage.S3Provider),
			downloadMetadata: &storagetypes.DownloadedFileMetadata{File: []byte("contents")},
		}

		resolver, clientService := resolverForProvider(source)
		svc := NewService(Config{
			Resolver:      resolver,
			ClientService: clientService,
			Backups: map[storage.ProviderType]BackupTarget{
				storage.R2Provider: {Provider: dest, ReadFromBackup: readFromBackup},
			},
		})

		result, err := svc.Restore(context.Background(), fileWithReplica())
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Equal(t, dest.lastDownloadFile.Key, "backup/prefix/org/file/name")
		assert.Equal(t, source.uploadCallCount, 1)
	}
}

func TestServiceRestoreWithoutReplicaIsNoop(t *testing.T) {
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

	file := fileWithReplica()
	file.BackupLocation = nil

	result, err := svc.Restore(context.Background(), file)
	assert.NilError(t, err)
	assert.Assert(t, result == nil)
	assert.Equal(t, source.uploadCallCount, 0)
}

func TestServiceRestoreDownloadErrorSkipsUpload(t *testing.T) {
	downloadErr := errors.New("replica unreadable")

	source := &fakeProvider{id: string(storage.R2Provider)}
	dest := &fakeProvider{id: string(storage.S3Provider), downloadErr: downloadErr}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	result, err := svc.Restore(context.Background(), fileWithReplica())
	assert.ErrorIs(t, err, downloadErr)
	assert.Assert(t, result == nil)
	assert.Equal(t, source.uploadCallCount, 0)
}

func TestServiceRestoreMetrics(t *testing.T) {
	source := &fakeProvider{
		id:             string(storage.R2Provider),
		uploadMetadata: &storagetypes.UploadedFileMetadata{},
	}
	dest := &fakeProvider{
		id:               string(storage.S3Provider),
		downloadMetadata: &storagetypes.DownloadedFileMetadata{File: []byte("contents"), Size: 8},
	}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]BackupTarget{
			storage.R2Provider: {Provider: dest},
		},
	})

	attempts := metrics.StorageRestoreAttempts.WithLabelValues(
		string(storage.S3Provider), string(storage.R2Provider), "success")
	restoredBytes := metrics.StorageRestoreBytes.WithLabelValues(
		string(storage.S3Provider), string(storage.R2Provider))

	attemptsBefore := testutil.ToFloat64(attempts)
	bytesBefore := testutil.ToFloat64(restoredBytes)

	_, err := svc.Restore(context.Background(), fileWithReplica())
	assert.NilError(t, err)

	assert.Equal(t, testutil.ToFloat64(attempts), attemptsBefore+1)
	assert.Equal(t, testutil.ToFloat64(restoredBytes), bytesBefore+8)
}
