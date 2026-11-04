package resolver_test

import (
	"context"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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
	assert.Check(t, err)

	ctx := ent.NewContext(context.Background(), &ent.Client{})

	resolution := providerResolver.Resolve(ctx)
	assert.Check(t, resolution.IsPresent(), "expected resolver to return a result")

	resolved := resolution.MustGet()
	assert.Check(t, is.Equal("test-access", resolved.Output.AccessKeyID))
	assert.Check(t, is.Equal("test-secret", resolved.Output.SecretAccessKey))
	assert.Check(t, resolved.Config != nil)
	assert.Check(t, is.Equal("test-bucket", resolved.Config.Bucket))
	assert.Check(t, is.Equal("us-west-2", resolved.Config.Region))
}

func TestNewProviderDisk(t *testing.T) {
	bucket := filepath.Join(t.TempDir(), "bucket")

	provider, err := resolver.NewProvider(context.Background(), storage.DiskProvider, storage.Providers{
		Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: bucket}},
	})
	assert.Check(t, err)
	assert.Check(t, is.Equal(storage.DiskProvider, provider.ProviderType()))
}

func TestNewProviderDisabled(t *testing.T) {
	provider, err := resolver.NewProvider(context.Background(), storage.DiskProvider, storage.Providers{
		Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Bucket: t.TempDir()}},
	})
	assert.Check(t, err != nil)
	assert.Check(t, is.Nil(provider))
}

func TestNewProviderUnknownType(t *testing.T) {
	provider, err := resolver.NewProvider(context.Background(), storage.ProviderType("ftp"), storage.Providers{
		Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: "records"}},
	})
	assert.Check(t, err != nil)
	assert.Check(t, is.Nil(provider))
}
