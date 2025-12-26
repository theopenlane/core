package disk_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
	diskprovider "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
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
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.NoError(t, provider.Close())
			}
		})
	}
}

func TestDiskProviderUploadDownloadFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, ""))
	require.NoError(t, err)
	defer provider.Close()

	ctx := context.Background()
	content := "This is test file content for disk provider"
	uploadOpts := &storagetypes.UploadFileOptions{
		FileName:    "test-file.txt",
		ContentType: "text/plain",
	}

	reader := strings.NewReader(content)
	uploadResult, err := provider.Upload(ctx, reader, uploadOpts)
	require.NoError(t, err)
	require.NotNil(t, uploadResult)
	assert.Equal(t, uploadOpts.FileName, uploadResult.Key)

	fullPath := filepath.Join(tempDir, uploadResult.Key)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)

	downloadResult, err := provider.Download(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploadResult.Key}}, &storagetypes.DownloadFileOptions{})
	require.NoError(t, err)
	require.NotNil(t, downloadResult)
	assert.Equal(t, []byte(content), downloadResult.File)

	exists, err := provider.Exists(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploadResult.Key}})
	assert.NoError(t, err)
	assert.True(t, exists)

	err = provider.Delete(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploadResult.Key}}, &storagetypes.DeleteFileOptions{})
	assert.NoError(t, err)

	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDiskProviderGetPresignedURL(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, "http://localhost:8080/files"))
	require.NoError(t, err)
	defer provider.Close()

	file := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "test-file.txt"}}
	url, err := provider.GetPresignedURL(context.Background(), file, &storagetypes.PresignedURLOptions{Duration: time.Hour})
	require.NoError(t, err)
	assert.Contains(t, url, "http://localhost:8080/files/test-file.txt")
}

func TestDiskProviderExistsMissingFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(diskOptions(tempDir, ""))
	require.NoError(t, err)
	defer provider.Close()

	exists, err := provider.Exists(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "missing.txt"}})
	assert.NoError(t, err)
	assert.False(t, exists)
}
