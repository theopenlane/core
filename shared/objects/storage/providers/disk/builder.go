package disk

import (
	"context"

	"github.com/samber/mo"
	storage "github.com/theopenlane/shared/objects/storage"
	storagetypes "github.com/theopenlane/shared/objects/storage/types"
)

// Builder creates disk providers for the client pool
type Builder struct{}

// NewDiskBuilder creates a new Builder
func NewDiskBuilder() *Builder {
	return &Builder{}
}

// Build implements eddy.Builder
func (b *Builder) Build(_ context.Context, _ storage.ProviderCredentials, config *storage.ProviderOptions) (storagetypes.Provider, error) {
	if config == nil {
		config = storage.NewProviderOptions()
	}

	cfg := config.Clone()

	if cfg.Bucket == "" {
		cfg.Bucket = "./storage"
	}

	provider, err := NewDiskProvider(cfg)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// ProviderType implements eddy.Builder
func (b *Builder) ProviderType() string {
	return string(storagetypes.DiskProvider)
}

// NewDiskProviderFromCredentials creates a disk provider from credential struct
func NewDiskProviderFromCredentials(credentials storage.ProviderCredentials) mo.Result[storagetypes.Provider] {
	options := storage.NewProviderOptions(
		storage.WithCredentials(credentials),
		storage.WithBucket("./storage"),
	)
	provider, err := NewDiskProvider(options)
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok[storagetypes.Provider](provider)
}
