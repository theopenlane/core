package r2_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	r2provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/r2"
)

func TestR2ProviderCopy(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	credentials := storage.ProviderCredentials{AccountID: "test-account", AccessKeyID: minioUser, SecretAccessKey: minioSecret}
	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(credentials),
	)

	provider, err := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true)).Build(ctx, credentials, options)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	copier, ok := provider.(storage.Copier)
	require.True(t, ok, "the r2 provider must implement storage.Copier")

	content := []byte(`{"systemInternalID":"entity::copy"}`)
	source := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "vendors/copy me.json"}}

	_, err = provider.Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{FileName: source.Key, ContentType: "application/json"})
	require.NoError(t, err)

	destination := "vendors/processed/copy me.json"
	require.NoError(t, copier.Copy(ctx, source, &storage.CopyOptions{Key: destination}))

	copied, err := provider.Download(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: destination}}, &storagetypes.DownloadFileOptions{})
	require.NoError(t, err)
	assert.Equal(t, content, copied.File)

	exists, err := provider.Exists(ctx, source)
	require.NoError(t, err)
	assert.True(t, exists, "copy must leave the source in place")

	t.Run("missing source fails", func(t *testing.T) {
		err := copier.Copy(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "vendors/missing.json"}}, &storage.CopyOptions{Key: "vendors/processed/missing.json"})
		assert.Error(t, err)
	})
}
