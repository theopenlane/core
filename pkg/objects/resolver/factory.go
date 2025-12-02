package resolver

import (
	"github.com/theopenlane/core/pkg/objects/objstore"
	"github.com/theopenlane/core/pkg/objects/storage"
	dbprovider "github.com/theopenlane/core/pkg/objects/storage/providers/database"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	"github.com/theopenlane/core/pkg/objects/validators"
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
func NewServiceFromConfig(config storage.ProviderConfig, opts ...Option) *objstore.Service {
	runtime := serviceOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&runtime)
		}
	}

	clientService, resolver := buildWithRuntime(config, runtime)

	service := objstore.NewService(objstore.Config{
		Resolver:       resolver,
		ClientService:  clientService,
		ValidationFunc: validators.MimeTypeValidator,
	})

	return service
}

// Build constructs the cp client service and provider resolver from runtime configuration.
func Build(config storage.ProviderConfig) (*providerClientService, *providerResolver) {
	return buildWithRuntime(config, serviceOptions{})
}

func buildWithRuntime(config storage.ProviderConfig, runtime serviceOptions) (*providerClientService, *providerResolver) {
	pool := eddy.NewClientPool[storage.Provider](objstore.DefaultClientPoolTTL)
	clientService := eddy.NewClientService(pool, eddy.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials](cloneProviderOptions))

	// Create builder instances
	s3Builder := s3provider.NewS3Builder()
	r2Builder := r2provider.NewR2Builder()
	diskBuilder := disk.NewDiskBuilder()
	dbBuilder := dbprovider.NewBuilder()
	if runtime.tokenManagerFunc != nil {
		if tm := runtime.tokenManagerFunc(); tm != nil {
			dbBuilder = dbBuilder.WithOptions(
				dbprovider.WithTokenManager(tm),
				dbprovider.WithTokenClaims(runtime.tokenIssuer, runtime.tokenAudience),
			)
		}
	}

	// Create resolver and configure rules with builders
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	builderSet := providerBuilders{
		s3:   s3Builder,
		r2:   r2Builder,
		disk: diskBuilder,
		db:   dbBuilder,
	}
	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(builderSet),
		WithRuntimeOptions(runtime),
	)

	return clientService, resolver
}

func cloneProviderOptions(in *storage.ProviderOptions) *storage.ProviderOptions {
	if in == nil {
		return nil
	}

	return in.Clone()
}
