package disk_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	diskprovider "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewDiskProvider(t *testing.T) {
	tests := []struct {
		name          string
		config        *diskprovider.Config
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration with explicit base path",
			config: &diskprovider.Config{
				BasePath: "/tmp/test-storage",
			},
			expectError: false,
		},
		{
			name: "valid configuration with custom URL",
			config: &diskprovider.Config{
				BasePath: "/tmp/test-storage",
				LocalURL: "file://localhost",
			},
			expectError: false,
		},
		{
			name:          "nil configuration",
			config:        nil,
			expectError:   true, // Nil config will cause panic or error
			errorContains: "path",
		},
		{
			name: "empty base path",
			config: &diskprovider.Config{
				BasePath: "", // Empty path should cause error
			},
			expectError:   true,
			errorContains: "path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := diskprovider.NewDiskProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)

				// Test that we can close the provider
				err = provider.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiskProviderUploadDownloadFlow(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := &diskprovider.Config{
		BasePath: tempDir,
	}

	provider, err := diskprovider.NewDiskProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	ctx := context.Background()
	testContent := "This is test file content for disk provider"

	// Test Upload
	uploadOpts := &storagetypes.UploadFileOptions{
		FileName:    "test-file.txt",
		ContentType: "text/plain",
		Metadata: map[string]string{
			"test_key": "test_value",
		},
	}

	reader := strings.NewReader(testContent)
	uploadResult, err := provider.Upload(ctx, reader, uploadOpts)
	require.NoError(t, err)
	require.NotNil(t, uploadResult)

	assert.NotEmpty(t, uploadResult.Key)
	assert.Equal(t, int64(len(testContent)), uploadResult.Size)
	assert.NotEmpty(t, uploadResult.FolderDestination)

	// Verify file exists on disk
	fullPath := filepath.Join(tempDir, uploadResult.Key)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err, "Uploaded file should exist on disk")

	// Test Download
	downloadOpts := &storagetypes.DownloadFileOptions{
		FileName: uploadResult.Key,
	}

	downloadResult, err := provider.Download(ctx, downloadOpts)
	require.NoError(t, err)
	require.NotNil(t, downloadResult)

	assert.Equal(t, []byte(testContent), downloadResult.File)
	assert.Equal(t, int64(len(testContent)), downloadResult.Size)

	// Test Exists
	exists, err := provider.Exists(ctx, uploadResult.Key)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existent file
	exists, err = provider.Exists(ctx, "non-existent-file.txt")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test Delete
	err = provider.Delete(ctx, uploadResult.Key)
	assert.NoError(t, err)

	// Verify file is deleted
	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err), "File should be deleted from disk")

	// Test exists after delete
	exists, err = provider.Exists(ctx, uploadResult.Key)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestDiskProviderGetPresignedURL(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := &diskprovider.Config{
		BasePath: tempDir,
	}

	provider, err := diskprovider.NewDiskProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	// Disk provider typically doesn't support presigned URLs in the traditional sense
	// This test verifies the behavior (might return local file path or error)
	url, err := provider.GetPresignedURL("test-file.txt", 1*time.Hour)

	// Depending on implementation, this might return an error or a local file path
	if err != nil {
		assert.Contains(t, err.Error(), "presigned")
	} else {
		assert.NotEmpty(t, url)
	}
}

func TestDiskProviderGetScheme(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-scheme-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	provider, err := diskprovider.NewDiskProvider(&diskprovider.Config{
		BasePath: tempDir,
	})
	require.NoError(t, err)
	defer provider.Close()

	scheme := provider.GetScheme()
	assert.NotNil(t, scheme)
	assert.Equal(t, "file://", *scheme)
}

func TestDiskProviderErrorCases(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "disk-provider-error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := &diskprovider.Config{
		BasePath: tempDir,
	}

	provider, err := diskprovider.NewDiskProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	ctx := context.Background()

	tests := []struct {
		name        string
		operation   func() error
		expectError bool
	}{
		{
			name: "download non-existent file",
			operation: func() error {
				_, err := provider.Download(ctx, &storagetypes.DownloadFileOptions{
					FileName: "non-existent-file.txt",
				})
				return err
			},
			expectError: true,
		},
		{
			name: "delete non-existent file",
			operation: func() error {
				return provider.Delete(ctx, "non-existent-file.txt")
			},
			expectError: false, // Disk provider doesn't error on deleting non-existent files
		},
		{
			name: "upload with empty key",
			operation: func() error {
				_, err := provider.Upload(ctx, strings.NewReader("test"), &storagetypes.UploadFileOptions{
					FileName: "",
				})
				return err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiskConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *diskprovider.Config
		expected    string
		expectError bool
	}{
		{
			name: "explicit directory",
			config: &diskprovider.Config{
				BasePath: "/tmp/custom-storage",
			},
			expected:    "/tmp/custom-storage",
			expectError: false,
		},
		{
			name:        "nil config returns error",
			config:      nil,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := diskprovider.NewDiskProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)

				if provider != nil {
					err = provider.Close()
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestDiskProviderConcurrentOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "disk-provider-concurrent-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := &diskprovider.Config{
		BasePath: tempDir,
	}

	provider, err := diskprovider.NewDiskProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	ctx := context.Background()
	numGoroutines := 10

	// Test concurrent uploads
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			content := fmt.Sprintf("test content %d", id)
			uploadOpts := &storagetypes.UploadFileOptions{
				FileName:    fmt.Sprintf("test-file-%d.txt", id),
				ContentType: "text/plain",
			}

			_, err := provider.Upload(ctx, strings.NewReader(content), uploadOpts)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all files were created
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), numGoroutines)
}
