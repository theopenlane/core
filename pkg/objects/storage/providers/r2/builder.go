package r2

import (
	"context"

	"github.com/samber/mo"
	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Builder creates R2 providers for the client pool
type Builder struct{}

// NewR2Builder creates a new R2Builder
func NewR2Builder() *Builder {
	return &Builder{}
}

// Build implements eddy.Builder
func (b *Builder) Build(ctx context.Context, credentials storage.ProviderCredentials, config *storage.ProviderOptions) (storagetypes.Provider, error) {
	if config == nil {
		config = storage.NewProviderOptions()
	}

	cfg := config.Clone()
	cfg.Credentials = credentials

	if cfg.Bucket == "" || cfg.Credentials.AccountID == "" || cfg.Credentials.AccessKeyID == "" || cfg.Credentials.SecretAccessKey == "" {
		return nil, ErrR2CredentialsRequired
	}

	return NewR2Provider(cfg)
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
