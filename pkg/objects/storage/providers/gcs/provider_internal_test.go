package gcs

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/api/option"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestBuilderBuildMergesCredentials(t *testing.T) {
	credentials := storage.ProviderCredentials{AccessKeyID: "access", SecretAccessKey: "secret"}
	options := storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithExtra(storage.GCSProjectIDExtraKey, "project"))

	built, err := NewBuilder().WithOptions(WithClientOptions(option.WithHTTPClient(&http.Client{}), option.WithoutAuthentication())).Build(context.Background(), credentials, options)
	assert.NilError(t, err)
	t.Cleanup(func() { _ = built.Close() })

	provider, ok := built.(*Provider)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal("bucket", provider.bucket))
	assert.Check(t, is.Equal("project", provider.projectID))
	assert.Check(t, is.DeepEqual(credentials, provider.options.Credentials))
	assert.Check(t, options != provider.options)
	assert.Check(t, is.DeepEqual(storage.ProviderCredentials{}, options.Credentials))
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
			assert.NilError(t, err)
			t.Cleanup(func() { _ = provider.Close() })

			assert.Check(t, is.Equal(tt.projectID, provider.projectID))
		})
	}
}
