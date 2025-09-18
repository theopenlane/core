package objects

import (
	"context"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	"github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	"github.com/theopenlane/core/pkg/objects/storage/providers/s3"
)

// S3Builder builds S3 storage provider instances
type S3Builder struct {
	credentials map[string]string
	config      map[string]any
}

// WithCredentials sets the credentials for the S3 provider
func (b *S3Builder) WithCredentials(credentials map[string]string) cp.ClientBuilder[storage.Provider] {
	newBuilder := &S3Builder{
		credentials: credentials,
		config:      b.config,
	}

	return newBuilder
}

// WithConfig sets the configuration for the S3 provider
func (b *S3Builder) WithConfig(config map[string]any) cp.ClientBuilder[storage.Provider] {
	newBuilder := &S3Builder{
		credentials: b.credentials,
		config:      config,
	}

	return newBuilder
}

// Build creates a new S3 provider instance
func (b *S3Builder) Build(ctx context.Context) (storage.Provider, error) {
	builder := s3.NewS3Builder().WithCredentials(b.credentials).WithConfig(b.config)

	return builder.Build(ctx)
}

// ClientType returns the client type identifier
func (b *S3Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storage.S3Provider)
}

// R2Builder builds Cloudflare R2 storage provider instances
type R2Builder struct {
	credentials map[string]string
	config      map[string]any
}

// WithCredentials sets the credentials for the R2 provider
func (b *R2Builder) WithCredentials(credentials map[string]string) cp.ClientBuilder[storage.Provider] {
	newBuilder := &R2Builder{
		credentials: credentials,
		config:      b.config,
	}

	return newBuilder
}

// WithConfig sets the configuration for the R2 provider
func (b *R2Builder) WithConfig(config map[string]any) cp.ClientBuilder[storage.Provider] {
	newBuilder := &R2Builder{
		credentials: b.credentials,
		config:      config,
	}

	return newBuilder
}

// Build creates a new R2 provider instance
func (b *R2Builder) Build(ctx context.Context) (storage.Provider, error) {
	builder := r2.NewR2Builder().WithCredentials(b.credentials).WithConfig(b.config)

	return builder.Build(ctx)
}

// ClientType returns the client type identifier
func (b *R2Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storage.R2Provider)
}

// DiskBuilder builds local disk storage provider instances
type DiskBuilder struct {
	credentials map[string]string
	config      map[string]any
}

// WithCredentials sets the credentials for the disk provider (not used)
func (b *DiskBuilder) WithCredentials(credentials map[string]string) cp.ClientBuilder[storage.Provider] {
	newBuilder := &DiskBuilder{
		credentials: credentials,
		config:      b.config,
	}

	return newBuilder
}

// WithConfig sets the configuration for the disk provider
func (b *DiskBuilder) WithConfig(config map[string]any) cp.ClientBuilder[storage.Provider] {
	newBuilder := &DiskBuilder{
		credentials: b.credentials,
		config:      config,
	}

	return newBuilder
}

// Build creates a new disk provider instance
func (b *DiskBuilder) Build(ctx context.Context) (storage.Provider, error) {
	builder := disk.NewDiskBuilder().WithCredentials(b.credentials).WithConfig(b.config)

	return builder.Build(ctx)
}

// ClientType returns the client type identifier
func (b *DiskBuilder) ClientType() cp.ProviderType {
	return cp.ProviderType(storage.DiskProvider)
}

// GCSBuilder builds Google Cloud Storage provider instances
type GCSBuilder struct {
	credentials map[string]string
	config      map[string]any
}

// WithCredentials sets the credentials for the GCS provider
func (b *GCSBuilder) WithCredentials(credentials map[string]string) cp.ClientBuilder[storage.Provider] {
	newBuilder := &GCSBuilder{
		credentials: credentials,
		config:      b.config,
	}

	return newBuilder
}

// WithConfig sets the configuration for the GCS provider
func (b *GCSBuilder) WithConfig(config map[string]any) cp.ClientBuilder[storage.Provider] {
	newBuilder := &GCSBuilder{
		credentials: b.credentials,
		config:      config,
	}

	return newBuilder
}

// Build creates a new GCS provider instance
func (b *GCSBuilder) Build(_ context.Context) (storage.Provider, error) {
	return nil, storage.ErrProviderNotFound
}

// ClientType returns the client type identifier
func (b *GCSBuilder) ClientType() cp.ProviderType {
	return cp.ProviderType(storage.GCSProvider)
}
