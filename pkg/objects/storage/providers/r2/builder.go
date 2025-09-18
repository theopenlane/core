package r2

import (
	"context"
	"fmt"

	"github.com/samber/mo"
	"github.com/theopenlane/core/pkg/cp"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Builder creates R2 providers for the client pool
type Builder struct {
	credentials map[string]string
	options     map[string]any
}

// NewR2Builder creates a new R2Builder
func NewR2Builder() *Builder {
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
func (b *Builder) Build(_ context.Context) (storagetypes.Provider, error) {
	config, err := cp.StructFromCredentials[Config](b.credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to parse R2 credentials: %w", err)
	}

	if config.Bucket == "" || config.AccountID == "" || config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return nil, ErrR2CredentialsRequired
	}

	return NewR2Provider(&config)
}

// ClientType implements cp.ClientBuilder
func (b *Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storagetypes.R2Provider)
}

// NewR2ProviderFromCredentials creates an R2 provider from credential map
func NewR2ProviderFromCredentials(credentials map[string]string) mo.Result[storagetypes.Provider] {
	builder := NewR2Builder().WithCredentials(credentials)
	provider, err := builder.Build(context.Background())
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok(provider)
}
