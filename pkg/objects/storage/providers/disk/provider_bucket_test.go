package disk_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	diskprovider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/disk"
)

// TestDiskProviderReadsFromStoredBucket persists a file through a provider configured with one directory, then reads
// it back through a provider configured with a different directory using only the metadata persisted at upload time
func TestDiskProviderReadsFromStoredBucket(t *testing.T) {
	ctx := context.Background()
	legacyDir := t.TempDir()
	currentDir := t.TempDir()
	content := "evidence persisted before the directory change"

	legacy, err := diskprovider.NewDiskProvider(diskOptions(legacyDir, ""))
	assert.NilError(t, err)
	t.Cleanup(func() { _ = legacy.Close() })

	uploaded, err := legacy.Upload(ctx, strings.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:          "SOC 2 Logo.pdf",
		FolderDestination: "org/evidence",
		ContentType:       "application/pdf",
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(legacyDir, uploaded.Bucket))
	assert.Check(t, is.Equal("file://"+filepath.ToSlash(filepath.Join(legacyDir, uploaded.Key)), uploaded.FullURI))

	provider, err := diskprovider.NewDiskProvider(diskOptions(currentDir, ""))
	assert.NilError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

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

	t.Run("bucket parsed from uri", func(t *testing.T) {
		fromURI := &storagetypes.File{Key: uploaded.Key, FullURI: uploaded.FullURI}

		downloaded, err := provider.Download(ctx, fromURI, &storagetypes.DownloadFileOptions{})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(content, string(downloaded.File)))
	})

	t.Run("configured directory when nothing is stored", func(t *testing.T) {
		exists, err := provider.Exists(ctx, &storagetypes.File{Key: uploaded.Key})
		assert.NilError(t, err)
		assert.Check(t, !exists)
	})

	t.Run("delete", func(t *testing.T) {
		assert.NilError(t, provider.Delete(ctx, persisted, &storagetypes.DeleteFileOptions{}))

		_, err := os.Stat(filepath.Join(legacyDir, uploaded.Key))
		assert.Check(t, os.IsNotExist(err))
	})
}
