package gcs

import (
	"context"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// Builder creates GCS providers for the client pool
type Builder struct {
	options []Option
}

// NewBuilder creates a new Builder
func NewBuilder() *Builder {
	return &Builder{}
}

// WithOptions sets provider options for the builder
func (b *Builder) WithOptions(opts ...Option) *Builder {
	b.options = append(b.options, opts...)
	return b
}

// Build implements eddy.Builder
func (b *Builder) Build(ctx context.Context, credentials storage.ProviderCredentials, config *storage.ProviderOptions) (storagetypes.Provider, error) {
	if config == nil {
		config = storage.NewProviderOptions()
	}

	cfg := config.Clone()
	cfg.Credentials = credentials

	provider, err := NewProvider(ctx, cfg, b.options...)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// ProviderType implements eddy.Builder
func (b *Builder) ProviderType() string {
	return string(storage.GCSProvider)
}
