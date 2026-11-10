package disk_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	diskprovider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/disk"
)

func TestNewDiskBuilder(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	assert.Check(t, builder != nil)
}

func TestDiskBuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		credentials storage.ProviderCredentials
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:    "valid configuration",
			options: storage.NewProviderOptions(storage.WithBucket("/tmp/test-storage"), storage.WithLocalURL("http://localhost:8080/files")),
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
			provider, err := builder.Build(context.Background(), tt.credentials, tt.options)

			if tt.expectError {
				assert.Check(t, err != nil)
				assert.Check(t, is.Nil(provider))
			} else {
				assert.Check(t, err)
				assert.Check(t, provider != nil)
			}
		})
	}
}

func TestDiskBuilderProviderType(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	assert.Check(t, is.Equal(string(storagetypes.DiskProvider), builder.ProviderType()))
}
