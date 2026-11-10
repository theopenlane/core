package gcs_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/option"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/gcs"
)

// TestProviderReadsFromStoredBucket persists a file through a provider configured with one bucket, then reads it
// back through a provider configured with a different bucket using only the metadata persisted at upload time
func TestProviderReadsFromStoredBucket(t *testing.T) {
	ctx := context.Background()
	server := newFakeServer(t)
	content := "evidence persisted before the bucket rename"

	uploaded, err := newFakeProvider(t, server, storage.WithBucket(otherBucket)).Upload(ctx, strings.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:          "SOC 2 Logo.pdf",
		FolderDestination: "org/evidence",
		ContentType:       "application/pdf",
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(otherBucket, uploaded.Bucket))
	assert.Check(t, is.Equal("gs://"+otherBucket+"/"+uploaded.Key, uploaded.FullURI))

	provider := newFakeProvider(t, server)

	persisted := &storagetypes.File{
		Key:         uploaded.Key,
		Bucket:      uploaded.Bucket,
		FullURI:     uploaded.FullURI,
		ContentType: uploaded.ContentType,
	}

	t.Run("download", func(t *testing.T) {
		downloaded, err := provider.Download(ctx, persisted, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(content, string(downloaded.File)))
	})

	t.Run("exists", func(t *testing.T) {
		exists, err := provider.Exists(ctx, persisted)
		assert.NilError(t, err)
		assert.Check(t, exists)
	})

	t.Run("presigned url", func(t *testing.T) {
		signer, err := gcs.NewProvider(ctx, gcsOptions(), fakeClientOptions(server, option.WithCredentialsJSON(serviceAccountJSON(t))))
		assert.NilError(t, err)
		t.Cleanup(func() { _ = signer.Close() })

		signed, err := signer.GetPresignedURL(ctx, persisted, &storagetypes.PresignedURLOptions{Duration: time.Minute})
		assert.NilError(t, err)

		parsed, err := url.Parse(signed)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("storage.googleapis.com", parsed.Host))
		assert.Check(t, is.Equal("/"+otherBucket+"/"+uploaded.Key, parsed.Path))
	})

	t.Run("bucket parsed from uri", func(t *testing.T) {
		fromURI := &storagetypes.File{Key: uploaded.Key, FullURI: uploaded.FullURI}

		downloaded, err := provider.Download(ctx, fromURI, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(content, string(downloaded.File)))
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
