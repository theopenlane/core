package storagetypes_test

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
)

func TestFileMetadata_StoredBucket(t *testing.T) {
	t.Run("bucket recorded", func(t *testing.T) {
		m := storagetypes.FileMetadata{Bucket: "ol-trust-center", FullURI: "r2://opln/org/file.pdf", Key: "org/file.pdf"}
		assert.Check(t, is.Equal("ol-trust-center", m.StoredBucket()))
	})

	t.Run("parsed from uri with key suffix", func(t *testing.T) {
		m := storagetypes.FileMetadata{FullURI: "s3://ol-trust-center/org/obj/file.pdf", Key: "org/obj/file.pdf"}
		assert.Check(t, is.Equal("ol-trust-center", m.StoredBucket()))
	})

	t.Run("parsed from disk uri with absolute directory", func(t *testing.T) {
		m := storagetypes.FileMetadata{FullURI: "file:///var/lib/openlane/files/org/file.pdf", Key: "org/file.pdf"}
		assert.Check(t, is.Equal("/var/lib/openlane/files", m.StoredBucket()))
	})

	t.Run("parsed from uri without key", func(t *testing.T) {
		m := storagetypes.FileMetadata{FullURI: "gs://evidence-bucket/org/file.pdf"}
		assert.Check(t, is.Equal("evidence-bucket", m.StoredBucket()))
	})

	t.Run("parsed from uri when key does not match", func(t *testing.T) {
		m := storagetypes.FileMetadata{FullURI: "s3://evidence-bucket/org/file.pdf", Key: "other/file.pdf"}
		assert.Check(t, is.Equal("evidence-bucket", m.StoredBucket()))
	})

	t.Run("no bucket and no uri", func(t *testing.T) {
		m := storagetypes.FileMetadata{Key: "org/file.pdf"}
		assert.Check(t, is.Len(m.StoredBucket(), 0))
	})

	t.Run("uri without scheme", func(t *testing.T) {
		m := storagetypes.FileMetadata{FullURI: "org/file.pdf", Key: "org/file.pdf"}
		assert.Check(t, is.Len(m.StoredBucket(), 0))
	})
}
