package gcs

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestBuilderBuildMergesCredentials(t *testing.T) {
	credentials := storage.ProviderCredentials{AccessKeyID: "access", SecretAccessKey: "secret"}
	options := storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithExtra(storage.GCSProjectIDExtraKey, "project"))

	built, err := NewBuilder().WithOptions(WithClientOptions(option.WithHTTPClient(&http.Client{}), option.WithoutAuthentication())).Build(context.Background(), credentials, options)
	require.NoError(t, err)
	t.Cleanup(func() { _ = built.Close() })

	provider, ok := built.(*Provider)
	require.True(t, ok)
	assert.Equal(t, "bucket", provider.bucket)
	assert.Equal(t, "project", provider.projectID)
	assert.Equal(t, credentials, provider.options.Credentials)
	assert.NotSame(t, options, provider.options)
	assert.Equal(t, storage.ProviderCredentials{}, options.Credentials)
}

func TestNewProviderProjectIDExtra(t *testing.T) {
	tests := []struct {
		name      string
		options   *storage.ProviderOptions
		projectID string
	}{
		{
			name:      "string project id",
			options:   storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithExtra(storage.GCSProjectIDExtraKey, "project")),
			projectID: "project",
		},
		{
			name:    "non-string project id",
			options: storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithExtra(storage.GCSProjectIDExtraKey, 42)),
		},
		{
			name:    "no project id",
			options: storage.NewProviderOptions(storage.WithBucket("bucket")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(context.Background(), tt.options, WithClientOptions(option.WithHTTPClient(&http.Client{}), option.WithoutAuthentication()))
			require.NoError(t, err)
			t.Cleanup(func() { _ = provider.Close() })

			assert.Equal(t, tt.projectID, provider.projectID)
		})
	}
}
