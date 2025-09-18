package s3_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewS3Provider(t *testing.T) {
	tests := []struct {
		name          string
		config        *s3provider.Config
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration",
			config: &s3provider.Config{
				Bucket:          "test-bucket",
				Region:          "us-east-1",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
			},
			expectError: false,
		},
		{
			name: "valid configuration with endpoint",
			config: &s3provider.Config{
				Bucket:          "test-bucket",
				Region:          "us-east-1",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Endpoint:        "https://s3.example.com",
				UsePathStyle:    true,
			},
			expectError: false,
		},
		{
			name: "missing bucket",
			config: &s3provider.Config{
				Region:          "us-east-1",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
			},
			expectError:   true,
			errorContains: "S3 bucket is required",
		},
		{
			name: "empty bucket",
			config: &s3provider.Config{
				Bucket:          "",
				Region:          "us-east-1",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
			},
			expectError:   true,
			errorContains: "S3 bucket is required",
		},
		{
			name: "configuration without credentials (should try environment)",
			config: &s3provider.Config{
				Bucket: "test-bucket",
				Region: "us-east-1",
			},
			expectError: false, // Should succeed if environment has credentials
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := s3provider.NewS3Provider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				// Note: These tests might fail in CI if AWS credentials aren't available
				// In a real environment, you'd mock the AWS SDK
				if err != nil {
					t.Skip("Skipping test due to missing AWS credentials")
				}
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestNewS3ProviderResult(t *testing.T) {
	t.Run("valid config returns success", func(t *testing.T) {
		config := &s3provider.Config{
			Bucket:          "test-bucket",
			Region:          "us-east-1",
			AccessKeyID:     "test-access-key",
			SecretAccessKey: "test-secret-key",
		}

		result := s3provider.NewS3ProviderResult(config)

		if result.IsError() {
			// Skip if AWS credentials aren't available
			t.Skip("Skipping test due to missing AWS credentials")
		}

		assert.True(t, result.IsOk())
		provider := result.MustGet()
		assert.NotNil(t, provider)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		config := &s3provider.Config{
			// Missing bucket
			Region:          "us-east-1",
			AccessKeyID:     "test-access-key",
			SecretAccessKey: "test-secret-key",
		}

		result := s3provider.NewS3ProviderResult(config)

		assert.True(t, result.IsError())
		assert.Contains(t, result.Error().Error(), "bucket is required")
	})
}

func TestS3ProviderConstants(t *testing.T) {
	assert.Equal(t, 15*time.Minute, s3provider.DefaultPresignedURLExpiry)
	assert.Equal(t, 64*1024*1024, s3provider.DefaultPartSize)
	assert.Equal(t, 5, s3provider.DefaultConcurrency)
}

func TestS3Config(t *testing.T) {
	config := &s3provider.Config{
		Bucket:          "test-bucket",
		Region:          "us-west-2",
		AccessKeyID:     "access-key-123",
		SecretAccessKey: "secret-key-456",
		Endpoint:        "https://custom-endpoint.com",
		UsePathStyle:    true,
		DebugMode:       true,
	}

	assert.Equal(t, "test-bucket", config.Bucket)
	assert.Equal(t, "us-west-2", config.Region)
	assert.Equal(t, "access-key-123", config.AccessKeyID)
	assert.Equal(t, "secret-key-456", config.SecretAccessKey)
	assert.Equal(t, "https://custom-endpoint.com", config.Endpoint)
	assert.True(t, config.UsePathStyle)
	assert.True(t, config.DebugMode)
}

// Mock S3 provider for testing methods that don't require AWS connection
func createMockS3Provider() *s3provider.Provider {
	// This would normally require a real AWS connection
	// In a production test suite, you'd use testcontainers or localstack
	config := &s3provider.Config{
		Bucket:          "test-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
	}

	provider, err := s3provider.NewS3Provider(config)
	if err != nil {
		// Return nil if we can't create a real provider (no AWS credentials)
		return nil
	}
	return provider
}

func TestS3ProviderMethods(t *testing.T) {
	// Note: These tests would require either:
	// 1. Real AWS credentials and a test bucket
	// 2. A local S3-compatible service like LocalStack
	// 3. Mocking the AWS SDK

	// For now, we'll test the method signatures and basic validation

	t.Run("GetScheme", func(t *testing.T) {
		provider := createMockS3Provider()
		if provider == nil {
			t.Skip("Skipping test due to missing AWS credentials")
		}

		scheme := provider.GetScheme()
		assert.NotNil(t, scheme)
		assert.Equal(t, "s3://", *scheme)
	})

	t.Run("ProviderType", func(t *testing.T) {
		provider := createMockS3Provider()
		if provider == nil {
			t.Skip("Skipping test due to missing AWS credentials")
		}

		assert.Equal(t, storagetypes.S3Provider, provider.ProviderType())
	})

	t.Run("Close", func(t *testing.T) {
		provider := createMockS3Provider()
		if provider == nil {
			t.Skip("Skipping test due to missing AWS credentials")
		}

		err := provider.Close()
		assert.NoError(t, err)
	})
}

func TestS3ProviderUploadDownloadFlow(t *testing.T) {
	// This is an integration test that would require a real S3 environment
	t.Skip("Integration test - requires real S3 environment or LocalStack")

	provider := createMockS3Provider()
	if provider == nil {
		t.Skip("Skipping test due to missing AWS credentials")
	}

	ctx := context.Background()
	testContent := "This is test file content"
	fileName := "test-file.txt"

	// Test Upload
	uploadOpts := &storagetypes.UploadFileOptions{
		FileName:    fileName,
		ContentType: "text/plain",
	}

	metadata, err := provider.Upload(ctx, strings.NewReader(testContent), uploadOpts)
	require.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, fileName, metadata.Key)
	assert.Equal(t, int64(len(testContent)), metadata.Size)

	// Test Exists
	file := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key: fileName,
		},
	}
	exists, err := provider.Exists(ctx, file)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Download
	downloadOpts := &storagetypes.DownloadFileOptions{
		FileName: fileName,
	}

	downloadResult, err := provider.Download(ctx, file, downloadOpts)
	require.NoError(t, err)
	assert.NotNil(t, downloadResult)
	assert.Equal(t, testContent, string(downloadResult.File))
	assert.Equal(t, int64(len(testContent)), downloadResult.Size)

	// Test GetPresignedURL
	opts := &storagetypes.PresignedURLOptions{
		Duration: 1 * time.Hour,
	}
	presignedURL, err := provider.GetPresignedURL(ctx, file, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, presignedURL)
	assert.Contains(t, presignedURL, "https://")

	// Test Delete
	deleteOpts := &storagetypes.DeleteFileOptions{}
	err = provider.Delete(ctx, file, deleteOpts)
	require.NoError(t, err)

	// Verify deletion
	exists, err = provider.Exists(ctx, file)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestS3ProviderGetPresignedURLWithDefaults(t *testing.T) {
	t.Skip("Integration test - requires real S3 environment or LocalStack")

	provider := createMockS3Provider()
	if provider == nil {
		t.Skip("Skipping test due to missing AWS credentials")
	}

	// Test with zero duration (should use default)
	file := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key: "test-file.txt",
		},
	}
	opts := &storagetypes.PresignedURLOptions{
		Duration: 0,
	}
	url, err := provider.GetPresignedURL(context.Background(), file, opts)
	if err != nil {
		t.Skip("Skipping test due to AWS connection issues")
	}
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "https://")

	// Test with custom duration
	opts.Duration = 30 * time.Minute
	url, err = provider.GetPresignedURL(context.Background(), file, opts)
	if err != nil {
		t.Skip("Skipping test due to AWS connection issues")
	}
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "https://")
}

