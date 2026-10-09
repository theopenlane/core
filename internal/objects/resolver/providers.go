package resolver

import (
	"fmt"

	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// builderFor maps a provider type to its builder from the provided set
func builderFor(builders providerBuilders, provider storage.ProviderType) providerBuilder {
	switch provider {
	case storage.S3Provider:
		return builders.s3
	case storage.R2Provider:
		return builders.r2
	case storage.GCSProvider:
		return builders.gcs
	case storage.DiskProvider:
		return builders.disk
	case storage.DatabaseProvider:
		return builders.db
	default:
		return nil
	}
}

// providerEnabled returns whether a provider can be used based on configuration.
func (rc *ruleCoordinator) providerEnabled(provider storage.ProviderType) bool {
	return rc.config.Providers.ByType()[provider].Enabled
}

// providerResolution is an internal type for credential resolution before adding builder
type providerResolution struct {
	Output storage.ProviderCredentials
	Config *storage.ProviderOptions
}

// resolveProvider returns provider credentials from system integrations or config fallback
func (rc *ruleCoordinator) resolveProvider(provider storage.ProviderType) (*providerResolution, error) {
	return resolveProviderFromConfig(provider, rc.config, rc.runtime)
}

func resolveProviderFromConfig(provider storage.ProviderType, config storage.ProviderConfig, runtime serviceOptions) (*providerResolution, error) {
	options, creds, err := providerOptionsFromConfig(provider, config, runtime)
	if err != nil {
		return nil, err
	}

	return &providerResolution{
		Output: creds,
		Config: options,
	}, nil
}

func providerOptionsFromConfig(provider storage.ProviderType, config storage.ProviderConfig, runtime serviceOptions) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	switch provider {
	case storage.S3Provider:
		return s3Options(config.Providers.S3, runtime)
	case storage.R2Provider:
		return r2Options(config.Providers.R2, runtime)
	case storage.GCSProvider:
		return gcsOptions(config.Providers.GCS, runtime)
	case storage.DiskProvider:
		return diskOptions(config.Providers.Disk, runtime)
	case storage.DatabaseProvider:
		return databaseOptions(config.Providers.Database, runtime)
	default:
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errUnsupportedProvider, provider)
	}
}

// proxyPresignOptions builds the proxy presign options every provider applies
func proxyPresignOptions(runtime serviceOptions, enabled bool, baseURL string) []storage.ProviderOption {
	options := []storage.ProviderOption{storage.WithProxyPresignEnabled(enabled)}

	if runtime.tokenManagerFunc == nil {
		return options
	}

	tm := runtime.tokenManagerFunc()
	if tm == nil {
		return options
	}

	presignOptions := []storage.ProxyPresignOption{
		storage.WithProxyPresignTokenManager(tm),
		storage.WithProxyPresignTokenIssuer(runtime.tokenIssuer),
		storage.WithProxyPresignTokenAudience(runtime.tokenAudience),
	}

	if baseURL != "" {
		presignOptions = append(presignOptions, storage.WithProxyPresignBaseURL(baseURL))
	}

	if runtime.baseURL != "" {
		presignOptions = append(presignOptions, storage.WithProxyPresignBaseURL(runtime.baseURL))
	}

	return append(options, storage.WithProxyPresignConfig(storage.NewProxyPresignConfig(presignOptions...)))
}

// s3Options builds provider options and credentials from the S3 configuration
func s3Options(cfg storage.S3Config, runtime serviceOptions) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	if !cfg.Enabled {
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errProviderDisabled, storage.S3Provider)
	}

	creds := cfg.Credentials.ProviderCredentials()

	options := storage.NewProviderOptions(storage.WithCredentials(creds))
	options.Apply(proxyPresignOptions(runtime, cfg.ProxyPresignEnabled, cfg.BaseURL)...)

	if cfg.Bucket != "" {
		options.Apply(storage.WithBucket(cfg.Bucket))
	}

	region := cfg.Region
	if region == "" {
		region = objects.DefaultS3Region
	}

	options.Apply(storage.WithRegion(region))

	if cfg.Endpoint != "" {
		options.Apply(storage.WithEndpoint(cfg.Endpoint))
	}

	return options, creds, nil
}

// r2Options builds provider options and credentials from the R2 configuration
func r2Options(cfg storage.R2Config, runtime serviceOptions) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	if !cfg.Enabled {
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errProviderDisabled, storage.R2Provider)
	}

	creds := cfg.Credentials.ProviderCredentials()

	options := storage.NewProviderOptions(storage.WithCredentials(creds))
	options.Apply(proxyPresignOptions(runtime, cfg.ProxyPresignEnabled, cfg.BaseURL)...)

	if cfg.Bucket != "" {
		options.Apply(storage.WithBucket(cfg.Bucket))
	}

	if cfg.Endpoint != "" {
		options.Apply(storage.WithEndpoint(cfg.Endpoint))
	}

	return options, creds, nil
}

// gcsOptions builds provider options from the GCS configuration
func gcsOptions(cfg storage.GCSConfig, runtime serviceOptions) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	if !cfg.Enabled {
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errProviderDisabled, storage.GCSProvider)
	}

	options := storage.NewProviderOptions(proxyPresignOptions(runtime, cfg.ProxyPresignEnabled, cfg.BaseURL)...)
	options.Apply(storage.WithExtra(storage.GCSProjectIDExtraKey, cfg.ProjectID))

	if cfg.Bucket != "" {
		options.Apply(storage.WithBucket(cfg.Bucket))
	}

	if cfg.Endpoint != "" {
		options.Apply(storage.WithEndpoint(cfg.Endpoint))
	}

	return options, storage.ProviderCredentials{}, nil
}

// diskOptions builds provider options from the disk configuration
func diskOptions(cfg storage.DiskConfig, runtime serviceOptions) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	if !cfg.Enabled {
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errProviderDisabled, storage.DiskProvider)
	}

	options := storage.NewProviderOptions(proxyPresignOptions(runtime, cfg.ProxyPresignEnabled, cfg.BaseURL)...)

	bucket := cfg.Bucket
	if bucket == "" {
		bucket = objects.DefaultDevStorageBucket
	}

	options.Apply(storage.WithBucket(bucket))

	if cfg.Endpoint != "" {
		options.Apply(storage.WithLocalURL(cfg.Endpoint))
	}

	return options, storage.ProviderCredentials{}, nil
}

// databaseOptions builds provider options from the database configuration
func databaseOptions(cfg storage.DatabaseConfig, runtime serviceOptions) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	if !cfg.Enabled {
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errProviderDisabled, storage.DatabaseProvider)
	}

	options := storage.NewProviderOptions(proxyPresignOptions(runtime, true, cfg.BaseURL)...)

	if cfg.Bucket != "" {
		options.Apply(storage.WithBucket(cfg.Bucket))
	}

	return options, storage.ProviderCredentials{}, nil
}
