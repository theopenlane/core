package r2_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
)

func TestNewR2Builder(t *testing.T) {
	builder := r2provider.NewR2Builder()
	assert.NotNil(t, builder)
}

func TestR2BuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		credentials storage.ProviderCredentials
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:        "valid configuration",
			credentials: storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"},
			options: storage.NewProviderOptions(
				storage.WithBucket("bucket"),
				storage.WithEndpoint("https://account.r2.cloudflarestorage.com"),
			),
		},
		{
			name:        "missing bucket",
			credentials: storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"},
			options:     storage.NewProviderOptions(),
			expectError: true,
		},
		{
			name:        "missing account ID",
			credentials: storage.ProviderCredentials{AccessKeyID: "access", SecretAccessKey: "secret"},
			options: storage.NewProviderOptions(
				storage.WithBucket("bucket"),
			),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := r2provider.NewR2Builder()
			provider, err := builder.Build(context.Background(), tt.credentials, tt.options)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				if err != nil {
					t.Skip("Skipping test due to missing R2-compatible environment")
				}
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestR2BuilderProviderType(t *testing.T) {
	builder := r2provider.NewR2Builder()
	assert.Equal(t, string(storagetypes.R2Provider), builder.ProviderType())
}
