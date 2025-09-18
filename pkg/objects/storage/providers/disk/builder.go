package disk

import (
	"context"

	"github.com/samber/mo"
	"github.com/theopenlane/core/pkg/cp"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Builder creates disk providers for the client pool
type Builder struct {
	credentials map[string]string
	options     map[string]any
}

// NewDiskBuilder creates a new Builder
func NewDiskBuilder() *Builder {
	return &Builder{}
}

// WithCredentials implements cp.ClientBuilder
func (b *Builder) WithCredentials(credentials map[string]string) cp.ClientBuilder[storagetypes.Provider] {
	b.credentials = credentials
	return b
}

// WithConfig implements cp.ClientBuilder
func (b *Builder) WithConfig(config map[string]any) cp.ClientBuilder[storagetypes.Provider] {
	b.options = config

	return b
}

// Build implements cp.ClientBuilder
func (b *Builder) Build(context.Context) (storagetypes.Provider, error) {
	basePath := b.credentials["base_path"]

	// Check config options if credentials don't have base_path
	if basePath == "" && b.options != nil {
		if configBasePath, ok := b.options["base_path"].(string); ok {
			basePath = configBasePath
		}
	}

	if basePath == "" {
		// Use current working directory as default
		basePath = "./storage"
	}

	config := &Config{
		BasePath: basePath,
		LocalURL: b.credentials["local_url"],
	}

	return NewDiskProvider(config)
}

// ClientType implements cp.ClientBuilder
func (b *Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storagetypes.DiskProvider)
}

// NewDiskProviderFromCredentials creates a disk provider from credential map
func NewDiskProviderFromCredentials(credentials map[string]string) mo.Result[storagetypes.Provider] {
	builder := NewDiskBuilder().WithCredentials(credentials)
	provider, err := builder.Build(context.Background())
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok(provider)
}
