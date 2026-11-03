package disk_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	diskprovider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/disk"
)

func diskOptions(bucket string, localURL string) *storage.ProviderOptions {
	options := storage.NewProviderOptions(storage.WithBucket(bucket))
	if localURL != "" {
		options.Apply(storage.WithLocalURL(localURL))
	}
	return options
}

func TestNewDiskProvider(t *testing.T) {
	tests := []struct {
		name        string
		options     *storage.ProviderOptions
		expectError bool
		errContains string
	}{
		{
			name:    "valid configuration",
			options: diskOptions("/tmp/test-storage", ""),
		},
		{
			name:    "valid configuration with local URL",
			options: diskOptions("/tmp/test-storage", "http://localhost:8080/files"),
		},
		{
			name:        "nil options",
			options:     nil,
			expectError: true,
			errContains: "invalid folder path",
		},
		{
			name:        "empty bucket",
			options:     diskOptions("", ""),
			expectError: true,
			errContains: "invalid folder path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := diskprovider.NewDiskProvider(tt.options)
			if tt.expectError {
				assert.Check(t, err != nil)
				if tt.errContains != "" {
					assert.Check(t, is.Contains(err.Error(), tt.errContains))
				}
				assert.Check(t, is.Nil(provider))
			} else {
				assert.Check(t, err)
				assert.Check(t, provider != nil)
				assert.Check(t, provider.Close())
			}
		})
	}
}

func TestDiskProviderUploadDownloadFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	assert.NilError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, ""))
	assert.NilError(t, err)
	defer provider.Close()

	ctx := context.Background()
	content := "This is test file content for disk provider"
	uploadOpts := &storagetypes.UploadFileOptions{
		FileName:    "test-file.txt",
		ContentType: "text/plain",
	}

	reader := strings.NewReader(content)
	uploadResult, err := provider.Upload(ctx, reader, uploadOpts)
	assert.NilError(t, err)
	assert.Assert(t, uploadResult != nil)
	assert.Check(t, is.Equal(uploadOpts.FileName, uploadResult.Key))

	fullPath := filepath.Join(tempDir, uploadResult.Key)
	_, err = os.Stat(fullPath)
	assert.Check(t, err)

	downloadResult, err := provider.Download(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploadResult.Key}}, &storagetypes.DownloadFileOptions{})
	assert.NilError(t, err)
	assert.Assert(t, downloadResult != nil)
	assert.Check(t, is.DeepEqual([]byte(content), downloadResult.File))

	exists, err := provider.Exists(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploadResult.Key}})
	assert.Check(t, err)
	assert.Check(t, exists)

	err = provider.Delete(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploadResult.Key}}, &storagetypes.DeleteFileOptions{})
	assert.Check(t, err)

	_, err = os.Stat(fullPath)
	assert.Check(t, os.IsNotExist(err))
}

func TestDiskProviderGetPresignedURL(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	assert.NilError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, "http://localhost:8080/files"))
	assert.NilError(t, err)
	defer provider.Close()

	file := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "test-file.txt"}}
	url, err := provider.GetPresignedURL(context.Background(), file, &storagetypes.PresignedURLOptions{Duration: time.Hour})
	assert.NilError(t, err)
	assert.Check(t, is.Contains(url, "http://localhost:8080/files/test-file.txt"))
}

func TestDiskProviderListObjects(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	assert.NilError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, ""))
	assert.NilError(t, err)
	defer provider.Close()

	ctx := context.Background()
	uploads := []struct {
		folder string
		name   string
	}{
		{folder: "imports", name: "a.json"},
		{folder: "imports", name: "b.json"},
		{folder: "exports", name: "c.json"},
	}

	for _, upload := range uploads {
		_, err := provider.Upload(ctx, strings.NewReader("{}"), &storagetypes.UploadFileOptions{
			FileName:          upload.name,
			FolderDestination: upload.folder,
			ContentType:       "application/json",
		})
		assert.NilError(t, err)
	}

	keys, err := provider.ListObjects(ctx, "imports/", 0)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]string{"imports/a.json", "imports/b.json"}, slices.Sorted(slices.Values(keys))))

	keys, err = provider.ListObjects(ctx, "", 2)
	assert.NilError(t, err)
	assert.Check(t, is.Len(keys, 2))

	keys, err = provider.ListObjects(ctx, "missing/", 0)
	assert.NilError(t, err)
	assert.Check(t, is.Len(keys, 0))
}

func TestDiskProviderExistsMissingFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	assert.NilError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, ""))
	assert.NilError(t, err)
	defer provider.Close()

	exists, err := provider.Exists(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "missing.txt"}})
	assert.Check(t, err)
	assert.Check(t, !exists)
}
