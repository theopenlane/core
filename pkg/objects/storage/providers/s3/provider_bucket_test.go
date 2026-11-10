package s3_test

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
	s3provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/s3"
)

const legacyBucket = "ol-trust-center"

func newMinioS3Provider(t *testing.T, ctx context.Context, endpoint, bucket string) storage.Provider {
	t.Helper()

	credentials := storage.ProviderCredentials{AccessKeyID: minioUser, SecretAccessKey: minioSecret}
	options := storage.NewProviderOptions(
		storage.WithBucket(bucket),
		storage.WithEndpoint(endpoint),
		storage.WithRegion(testRegion),
		storage.WithCredentials(credentials),
	)

	provider, err := s3provider.NewS3Builder().WithOptions(s3provider.WithUsePathStyle(true)).Build(ctx, credentials, options)
	assert.NilError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	return provider
}

// TestS3ProviderReadsFromStoredBucket persists a file through a provider configured with one bucket, then reads it
// back through a provider configured with a different bucket using only the metadata persisted at upload time
func TestS3ProviderReadsFromStoredBucket(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, legacyBucket)
	createBucket(t, ctx, endpoint, minioBucket)

	content := []byte("evidence persisted before the bucket rename")

	uploaded, err := newMinioS3Provider(t, ctx, endpoint, legacyBucket).Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:          "SOC 2 Logo.pdf",
		FolderDestination: "org/evidence",
		ContentType:       "application/pdf",
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(legacyBucket, uploaded.Bucket))
	assert.Check(t, is.Equal("s3://"+legacyBucket+"/"+uploaded.Key, uploaded.FullURI))

	provider := newMinioS3Provider(t, ctx, endpoint, minioBucket)

	persisted := &storagetypes.File{
		Key:         uploaded.Key,
		Bucket:      uploaded.Bucket,
		Region:      uploaded.Region,
		FullURI:     uploaded.FullURI,
		ContentType: uploaded.ContentType,
	}

	t.Run("download", func(t *testing.T) {
		downloaded, err := provider.Download(ctx, persisted, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(content, downloaded.File))
	})

	t.Run("download with nil options", func(t *testing.T) {
		downloaded, err := provider.Download(ctx, persisted, nil)
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
