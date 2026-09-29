package resolver_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/objects/resolver"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestResolveProviderUsesConfigWhenCredentialSyncDisabled(t *testing.T) {
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.S3Config{
				ProviderCommon: storage.ProviderCommon{
					Enabled: true,
					Bucket:  "test-bucket",
				},
				Region: "us-west-2",
				Credentials: storage.AccessKeyCredentials{
					AccessKeyID:     "test-access",
					SecretAccessKey: "test-secret",
				},
			},
		},
	}

	_, providerResolver, err := resolver.Build(config)
	assert.NoError(t, err)

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

func TestNewProviderDisk(t *testing.T) {
	bucket := filepath.Join(t.TempDir(), "bucket")

	provider, err := resolver.NewProvider(context.Background(), storage.DiskProvider, storage.Providers{
		Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: bucket}},
	})
	assert.NoError(t, err)
	assert.Equal(t, storage.DiskProvider, provider.ProviderType())
}

func TestNewProviderDisabled(t *testing.T) {
	provider, err := resolver.NewProvider(context.Background(), storage.DiskProvider, storage.Providers{
		Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Bucket: t.TempDir()}},
	})
	assert.Error(t, err)
	assert.Nil(t, provider)
}

func TestNewProviderUnknownType(t *testing.T) {
	provider, err := resolver.NewProvider(context.Background(), storage.ProviderType("ftp"), storage.Providers{
		Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: "records"}},
	})
	assert.Error(t, err)
	assert.Nil(t, provider)
}
