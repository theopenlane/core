package disk_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	diskprovider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/disk"
)

func TestDiskProviderCopy(t *testing.T) {
	tempDir := t.TempDir()

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, ""))
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	ctx := context.Background()
	content := "copy me"
	source := &storagetypes.File{Key: "vendors/source.json"}

	_, err = provider.Upload(ctx, strings.NewReader(content), &storagetypes.UploadFileOptions{FileName: source.Key, ContentType: "application/json"})
	require.NoError(t, err)

	t.Run("creates the destination directories and keeps the source", func(t *testing.T) {
		destination := "vendors/processed/nested/source.json"
		require.NoError(t, provider.Copy(ctx, source, &storage.CopyOptions{Key: destination}))

		copied, err := os.ReadFile(filepath.Join(tempDir, destination))
		require.NoError(t, err)
		assert.Equal(t, content, string(copied))

		_, err = os.Stat(filepath.Join(tempDir, source.Key))
		assert.NoError(t, err)
	})

	t.Run("into another bucket directory", func(t *testing.T) {
		otherDir := t.TempDir()
		require.NoError(t, provider.Copy(ctx, source, &storage.CopyOptions{Bucket: otherDir, Key: "archive/source.json"}))

		copied, err := os.ReadFile(filepath.Join(otherDir, "archive/source.json"))
		require.NoError(t, err)
		assert.Equal(t, content, string(copied))
	})

	t.Run("missing source fails", func(t *testing.T) {
		err := provider.Copy(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "vendors/missing.json"}}, &storage.CopyOptions{Key: "vendors/processed/missing.json"})
		assert.True(t, os.IsNotExist(err), "expected a not-exist error, got %v", err)
	})
}
