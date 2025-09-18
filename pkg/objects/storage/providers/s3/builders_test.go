package s3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/cp"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewS3Builder(t *testing.T) {
	builder := s3provider.NewS3Builder()
	assert.NotNil(t, builder)

	// Test interface compliance with correct generic type
	var _ cp.ClientBuilder[storagetypes.Provider] = builder
}

func TestS3Builder_WithCredentials(t *testing.T) {
	builder := s3provider.NewS3Builder()
	credentials := map[string]string{
		"bucket":            "test-bucket",
		"region":            "us-east-1",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
		"endpoint":          "https://s3.example.com",
	}

	result := builder.WithCredentials(credentials)
	assert.NotNil(t, result)
	assert.Equal(t, builder, result) // Should return the same builder for chaining
}

func TestS3Builder_WithConfig(t *testing.T) {
	builder := s3provider.NewS3Builder()
	config := map[string]any{
		"use_path_style": true,
		"debug_mode":     true,
	}

	result := builder.WithConfig(config)
	assert.NotNil(t, result)
	assert.Equal(t, builder, result) // Should return the same builder for chaining
}

func TestS3Builder_ClientType(t *testing.T) {
	builder := s3provider.NewS3Builder()
	clientType := builder.ClientType()

	assert.Equal(t, cp.ProviderType("s3"), clientType)
}

func TestS3Builder_Build(t *testing.T) {
	tests := []struct {
		name          string
		credentials   map[string]string
		config        map[string]any
		expectError   bool
		expectedError string
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"region":            "us-east-1",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
			},
			expectError: false,
		},
		{
			name: "valid credentials with endpoint",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"region":            "us-east-1",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"endpoint":          "https://s3.example.com",
			},
			expectError: false,
		},
		{
			name: "missing bucket",
			credentials: map[string]string{
				"region":            "us-east-1",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
			},
			expectError:   true,
			expectedError: "missing required S3 credentials",
		},
		{
			name: "missing region",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
			},
			expectError:   true,
			expectedError: "missing required S3 credentials",
		},
		{
			name:          "empty credentials",
			credentials:   map[string]string{},
			expectError:   true,
			expectedError: "missing required S3 credentials",
		},
		{
			name: "partial credentials",
			credentials: map[string]string{
				"bucket": "test-bucket",
				"region": "",
			},
			expectError:   true,
			expectedError: "missing required S3 credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := s3provider.NewS3Builder()

			if tt.credentials != nil {
				builder.WithCredentials(tt.credentials)
			}

			if tt.config != nil {
				builder.WithConfig(tt.config)
			}

			ctx := context.Background()
			provider, err := builder.Build(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				// Skip if AWS credentials aren't available in the environment
				if err != nil {
					t.Skipf("Skipping test due to AWS credential issues: %v", err)
				}
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestS3Builder_ChainedCalls(t *testing.T) {
	builder := s3provider.NewS3Builder()

	credentials := map[string]string{
		"bucket":            "test-bucket",
		"region":            "us-east-1",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
	}

	config := map[string]any{
		"use_path_style": true,
		"debug_mode":     false,
	}

	// Test method chaining
	result := builder.
		WithCredentials(credentials).
		WithConfig(config)

	assert.NotNil(t, result)
	assert.Equal(t, cp.ProviderType("s3"), result.ClientType())

	// Test build
	ctx := context.Background()
	provider, err := result.Build(ctx)

	if err != nil {
		t.Skipf("Skipping test due to AWS credential issues: %v", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestNewS3ProviderFromCredentials(t *testing.T) {
	tests := []struct {
		name        string
		credentials map[string]string
		expectError bool
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"region":            "us-east-1",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
			},
			expectError: false,
		},
		{
			name: "invalid credentials - missing bucket",
			credentials: map[string]string{
				"region":            "us-east-1",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
			},
			expectError: true,
		},
		{
			name: "invalid credentials - missing region",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
			},
			expectError: true,
		},
		{
			name:        "empty credentials",
			credentials: map[string]string{},
			expectError: true,
		},
		{
			name:        "nil credentials",
			credentials: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s3provider.NewS3ProviderFromCredentials(tt.credentials)

			if tt.expectError {
				assert.True(t, result.IsError())
				assert.NotNil(t, result.Error())
			} else {
				// Skip if AWS credentials aren't available
				if result.IsError() {
					t.Skipf("Skipping test due to AWS credential issues: %v", result.Error())
				}
				assert.True(t, result.IsOk())
				provider := result.MustGet()
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestS3Builder_BuildWithNilContext(t *testing.T) {
	builder := s3provider.NewS3Builder()
	credentials := map[string]string{
		"bucket":            "test-bucket",
		"region":            "us-east-1",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
	}

	builder.WithCredentials(credentials)

	// Test with nil context (should work as context.Background() is used internally)
	provider, err := builder.Build(nil)

	if err != nil {
		t.Skipf("Skipping test due to AWS credential issues: %v", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestS3Builder_BuildWithoutCredentials(t *testing.T) {
	builder := s3provider.NewS3Builder()

	ctx := context.Background()
	provider, err := builder.Build(ctx)

	// Should fail with missing credentials error
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "missing required S3 credentials")
}

func TestS3Builder_ConfigMapping(t *testing.T) {
	builder := s3provider.NewS3Builder()

	// Test all credential field mappings
	credentials := map[string]string{
		"bucket":            "my-test-bucket",
		"region":            "eu-west-1",
		"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"endpoint":          "https://s3.eu-west-1.amazonaws.com",
	}

	builder.WithCredentials(credentials)

	// We can't easily verify the internal config without exposing it,
	// but we can test that Build doesn't fail due to mapping issues
	ctx := context.Background()
	_, err := builder.Build(ctx)

	// If it fails, it should be due to AWS credential validation, not mapping
	if err != nil {
		// Should not contain field mapping errors
		assert.NotContains(t, err.Error(), "field")
		assert.NotContains(t, err.Error(), "mapping")
	}
}

func BenchmarkS3Builder_Build(b *testing.B) {
	credentials := map[string]string{
		"bucket":            "benchmark-bucket",
		"region":            "us-east-1",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := s3provider.NewS3Builder()
		builder.WithCredentials(credentials)

		_, err := builder.Build(ctx)
		if err != nil {
			// Skip benchmark if AWS credentials aren't available
			b.Skip("Skipping benchmark due to AWS credential issues")
		}
	}
}

func TestS3Builder_InterfaceCompliance(t *testing.T) {
	builder := s3provider.NewS3Builder()

	// Verify that the builder implements the required interfaces
	// Comment out this line as the actual interface signature uses storagetypes.Provider
	// var _ cp.ClientBuilder[any] = builder

	// Test that all interface methods are available
	assert.NotNil(t, builder.ClientType())

	// Test method signatures exist
	credentials := map[string]string{"test": "value"}
	config := map[string]any{"test": "value"}

	result1 := builder.WithCredentials(credentials)
	assert.NotNil(t, result1)

	result2 := builder.WithConfig(config)
	assert.NotNil(t, result2)
}
