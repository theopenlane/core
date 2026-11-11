package r2_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	r2provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/r2"
)

const legacyBucket = "ol-trust-center"

func newMinioR2Provider(t *testing.T, ctx context.Context, endpoint, bucket string) storage.Provider {
	t.Helper()

	credentials := storage.ProviderCredentials{AccountID: "test-account", AccessKeyID: minioUser, SecretAccessKey: minioSecret}
	options := storage.NewProviderOptions(
		storage.WithBucket(bucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(credentials),
	)

	provider, err := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true)).Build(ctx, credentials, options)
	assert.NilError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	return provider
}

// TestR2ProviderReadsFromStoredBucket persists a file through a provider configured with one bucket, then reads it
// back through a provider configured with a different bucket using only the metadata persisted at upload time
func TestR2ProviderReadsFromStoredBucket(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, legacyBucket)
	createBucket(t, ctx, endpoint, minioBucket)

	content := []byte("evidence persisted before the bucket rename")

	uploaded, err := newMinioR2Provider(t, ctx, endpoint, legacyBucket).Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:          "SOC 2 Logo.pdf",
		FolderDestination: "org/evidence",
		ContentType:       "application/pdf",
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(legacyBucket, uploaded.Bucket))
	assert.Check(t, is.Equal("r2://"+legacyBucket+"/"+uploaded.Key, uploaded.FullURI))

	provider := newMinioR2Provider(t, ctx, endpoint, minioBucket)

	persisted := &storagetypes.File{
		Key:         uploaded.Key,
		Bucket:      uploaded.Bucket,
		FullURI:     uploaded.FullURI,
		ContentType: uploaded.ContentType,
	}

	t.Run("download", func(t *testing.T) {
		downloaded, err := provider.Download(ctx, persisted, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(content, downloaded.File))
	})

	t.Run("exists", func(t *testing.T) {
		exists, err := provider.Exists(ctx, persisted)
		assert.NilError(t, err)
		assert.Check(t, exists)
	})

	t.Run("presigned url", func(t *testing.T) {
		url, err := provider.GetPresignedURL(ctx, persisted, &storagetypes.PresignedURLOptions{Duration: time.Minute})
		assert.NilError(t, err)
		assert.Check(t, is.Contains(url, "/"+legacyBucket+"/"))

		resp, err := http.Get(url)
		assert.NilError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(http.StatusOK, resp.StatusCode))
		assert.Check(t, is.DeepEqual(content, body))
	})

	t.Run("bucket parsed from uri", func(t *testing.T) {
		fromURI := &storagetypes.File{Key: uploaded.Key, FullURI: uploaded.FullURI}

		downloaded, err := provider.Download(ctx, fromURI, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(content, downloaded.File))
	})

	t.Run("configured bucket when nothing is stored", func(t *testing.T) {
		exists, err := provider.Exists(ctx, &storagetypes.File{Key: uploaded.Key})
		assert.NilError(t, err)
		assert.Check(t, !exists)
	})

	t.Run("delete", func(t *testing.T) {
		assert.NilError(t, provider.Delete(ctx, persisted, &storagetypes.DeleteFileOptions{}))

		exists, err := provider.Exists(ctx, persisted)
		assert.NilError(t, err)
		assert.Check(t, !exists)
	})
}