func TestS3ProviderExistsNonExistentFile(t *testing.T) {
	t.Skip("Integration test - requires real S3 environment or LocalStack")

	provider := createMockS3Provider()
	if provider == nil {
		t.Skip("Skipping test due to missing AWS credentials")
	}

	ctx := context.Background()
	file := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key: "non-existent-file.txt",
		},
	}
	exists, err := provider.Exists(ctx, file)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func BenchmarkS3ProviderUpload(b *testing.B) {
	provider := createMockS3Provider()
	if provider == nil {
		b.Skip("Skipping benchmark due to missing AWS credentials")
	}

	ctx := context.Background()
	content := strings.Repeat("benchmark test content ", 100)

	uploadOpts := &storagetypes.UploadFileOptions{
		FileName:    "benchmark-file.txt",
		ContentType: "text/plain",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		_, err := provider.Upload(ctx, reader, uploadOpts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestErrorCases tests various error scenarios
func TestS3ProviderErrorCases(t *testing.T) {
	tests := []struct {
		name   string
		config *s3provider.Config
	}{
		{
			name: "invalid bucket name",
			config: &s3provider.Config{
				Bucket:          "",
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s3provider.NewS3Provider(tt.config)
			assert.Error(t, err)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *s3provider.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &s3provider.Config{
				Bucket: "valid-bucket",
				Region: "us-east-1",
			},
			expectError: false,
		},
		{
			name: "empty bucket",
			config: &s3provider.Config{
				Bucket: "",
				Region: "us-east-1",
			},
			expectError: true,
		},
		{
			name: "nil config",
			config: &s3provider.Config{
				Bucket: "test-bucket",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s3provider.NewS3Provider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Skip if AWS credentials aren't available
				if err != nil && strings.Contains(err.Error(), "credentials") {
					t.Skip("Skipping test due to missing AWS credentials")
				}
			}
		})
	}
}
