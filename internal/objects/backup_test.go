package objects

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/eddy"
)

// resolverForProvider builds a resolver + client service that always resolves to the given provider,
// standing in for the source provider a backup downloads from
func resolverForProvider(source storage.Provider) (*eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], *eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]) {
	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](pool)

	builder := &eddy.BuilderFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Type: string(source.ProviderType()),
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
		Backups: map[storage.ProviderType]storage.Provider{
			storage.R2Provider: dest,
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
		Backups: map[storage.ProviderType]storage.Provider{
			storage.R2Provider: dest,
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

	// the backup writes to the destination once, under the same object key as the source
	assert.Equal(t, dest.uploadCallCount, 1)
	assert.Equal(t, dest.lastUploadOpts.FileName, "org/file/name")
}

func TestServiceBackupDownloadErrorSkipsUpload(t *testing.T) {
	downloadErr := errors.New("download failed")

	source := &fakeProvider{id: string(storage.R2Provider), downloadErr: downloadErr}
	dest := &fakeProvider{id: string(storage.S3Provider)}

	resolver, clientService := resolverForProvider(source)
	svc := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
		Backups: map[storage.ProviderType]storage.Provider{
			storage.R2Provider: dest,
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
