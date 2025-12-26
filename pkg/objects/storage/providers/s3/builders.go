package s3

import (
	"context"

	"github.com/samber/mo"
	"github.com/theopenlane/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// Builder creates S3 providers for the client pool
type Builder struct {
	opts []Option
}

// NewS3Builder creates a new S3Builder
func NewS3Builder() *Builder {
	return &Builder{}
}

// WithOptions allows configuring provider-specific options
func (b *Builder) WithOptions(opts ...Option) *Builder {
	b.opts = append(b.opts, opts...)

	return b
}

// Build implements eddy.Builder
func (b *Builder) Build(_ context.Context, credentials storage.ProviderCredentials, config *storage.ProviderOptions) (storagetypes.Provider, error) {
	if config == nil {
		config = storage.NewProviderOptions()
	}

	cfg := config.Clone()
	cfg.Credentials = credentials

	if cfg.Bucket == "" || cfg.Region == "" {
		return nil, ErrS3CredentialsRequired
	}

	if cfg.Credentials.AccessKeyID == "" || cfg.Credentials.SecretAccessKey == "" {
		return nil, ErrS3SecretCredentialRequired
	}

	provider, err := NewS3Provider(cfg, b.opts...)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// ProviderType implements eddy.Builder
func (b *Builder) ProviderType() string {
	return string(storagetypes.S3Provider)
}

// NewS3ProviderFromCredentials creates an S3 provider from provider credentials and optional configuration
func NewS3ProviderFromCredentials(credentials storage.ProviderCredentials, options *storage.ProviderOptions, opts ...Option) mo.Result[storagetypes.Provider] {
	cfg := storage.NewProviderOptions()
	if options != nil {
		cfg = options.Clone()
	}

	cfg.Credentials = credentials

	provider, err := NewS3Provider(cfg, opts...)
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok[storagetypes.Provider](provider)
}
