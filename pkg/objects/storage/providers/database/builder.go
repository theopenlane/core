package database

import (
	"context"

	"github.com/theopenlane/iam/tokens"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Option configures the database provider builder.
type Option func(*Builder)

// WithTokenManager supplies the token manager used for presigned URL generation.
func WithTokenManager(tm *tokens.TokenManager) Option {
	return func(b *Builder) {
		b.tokenManager = tm
	}
}

// WithTokenClaims configures issuer and audience values for presigned tokens.
func WithTokenClaims(issuer, audience string) Option {
	return func(b *Builder) {
		b.tokenIssuer = issuer
		b.tokenAudience = audience
	}
}

// Builder creates database providers for the client pool.
type Builder struct {
	tokenManager  *tokens.TokenManager
	tokenAudience string
	tokenIssuer   string
}

// NewBuilder returns a new database provider builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// WithOptions allows applying builder-specific options.
func (b *Builder) WithOptions(opts ...Option) *Builder {
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}

	return b
}

// Build implements eddy.Builder.
func (b *Builder) Build(ctx context.Context, credentials storage.ProviderCredentials, config *storage.ProviderOptions) (storagetypes.Provider, error) {
	if config == nil {
		config = storage.NewProviderOptions()
	}

	provider := &Provider{
		options:       config.Clone(),
		tokenManager:  b.tokenManager,
		tokenAudience: b.tokenAudience,
		tokenIssuer:   b.tokenIssuer,
	}

	return provider, nil
}

// ProviderType implements eddy.Builder.
func (b *Builder) ProviderType() string {
	return string(storage.DatabaseProvider)
}
