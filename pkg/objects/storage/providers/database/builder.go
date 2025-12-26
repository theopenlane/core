package database

import (
	"context"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
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
func (b *Builder) Build(_ context.Context, _ storage.ProviderCredentials, config *storage.ProviderOptions) (storagetypes.Provider, error) {
	if config == nil {
		config = storage.NewProviderOptions()
	}

	options := config.Clone()
	if !options.ProxyPresignEnabled {
		options.ProxyPresignEnabled = true
	}

	options.ProxyPresignConfig = storage.ApplyProxyPresignOptions(options.ProxyPresignConfig)

	if options.ProxyPresignConfig.TokenManager == nil && b.tokenManager != nil {
		options.ProxyPresignConfig.TokenManager = b.tokenManager
	}
	if options.ProxyPresignConfig.TokenIssuer == "" && b.tokenIssuer != "" {
		options.ProxyPresignConfig.TokenIssuer = b.tokenIssuer
	}
	if options.ProxyPresignConfig.TokenAudience == "" && b.tokenAudience != "" {
		options.ProxyPresignConfig.TokenAudience = b.tokenAudience
	}

	provider := &Provider{
		options:     options,
		proxyConfig: options.ProxyPresignConfig,
	}

	return provider, nil
}

// ProviderType implements eddy.Builder.
func (b *Builder) ProviderType() string {
	return string(storage.DatabaseProvider)
}
