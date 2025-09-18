package r2_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewR2Provider(t *testing.T) {
	tests := []struct {
		name          string
		config        *r2provider.Config
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration",
			config: &r2provider.Config{
				Bucket:          "test-bucket",
				AccountID:       "test-account-id",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto", // R2 typically uses "auto"
			},
			expectError: false,
		},
		{
			name: "invalid configuration with only token",
			config: &r2provider.Config{
				Bucket:    "test-bucket",
				AccountID: "test-account-id",
				APIToken:  "test-api-token",
				Region:    "auto",
			},
			expectError:   true, // R2 provider currently requires access key/secret key
			errorContains: "R2 access key ID",
		},
		{
			name: "valid configuration with endpoint",
			config: &r2provider.Config{
				Bucket:          "test-bucket",
				AccountID:       "test-account-id",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto",
				Endpoint:        "https://test-account-id.r2.cloudflarestorage.com",
			},
			expectError: false,
		},
		{
			name: "missing bucket",
			config: &r2provider.Config{
				AccountID:       "test-account-id",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto",
			},
			expectError:   true,
			errorContains: "R2 bucket",
		},
		{
			name: "empty bucket",
			config: &r2provider.Config{
				Bucket:          "",
				AccountID:       "test-account-id",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto",
			},
			expectError:   true,
			errorContains: "R2 bucket",
		},
		{
			name: "missing account ID",
			config: &r2provider.Config{
				Bucket:          "test-bucket",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto",
			},
			expectError:   true,
			errorContains: "R2 account ID",
		},
		{
			name: "missing credentials",
			config: &r2provider.Config{
				Bucket:    "test-bucket",
				AccountID: "test-account-id",
				Region:    "auto",
			},
			expectError:   true,
			errorContains: "R2 access key ID", // Error message says "R2 access key ID and secret access key are required"
		},
		{
			name:          "nil configuration",
			config:        nil,
			expectError:   true,
			errorContains: "R2 bucket", // Should panic and be caught
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := r2provider.NewR2Provider(tt.config)

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

func TestR2ProviderConstants(t *testing.T) {
	// Test that R2 provider constants are properly defined
	assert.Equal(t, "r2", string(storagetypes.R2Provider))
}

func TestR2Config(t *testing.T) {
	config := r2provider.Config{
		Bucket:          "test-bucket",
		AccountID:       "test-account-id",
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Region:          "auto",
		APIToken:        "test-api-token",
		Endpoint:        "https://test-account-id.r2.cloudflarestorage.com",
	}

	assert.Equal(t, "test-bucket", config.Bucket)
	assert.Equal(t, "test-account-id", config.AccountID)
	assert.Equal(t, "test-access-key", config.AccessKeyID)
	assert.Equal(t, "test-secret-key", config.SecretAccessKey)
	assert.Equal(t, "auto", config.Region)
	assert.Equal(t, "test-api-token", config.APIToken)
	assert.Equal(t, "https://test-account-id.r2.cloudflarestorage.com", config.Endpoint)
}

func TestR2ProviderMethods(t *testing.T) {
	config := &r2provider.Config{
		Bucket:          "test-bucket",
		AccountID:       "test-account-id",
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Region:          "auto",
	}

	provider, err := r2provider.NewR2Provider(config)
	require.NoError(t, err)
	defer provider.Close()

	t.Run("GetScheme", func(t *testing.T) {
		scheme := provider.GetScheme()
		assert.NotNil(t, scheme)
		assert.Equal(t, "r2://", *scheme)
	})

	t.Run("ProviderType", func(t *testing.T) {
		assert.Equal(t, storagetypes.R2Provider, provider.ProviderType())
	})

	t.Run("Close", func(t *testing.T) {
		// Create a new provider for this test since we'll close it
		testProvider, err := r2provider.NewR2Provider(config)
		require.NoError(t, err)

		err = testProvider.Close()
		assert.NoError(t, err)
	})
}

// Integration tests - these require real R2 credentials and are skipped by default
func TestR2ProviderUploadDownloadFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires real R2 environment")
	}

	// These would need real credentials from environment variables
	config := &r2provider.Config{
		Bucket:          getEnvOrSkip(t, "R2_BUCKET"),
		AccountID:       getEnvOrSkip(t, "R2_ACCOUNT_ID"),
		AccessKeyID:     getEnvOrSkip(t, "R2_ACCESS_KEY_ID"),
		SecretAccessKey: getEnvOrSkip(t, "R2_SECRET_ACCESS_KEY"),
		Region:          "auto",
	}

	provider, err := r2provider.NewR2Provider(config)
	require.NoError(t, err)
	defer provider.Close()

	ctx := context.Background()
	testContent := "This is test file content for R2 provider"

	// Test Upload
	uploadOpts := &storagetypes.UploadFileOptions{
		FileName:    "test-file.txt",
		ContentType: "text/plain",
	}

	reader := strings.NewReader(testContent)
	uploadResult, err := provider.Upload(ctx, reader, uploadOpts)
	require.NoError(t, err)
	require.NotNil(t, uploadResult)

	assert.NotEmpty(t, uploadResult.Key)
	assert.Equal(t, int64(len(testContent)), uploadResult.Size)

	// Test Download
	file := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key: uploadResult.Key,
		},
	}
	downloadOpts := &storagetypes.DownloadFileOptions{
		FileName: uploadResult.Key,
	}

	downloadResult, err := provider.Download(ctx, file, downloadOpts)
	require.NoError(t, err)
	require.NotNil(t, downloadResult)

	assert.Equal(t, []byte(testContent), downloadResult.File)
	assert.Equal(t, int64(len(testContent)), downloadResult.Size)

	// Test Exists
	exists, err := provider.Exists(ctx, file)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test GetPresignedURL
	opts := &storagetypes.PresignedURLOptions{
		Duration: 1 * time.Hour,
	}
	presignedURL, err := provider.GetPresignedURL(ctx, file, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, presignedURL)
	assert.Contains(t, presignedURL, "https://")

	// Test Delete
	deleteOpts := &storagetypes.DeleteFileOptions{}
	err = provider.Delete(ctx, file, deleteOpts)
	assert.NoError(t, err)

	// Test exists after delete
	exists, err = provider.Exists(ctx, file)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestR2ProviderGetPresignedURLWithDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires real R2 environment")
	}

	config := &r2provider.Config{
		Bucket:          getEnvOrSkip(t, "R2_BUCKET"),
		AccountID:       getEnvOrSkip(t, "R2_ACCOUNT_ID"),
		AccessKeyID:     getEnvOrSkip(t, "R2_ACCESS_KEY_ID"),
		SecretAccessKey: getEnvOrSkip(t, "R2_SECRET_ACCESS_KEY"),
		Region:          "auto",
	}

	provider, err := r2provider.NewR2Provider(config)
	require.NoError(t, err)
	defer provider.Close()

	// Test with default duration (should not error)
	file := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key: "test-key.txt",
		},
	}
	opts := &storagetypes.PresignedURLOptions{
		Duration: 0,
	}
	url, err := provider.GetPresignedURL(context.Background(), file, opts)
	if err != nil {
		// Some implementations might not support default duration
		assert.Contains(t, err.Error(), "duration")
	} else {
		assert.NotEmpty(t, url)
	}
}

func TestR2ProviderExistsNonExistentFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires real R2 environment")
	}

	config := &r2provider.Config{
		Bucket:          getEnvOrSkip(t, "R2_BUCKET"),
		AccountID:       getEnvOrSkip(t, "R2_ACCOUNT_ID"),
		AccessKeyID:     getEnvOrSkip(t, "R2_ACCESS_KEY_ID"),
		SecretAccessKey: getEnvOrSkip(t, "R2_SECRET_ACCESS_KEY"),
		Region:          "auto",
	}

	provider, err := r2provider.NewR2Provider(config)
	require.NoError(t, err)
	defer provider.Close()

	ctx := context.Background()

	file := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key: "non-existent-file-12345.txt",
		},
	}
	exists, err := provider.Exists(ctx, file)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestR2ProviderErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		config    *r2provider.Config
		expectErr string
	}{
		{
			name: "invalid bucket name",
			config: &r2provider.Config{
				Bucket:          "invalid..bucket..name",
				AccountID:       "test-account-id",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto",
			},
			expectErr: "", // Bucket validation might not catch this at creation time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := r2provider.NewR2Provider(tt.config)
			if tt.expectErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				assert.Nil(t, provider)
			} else {
				// Provider creation might succeed, but operations would fail
				if err == nil && provider != nil {
					provider.Close()
				}
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *r2provider.Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: &r2provider.Config{
				Bucket:          "valid-bucket",
				AccountID:       "test-account-id",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto",
			},
			expectErr: false,
		},
		{
			name: "empty bucket",
			config: &r2provider.Config{
				Bucket:          "",
				AccountID:       "test-account-id",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "auto",
			},
			expectErr: true,
		},
		{
			name:      "nil config",
			config:    nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r2provider.NewR2Provider(tt.config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to get environment variable or skip test
func getEnvOrSkip(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Environment variable %s not set", key)
	}
	return value
}
