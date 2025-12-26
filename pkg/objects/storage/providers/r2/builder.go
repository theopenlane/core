package r2

import (
	"context"

	"github.com/samber/mo"
	"github.com/theopenlane/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// Builder creates R2 providers for the client pool
type Builder struct {
	options []Option
}

// NewR2Builder creates a new R2Builder
func NewR2Builder() *Builder {
	return &Builder{}
}

// WithOptions sets provider options for the builder
func (b *Builder) WithOptions(opts ...Option) *Builder {
	b.options = append(b.options, opts...)
	return b
}

// Build implements eddy.Builder
func (b *Builder) Build(_ context.Context, credentials storage.ProviderCredentials, config *storage.ProviderOptions) (storagetypes.Provider, error) {
	if config == nil {
		config = storage.NewProviderOptions()
	}

	cfg := config.Clone()
	cfg.Credentials = credentials

	if cfg.Bucket == "" || cfg.Credentials.AccessKeyID == "" || cfg.Credentials.SecretAccessKey == "" {
		return nil, ErrR2CredentialsRequired
	}

	return NewR2Provider(cfg, b.options...)
}

// ProviderType implements eddy.Builder
func (b *Builder) ProviderType() string {
	return string(storagetypes.R2Provider)
}

// NewR2ProviderFromCredentials creates an R2 provider using the supplied credentials and options
func NewR2ProviderFromCredentials(credentials storage.ProviderCredentials, options *storage.ProviderOptions) mo.Result[storagetypes.Provider] {
	cfg := storage.NewProviderOptions()
	if options != nil {
		cfg = options.Clone()
	}
	cfg.Credentials = credentials

	provider, err := NewR2Provider(cfg)
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok[storagetypes.Provider](provider)
}
