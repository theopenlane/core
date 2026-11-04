package gcs_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestProviderCopy(t *testing.T) {
	const content = "copy me"

	source := &storagetypes.File{Key: "vendors/source.json"}

	t.Run("within the provider bucket", func(t *testing.T) {
		server := newFakeServer(t, fakeObject(testBucket, source.Key, content))
		provider := newFakeProvider(t, server)

		assert.NilError(t, provider.Copy(context.Background(), source, &storage.CopyOptions{Key: "vendors/processed/source.json"}))

		copied, err := provider.Download(context.Background(), &storagetypes.File{Key: "vendors/processed/source.json"}, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]byte(content), copied.File))

		exists, err := provider.Exists(context.Background(), source)
		assert.NilError(t, err)
		assert.Check(t, exists, "copy must leave the source in place")
	})

	t.Run("into another bucket", func(t *testing.T) {
		server := newFakeServer(t, fakeObject(testBucket, source.Key, content))
		provider := newFakeProvider(t, server)

		assert.NilError(t, provider.Copy(context.Background(), source, &storage.CopyOptions{Bucket: otherBucket, Key: "archive/source.json"}))

		copied, err := provider.Download(context.Background(), &storagetypes.File{Bucket: otherBucket, Key: "archive/source.json"}, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]byte(content), copied.File))
	})

	t.Run("source in another bucket", func(t *testing.T) {
		server := newFakeServer(t, fakeObject(otherBucket, source.Key, content))
		provider := newFakeProvider(t, server)

		assert.NilError(t, provider.Copy(context.Background(), &storagetypes.File{Bucket: otherBucket, Key: source.Key}, &storage.CopyOptions{Key: "imported/source.json"}))

		copied, err := provider.Download(context.Background(), &storagetypes.File{Key: "imported/source.json"}, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]byte(content), copied.File))
	})

	t.Run("missing source fails", func(t *testing.T) {
		provider := newFakeProvider(t, newFakeServer(t))

		assert.Check(t, provider.Copy(context.Background(), source, &storage.CopyOptions{Key: "vendors/processed/source.json"}) != nil)
	})
}
