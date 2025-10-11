package disk_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/objects/storage"
	diskprovider "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewDiskBuilder(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	assert.NotNil(t, builder)
}

func TestDiskBuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		credentials storage.ProviderCredentials
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:        "valid configuration",
			credentials: storage.ProviderCredentials{Endpoint: "http://localhost:8080/files"},
			options:     storage.NewProviderOptions(storage.WithBucket("/tmp/test-storage")),
		},
		{
			name:        "missing bucket uses default",
			credentials: storage.ProviderCredentials{},
			options:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := diskprovider.NewDiskBuilder()
			provider, err := builder.Build(tt.credentials, tt.options)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestDiskBuilderProviderType(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	assert.Equal(t, string(storagetypes.DiskProvider), builder.ProviderType())
}
