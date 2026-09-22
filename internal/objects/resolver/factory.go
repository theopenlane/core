package resolver

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/internal/objects/validators"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	dbprovider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/database"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/disk"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/gcs"
	r2provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/r2"
	s3provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/s3"
	"github.com/theopenlane/eddy"
	"github.com/theopenlane/iam/tokens"
)

type Option func(*serviceOptions)

type serviceOptions struct {
	tokenManagerFunc func() *tokens.TokenManager
	tokenAudience    string
	tokenIssuer      string
	baseURL          string
}

// providerResolver simplifies references to the eddy resolver used for object providers
type providerResolver = eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]

// providerClientService simplifies references to the eddy client service used for object providers
type providerClientService = eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]

// BuilderOption customises the provider builders NewProvider selects from
type BuilderOption func(*providerBuilders)

// WithS3Options applies provider options to the S3 builder
func WithS3Options(opts ...s3provider.Option) BuilderOption {
	return func(b *providerBuilders) {
		if builder, ok := b.s3.(*s3provider.Builder); ok {
			b.s3 = builder.WithOptions(opts...)
		}
	}
}

// WithR2Options applies provider options to the R2 builder
func WithR2Options(opts ...r2provider.Option) BuilderOption {
	return func(b *providerBuilders) {
		if builder, ok := b.r2.(*r2provider.Builder); ok {
			b.r2 = builder.WithOptions(opts...)
		}
	}
}

// WithGCSOptions applies provider options to the GCS builder
func WithGCSOptions(opts ...gcs.Option) BuilderOption {
	return func(b *providerBuilders) {
		if builder, ok := b.gcs.(*gcs.Builder); ok {
			b.gcs = builder.WithOptions(opts...)
		}
	}
}

// WithPresignConfig configures presigned URL token generation for providers that support it.
func WithPresignConfig(tokenManager func() *tokens.TokenManager, issuer, audience string) Option {
	return func(opts *serviceOptions) {
		opts.tokenManagerFunc = tokenManager
		opts.tokenIssuer = issuer
		opts.tokenAudience = audience
	}
}

// WithPresignBaseURL configures the base URL for presigned download URLs.
func WithPresignBaseURL(baseURL string) Option {
	return func(opts *serviceOptions) {
		opts.baseURL = baseURL
	}
}

// NewServiceFromConfig constructs a storage service complete with resolver rules derived from runtime configuration.
func NewServiceFromConfig(config storage.ProviderConfig, opts ...Option) (*objects.Service, error) {
	runtime := serviceOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&runtime)
		}
	}

	clientService, resolver, backups, err := buildWithRuntime(config, runtime)
	if err != nil {
		return nil, err
	}

	service := objects.NewService(objects.Config{
		Resolver:       resolver,
		ClientService:  clientService,
		ValidationFunc: validators.MimeTypeValidator,
		Backups:        backups,
	})

	return service, nil
}

// Build constructs the cp client service and provider resolver from runtime configuration.
func Build(config storage.ProviderConfig) (*providerClientService, *providerResolver, error) {
	clientService, resolver, _, err := buildWithRuntime(config, serviceOptions{})
	return clientService, resolver, err
}

// NewProvider builds one storage provider from the provider set, selecting the entry for the provider type
func NewProvider(ctx context.Context, provider storage.ProviderType, providers storage.Providers, opts ...BuilderOption) (storage.Provider, error) {
	builders := defaultBuilders(serviceOptions{})
	for _, opt := range opts {
		opt(&builders)
	}

	builder := builderFor(builders, provider)
	if builder == nil {
		return nil, fmt.Errorf("%w: %s", errUnsupportedProvider, provider)
	}

	options, creds, err := providerOptionsFromConfig(provider, storage.ProviderConfig{Providers: providers}, serviceOptions{})
	if err != nil {
		return nil, err
	}

	return builder.Build(ctx, creds, options)
}

// defaultBuilders constructs the builder set shared by the resolver rules and NewProvider
func defaultBuilders(runtime serviceOptions) providerBuilders {
	dbBuilder := dbprovider.NewBuilder()
	if runtime.tokenManagerFunc != nil {
		if tm := runtime.tokenManagerFunc(); tm != nil {
			dbBuilder = dbBuilder.WithOptions(
				dbprovider.WithTokenManager(tm),
				dbprovider.WithTokenClaims(runtime.tokenIssuer, runtime.tokenAudience),
			)
		}
	}

	return providerBuilders{
		s3:   s3provider.NewS3Builder(),
		r2:   r2provider.NewR2Builder(),
		gcs:  gcs.NewBuilder(),
		disk: disk.NewDiskBuilder(),
		db:   dbBuilder,
	}
}

func buildWithRuntime(config storage.ProviderConfig, runtime serviceOptions) (*providerClientService, *providerResolver, map[storage.ProviderType]objects.BackupTarget, error) {
	pool := eddy.NewClientPool[storage.Provider](objects.DefaultClientPoolTTL)
	clientService := eddy.NewClientService(pool, eddy.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials]((*storage.ProviderOptions).Clone))

	// Create resolver and configure rules with builders
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(defaultBuilders(runtime)),
		WithRuntimeOptions(runtime),
	)

	backups, err := buildBackupProviders(resolver, config)
	if err != nil {
		return nil, nil, nil, err
	}

	return clientService, resolver, backups, nil
}
