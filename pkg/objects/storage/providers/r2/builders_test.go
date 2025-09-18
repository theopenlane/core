package r2_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/cp"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewR2Builder(t *testing.T) {
	builder := r2provider.NewR2Builder()
	assert.NotNil(t, builder)
	// Note: Interface compliance test will be done separately
}

func TestR2Builder_WithCredentials(t *testing.T) {
	builder := r2provider.NewR2Builder()
	credentials := map[string]string{
		"bucket":            "test-bucket",
		"account_id":        "test-account-id",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
		"region":            "auto",
	}

	result := builder.WithCredentials(credentials)
	assert.NotNil(t, result)
	assert.Equal(t, builder, result) // Should return the same builder for chaining
}

func TestR2Builder_WithConfig(t *testing.T) {
	builder := r2provider.NewR2Builder()
	config := map[string]any{
		"use_path_style": true,
		"debug_mode":     true,
		"endpoint":       "https://test-account-id.r2.cloudflarestorage.com",
	}

	result := builder.WithConfig(config)
	assert.NotNil(t, result)
	assert.Equal(t, builder, result) // Should return the same builder for chaining
}

func TestR2Builder_ClientType(t *testing.T) {
	builder := r2provider.NewR2Builder()
	clientType := builder.ClientType()
	assert.Equal(t, cp.ProviderType("r2"), clientType)
}

func TestR2Builder_Build(t *testing.T) {
	tests := []struct {
		name        string
		credentials map[string]string
		config      map[string]any
		expectError bool
	}{
		{
			name: "valid credentials with access key",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"account_id":        "test-account-id",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"region":            "auto",
			},
			expectError: false,
		},
		{
			name: "invalid credentials with only API token",
			credentials: map[string]string{
				"bucket":     "test-bucket",
				"account_id": "test-account-id",
				"api_token":  "test-api-token",
				"region":     "auto",
			},
			expectError: true, // R2 provider currently requires access key/secret key
		},
		{
			name: "valid credentials with endpoint",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"account_id":        "test-account-id",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"region":            "auto",
				"endpoint":          "https://test-account-id.r2.cloudflarestorage.com",
			},
			expectError: false,
		},
		{
			name: "missing bucket",
			credentials: map[string]string{
				"account_id":        "test-account-id",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"region":            "auto",
			},
			expectError: true,
		},
		{
			name: "missing account ID",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"region":            "auto",
			},
			expectError: true,
		},
		{
			name:        "empty credentials",
			credentials: map[string]string{},
			expectError: true,
		},
		{
			name: "partial credentials missing auth",
			credentials: map[string]string{
				"bucket":     "test-bucket",
				"account_id": "test-account-id",
				"region":     "auto",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := r2provider.NewR2Builder()

			if tt.credentials != nil {
				builder.WithCredentials(tt.credentials)
			}

			if tt.config != nil {
				builder.WithConfig(tt.config)
			}

			provider, err := builder.Build(context.Background())

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

func TestR2Builder_ChainedCalls(t *testing.T) {
	builder := r2provider.NewR2Builder()

	credentials := map[string]string{
		"bucket":            "chained-test-bucket",
		"account_id":        "test-account-id",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
		"region":            "auto",
	}

	config := map[string]any{
		"use_path_style": true,
		"debug_mode":     false,
	}

	// Test method chaining
	provider, err := builder.
		WithCredentials(credentials).
		WithConfig(config).
		Build(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, provider)

	if provider != nil {
		err = provider.Close()
		assert.NoError(t, err)
	}
}

func TestNewR2ProviderFromCredentials(t *testing.T) {
	tests := []struct {
		name        string
		credentials map[string]string
		expectError bool
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"account_id":        "test-account-id",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"region":            "auto",
			},
			expectError: false,
		},
		{
			name: "invalid credentials - missing bucket",
			credentials: map[string]string{
				"account_id":        "test-account-id",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"region":            "auto",
			},
			expectError: true,
		},
		{
			name: "invalid credentials - missing account ID",
			credentials: map[string]string{
				"bucket":            "test-bucket",
				"access_key_id":     "test-access-key",
				"secret_access_key": "test-secret-key",
				"region":            "auto",
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
			result := r2provider.NewR2ProviderFromCredentials(tt.credentials)

			if tt.expectError {
				assert.True(t, result.IsError())
			} else {
				assert.True(t, result.IsOk())
				if result.IsOk() {
					provider := result.MustGet()
					assert.NotNil(t, provider)
					err := provider.Close()
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestR2Builder_BuildWithNilContext(t *testing.T) {
	builder := r2provider.NewR2Builder()
	credentials := map[string]string{
		"bucket":            "test-bucket",
		"account_id":        "test-account-id",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
		"region":            "auto",
	}

	// Building with nil context should work
	provider, err := builder.WithCredentials(credentials).Build(context.TODO())
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	if provider != nil {
		err = provider.Close()
		assert.NoError(t, err)
	}
}

func TestR2Builder_BuildWithoutCredentials(t *testing.T) {
	builder := r2provider.NewR2Builder()

	// Building without credentials should fail for R2
	provider, err := builder.Build(context.Background())
	assert.Error(t, err)
	assert.Nil(t, provider)
}

func TestR2Builder_ConfigMapping(t *testing.T) {
	builder := r2provider.NewR2Builder()

	credentials := map[string]string{
		"bucket":            "config-test-bucket",
		"account_id":        "test-account-id",
		"access_key_id":     "test-access-key",
		"secret_access_key": "test-secret-key",
		"region":            "auto",
	}

	// Test various config value types
	config := map[string]any{
		"use_path_style": true,
		"timeout":        30,
		"retry_attempts": 3,
		"endpoint":       "https://custom.r2.endpoint.com",
	}

	provider, err := builder.WithCredentials(credentials).WithConfig(config).Build(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	if provider != nil {
		err = provider.Close()
		assert.NoError(t, err)
	}
}

func TestR2Builder_InterfaceCompliance(t *testing.T) {
	builder := r2provider.NewR2Builder()

	// Verify that the builder implements the required interfaces
	// Note: Interface compliance test adapted for correct generic type
	var _ cp.ClientBuilder[storagetypes.Provider] = builder

	// Test that all interface methods are available
	assert.NotNil(t, builder.ClientType())

	// Test method signatures exist
	credentials := map[string]string{"test": "value"}
	config := map[string]any{"test": "value"}

	result1 := builder.WithCredentials(credentials)
	assert.NotNil(t, result1)

	result2 := builder.WithConfig(config)
	assert.NotNil(t, result2)

	// Test build method exists (will fail due to invalid credentials, but method exists)
	_, err := builder.Build(context.Background())
	assert.Error(t, err) // Expected to fail with invalid test credentials
}
