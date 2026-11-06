package gcs_test

import (
	"context"
	"slices"
	"testing"

	"google.golang.org/api/option"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/gcs"
)

func TestNewBuilder(t *testing.T) {
	builder := gcs.NewBuilder()
	assert.Check(t, builder != nil)
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
				assert.Check(t, is.ErrorIs(err, gcs.ErrBucketRequired))
				assert.Check(t, is.Nil(provider))
			} else {
				if err != nil {
					t.Skip("Skipping test due to missing GCS environment")
				}
				assert.Check(t, err)
				assert.Check(t, provider != nil)
				assert.Check(t, is.Equal(storage.GCSProvider, provider.ProviderType()))
			}
		})
	}
}

func TestBuilderWithOptions(t *testing.T) {
	server := newFakeServer(t, fakeObject(testBucket, "seeded.txt", "seeded"))

	builder := gcs.NewBuilder().WithOptions(fakeClientOptions(server, option.WithoutAuthentication()))
	provider, err := builder.Build(context.Background(), storage.ProviderCredentials{}, gcsOptions())
	assert.NilError(t, err)
	assert.Assert(t, provider != nil)
	t.Cleanup(func() { _ = provider.Close() })

	exists, err := provider.Exists(context.Background(), &storagetypes.File{Key: "seeded.txt"})
	assert.NilError(t, err)
	assert.Check(t, exists)

	buckets, err := provider.ListBuckets()
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values([]string{testBucket, otherBucket})), slices.Sorted(slices.Values(buckets))))
}

func TestWithServiceAccountKey(t *testing.T) {
	opt, err := gcs.WithServiceAccountKey(context.Background(), serviceAccountJSON(t))
	assert.NilError(t, err)
	assert.Check(t, opt != nil)

	opt, err = gcs.WithServiceAccountKey(context.Background(), []byte(`{"type":"authorized_user"}`))
	assert.Check(t, is.ErrorIs(err, gcs.ErrServiceAccountKeyInvalid))
	assert.Check(t, is.Nil(opt))

	opt, err = gcs.WithServiceAccountKey(context.Background(), nil)
	assert.Check(t, is.ErrorIs(err, gcs.ErrServiceAccountKeyInvalid))
	assert.Check(t, is.Nil(opt))
}

func TestBuilderProviderType(t *testing.T) {
	builder := gcs.NewBuilder()
	assert.Check(t, is.Equal(string(storage.GCSProvider), builder.ProviderType()))
}
