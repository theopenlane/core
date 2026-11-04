package s3_test

import (
	"bytes"
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	s3provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/s3"
)

func TestCopySource(t *testing.T) {
	tests := []struct {
		name     string
		bucket   string
		key      string
		expected string
	}{
		{name: "plain key", bucket: "bucket", key: "object.json", expected: "bucket/object.json"},
		{name: "nested key keeps separators", bucket: "bucket", key: "vendors/processed/object.json", expected: "bucket/vendors/processed/object.json"},
		{name: "reserved characters are escaped per segment", bucket: "bucket", key: "vendors/a b+c?.json", expected: "bucket/vendors/a%20b+c%3F.json"},
		{name: "bucket is never escaped", bucket: "my bucket", key: "x", expected: "my bucket/x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, is.Equal(tt.expected, s3provider.CopySource(tt.bucket, tt.key)))
		})
	}
}

func TestS3ProviderCopy(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithRegion(testRegion),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{AccessKeyID: minioUser, SecretAccessKey: minioSecret}),
	)

	provider, err := s3provider.NewS3Builder().WithOptions(s3provider.WithUsePathStyle(true)).Build(ctx, storage.ProviderCredentials{AccessKeyID: minioUser, SecretAccessKey: minioSecret}, options)
	assert.NilError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	copier, ok := provider.(storage.Copier)
	assert.Assert(t, ok, "the s3 provider must implement storage.Copier")

	content := []byte(`[{"systemInternalID":"entity::copy"}]`)
	source := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "vendors/copy me.json"}}

	_, err = provider.Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{FileName: source.Key, ContentType: "application/json"})
	assert.NilError(t, err)

	destination := "vendors/processed/copy me.json"
	assert.NilError(t, copier.Copy(ctx, source, &storage.CopyOptions{Key: destination}))

	copied, err := provider.Download(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: destination}}, &storagetypes.DownloadFileOptions{})
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(content, copied.File))

	exists, err := provider.Exists(ctx, source)
	assert.NilError(t, err)
	assert.Check(t, exists, "copy must leave the source in place")

	assert.NilError(t, provider.Delete(ctx, source, &storagetypes.DeleteFileOptions{}))

	exists, err = provider.Exists(ctx, source)
	assert.NilError(t, err)
	assert.Check(t, !exists)

	t.Run("missing source fails", func(t *testing.T) {
		err := copier.Copy(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "vendors/missing.json"}}, &storage.CopyOptions{Key: "vendors/processed/missing.json"})
		assert.Check(t, err != nil)
	})
}
