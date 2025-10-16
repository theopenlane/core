package resolver

import (
	"fmt"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// providerEnabled returns whether a provider can be used based on configuration.
func (rc *ruleCoordinator) providerEnabled(provider storage.ProviderType) bool {
	switch provider {
	case storage.R2Provider:
		return rc.config.Providers.CloudflareR2.Enabled
	case storage.S3Provider:
		return rc.config.Providers.S3.Enabled
	case storage.DiskProvider:
		return rc.config.Providers.Disk.Enabled
	case storage.DatabaseProvider:
		return rc.config.Providers.Database.Enabled
	default:
		return false
	}
}

// providerResolution is an internal type for credential resolution before adding builder
type providerResolution struct {
	Output storage.ProviderCredentials
	Config *storage.ProviderOptions
}

// resolveProvider returns provider credentials from system integrations or config fallback
func (rc *ruleCoordinator) resolveProvider(provider storage.ProviderType) (*providerResolution, error) {
	return resolveProviderFromConfig(provider, rc.config)
}

func resolveProviderFromConfig(provider storage.ProviderType, config storage.ProviderConfig) (*providerResolution, error) {
	options, creds, err := providerOptionsFromConfig(provider, config)
	if err != nil {
		return nil, err
	}

	return &providerResolution{
		Output: creds,
		Config: options,
	}, nil
}

func providerOptionsFromConfig(provider storage.ProviderType, config storage.ProviderConfig) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	var providerCfg storage.ProviderConfigs

	switch provider {
	case storage.S3Provider:
		providerCfg = config.Providers.S3
	case storage.R2Provider:
		providerCfg = config.Providers.CloudflareR2
	case storage.GCSProvider:
		providerCfg = config.Providers.GCS
	case storage.DiskProvider:
		providerCfg = config.Providers.Disk
	case storage.DatabaseProvider:
		providerCfg = config.Providers.Database
	default:
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errUnsupportedProvider, provider)
	}

	if !providerCfg.Enabled {
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errProviderDisabled, provider)
	}

	options := storage.NewProviderOptions(storage.WithCredentials(providerCfg.Credentials))

	switch provider {
	case storage.S3Provider:
		if providerCfg.Bucket != "" {
			options.Apply(storage.WithBucket(providerCfg.Bucket))
		}
		region := providerCfg.Region
		if region == "" {
			region = objects.DefaultS3Region
		}
		options.Apply(storage.WithRegion(region))
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithEndpoint(providerCfg.Endpoint))
		}
	case storage.R2Provider, storage.GCSProvider:
		if providerCfg.Bucket != "" {
			options.Apply(storage.WithBucket(providerCfg.Bucket))
		}
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithEndpoint(providerCfg.Endpoint))
		}
	case storage.DiskProvider:
		bucket := providerCfg.Bucket
		if bucket == "" {
			bucket = objects.DefaultDevStorageBucket
		}
		options.Apply(storage.WithBucket(bucket), storage.WithBasePath(bucket))
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithLocalURL(providerCfg.Endpoint))
		}
	case storage.DatabaseProvider:
		if providerCfg.Bucket != "" {
			options.Apply(storage.WithBucket(providerCfg.Bucket))
		}
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithEndpoint(providerCfg.Endpoint))
		}
	}

	return options, providerCfg.Credentials, nil
}
