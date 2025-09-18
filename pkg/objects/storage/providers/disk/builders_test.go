package disk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/cp"
	diskprovider "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewDiskBuilder(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	assert.NotNil(t, builder)
	// Note: Interface compliance test skipped due to generic type differences
}

func TestDiskBuilder_WithCredentials(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	credentials := map[string]string{
		"bucket": "/tmp/test-storage",
	}

	result := builder.WithCredentials(credentials)
	assert.NotNil(t, result)
	assert.Equal(t, builder, result) // Should return the same builder for chaining
}

func TestDiskBuilder_WithConfig(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	config := map[string]any{
		"bucket":   "/tmp/test-storage",
		"create_dirs": true,
	}

	result := builder.WithConfig(config)
	assert.NotNil(t, result)
	assert.Equal(t, builder, result) // Should return the same builder for chaining
}

func TestDiskBuilder_ClientType(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	clientType := builder.ClientType()
	assert.Equal(t, cp.ProviderType("disk"), clientType)
}

func TestDiskBuilder_Build(t *testing.T) {
	tests := []struct {
		name        string
		credentials map[string]string
		config      map[string]any
		expectError bool
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"bucket": "/tmp/test-storage",
			},
			expectError: false,
		},
		{
			name: "valid credentials with config",
			credentials: map[string]string{
				"bucket": "/tmp/test-storage",
			},
			config: map[string]any{
				"create_dirs": true,
			},
			expectError: false,
		},
		{
			name:        "empty credentials (should use defaults)",
			credentials: map[string]string{},
			expectError: false,
		},
		{
			name:        "nil credentials (should use defaults)",
			credentials: nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := diskprovider.NewDiskBuilder()

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

func TestDiskBuilder_ChainedCalls(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()

	credentials := map[string]string{
		"bucket": "/tmp/chained-test",
	}

	config := map[string]any{
		"create_dirs": true,
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

func TestNewDiskProviderFromCredentials(t *testing.T) {
	tests := []struct {
		name        string
		credentials map[string]string
		expectError bool
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"bucket": "/tmp/test-disk-provider",
			},
			expectError: false,
		},
		{
			name:        "empty credentials",
			credentials: map[string]string{},
			expectError: false, // Should use defaults
		},
		{
			name:        "nil credentials",
			credentials: nil,
			expectError: false, // Should use defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diskprovider.NewDiskProviderFromCredentials(tt.credentials)

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

func TestDiskBuilder_BuildWithNilContext(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()

	// Building with nil context should work (context isn't typically used for disk)
	provider, err := builder.Build(nil)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	if provider != nil {
		err = provider.Close()
		assert.NoError(t, err)
	}
}

func TestDiskBuilder_BuildWithoutCredentials(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()

	// Building without credentials should use defaults
	provider, err := builder.Build(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	if provider != nil {
		err = provider.Close()
		assert.NoError(t, err)
	}
}

func TestDiskBuilder_ConfigMapping(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()

	// Test various config value types
	config := map[string]any{
		"bucket":   "/tmp/config-test",
		"create_dirs": true,
		"permissions": 0755,
	}

	provider, err := builder.WithConfig(config).Build(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	if provider != nil {
		err = provider.Close()
		assert.NoError(t, err)
	}
}

func TestDiskBuilder_InterfaceCompliance(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()

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

	// Test build method exists
	provider, err := builder.Build(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	if provider != nil {
		err = provider.Close()
		assert.NoError(t, err)
	}
}
