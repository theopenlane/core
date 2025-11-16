package resolver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/objects/resolver"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestResolveProviderUsesConfigWhenCredentialSyncDisabled(t *testing.T) {
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: true,
				Region:  "us-west-2",
				Bucket:  "test-bucket",
				Credentials: storage.ProviderCredentials{
					AccessKeyID:     "test-access",
					SecretAccessKey: "test-secret",
				},
			},
		},
	}

	_, providerResolver := resolver.Build(config)

	ctx := ent.NewContext(context.Background(), &ent.Client{})

	resolution := providerResolver.Resolve(ctx)
	assert.True(t, resolution.IsPresent(), "expected resolver to return a result")

	resolved := resolution.MustGet()
	assert.Equal(t, "test-access", resolved.Output.AccessKeyID)
	assert.Equal(t, "test-secret", resolved.Output.SecretAccessKey)
	assert.NotNil(t, resolved.Config)
	assert.Equal(t, "test-bucket", resolved.Config.Bucket)
	assert.Equal(t, "us-west-2", resolved.Config.Region)
}

