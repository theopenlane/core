package s3

import (
	"context"
	"fmt"

	"github.com/samber/mo"
	"github.com/theopenlane/core/pkg/cp"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Builder creates S3 providers for the client pool
type Builder struct {
	credentials map[string]string
	options     map[string]any
}

// NewS3Builder creates a new S3Builder
func NewS3Builder() *Builder {
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
		return nil, fmt.Errorf("failed to parse S3 credentials: %w", err)
	}

	if config.Bucket == "" || config.Region == "" {
		return nil, ErrS3CredentialsRequired
	}

	return NewS3Provider(&config)
}

// ClientType implements cp.ClientBuilder
func (b *Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storagetypes.S3Provider)
}

// NewS3ProviderFromCredentials creates an S3 provider from credential map
func NewS3ProviderFromCredentials(credentials map[string]string) mo.Result[storagetypes.Provider] {
	builder := NewS3Builder().WithCredentials(credentials)
	provider, err := builder.Build(context.Background())
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok(provider)
}
