package gcs_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/gcs"
)

func TestNewBuilder(t *testing.T) {
	builder := gcs.NewBuilder()
	assert.NotNil(t, builder)
}

func TestBuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:    "valid configuration",
			options: storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithExtra(storage.GCSProjectIDExtraKey, "project")),
		},
		{
			name:        "missing bucket",
			options:     storage.NewProviderOptions(),
			expectError: true,
		},
		{
			name:        "nil options",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := gcs.NewBuilder().WithOptions(gcs.WithClientOptions(option.WithoutAuthentication()))
			provider, err := builder.Build(context.Background(), storage.ProviderCredentials{}, tt.options)

			if tt.expectError {
				assert.ErrorIs(t, err, gcs.ErrBucketRequired)
				assert.Nil(t, provider)
			} else {
				if err != nil {
					t.Skip("Skipping test due to missing GCS environment")
				}
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, storage.GCSProvider, provider.ProviderType())
			}
		})
	}
}

func TestBuilderWithOptions(t *testing.T) {
	server := newFakeServer(t, fakeObject(testBucket, "seeded.txt", "seeded"))

	builder := gcs.NewBuilder().WithOptions(fakeClientOptions(server, option.WithoutAuthentication()))
	provider, err := builder.Build(context.Background(), storage.ProviderCredentials{}, gcsOptions())
	require.NoError(t, err)
	require.NotNil(t, provider)
	t.Cleanup(func() { _ = provider.Close() })

	exists, err := provider.Exists(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "seeded.txt"}})
	require.NoError(t, err)
	assert.True(t, exists)

	buckets, err := provider.ListBuckets()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{testBucket, otherBucket}, buckets)
}

func TestBuilderProviderType(t *testing.T) {
	builder := gcs.NewBuilder()
	assert.Equal(t, string(storage.GCSProvider), builder.ProviderType())
}
