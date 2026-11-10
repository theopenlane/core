package r2_test

import (
	"bytes"
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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
	assert.NilError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	copier, ok := provider.(storage.Copier)
	assert.Assert(t, ok, "the r2 provider must implement storage.Copier")

	content := []byte(`{"systemInternalID":"entity::copy"}`)
	source := &storagetypes.File{Key: "vendors/copy me.json"}

	_, err = provider.Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{FileName: source.Key, ContentType: "application/json"})
	assert.NilError(t, err)

	destination := "vendors/processed/copy me.json"
	assert.NilError(t, copier.Copy(ctx, source, &storage.CopyOptions{Key: destination}))

	copied, err := provider.Download(ctx, &storagetypes.File{Key: destination}, &storagetypes.DownloadFileOptions{})
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(content, copied.File))

	exists, err := provider.Exists(ctx, source)
	assert.NilError(t, err)
	assert.Check(t, exists, "copy must leave the source in place")

	t.Run("missing source fails", func(t *testing.T) {
		err := copier.Copy(ctx, &storagetypes.File{Key: "vendors/missing.json"}, &storage.CopyOptions{Key: "vendors/processed/missing.json"})
		assert.Check(t, err != nil)
	})
}
