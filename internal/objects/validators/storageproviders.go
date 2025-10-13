package validators

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	disk "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

const (
	StorageValidationTimeout     = 10 * time.Second
	StorageCredentialSyncTimeout = 10 * time.Second
)

// ValidateConfiguredStorageProviders checks connectivity and configuration of all enabled storage providers
func ValidateConfiguredStorageProviders(ctx context.Context, cfg storage.ProviderConfig) []error {
	if !cfg.Enabled {
		log.Info().Msg("object storage disabled; skipping validation")
		return nil
	}

	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) <= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, StorageValidationTimeout)
		defer cancel()
	}

	if cfg.DevMode {
		if cfg.Providers.Disk.Enabled {
			bucket := cfg.Providers.Disk.Bucket
			if bucket == "" {
				bucket = objects.DefaultDevStorageBucket
			}

			if err := ensureDirectoryExists(bucket); err != nil {
				return []error{fmt.Errorf("ensure dev storage directory %s: %w", bucket, err)}
			}
		}

		return nil
	}

	var errs []error

	if err := validateDiskProvider(ctx, cfg.Providers.Disk); err != nil {
		errs = append(errs, err)
	}
	if err := validateS3Provider(ctx, cfg.Providers.S3); err != nil {
		errs = append(errs, err)
	}
	if err := validateR2Provider(ctx, cfg.Providers.CloudflareR2); err != nil {
		errs = append(errs, err)
	}
	if err := validateDatabaseProvider(ctx, cfg.Providers.Database); err != nil {
		errs = append(errs, err)
	}

	if cfg.Providers.GCS.Enabled {
		log.Warn().Msg("skipping GCS provider validation (not implemented)")
	}

	return errs
}

// ValidateAvailabilityByProvider validates only providers that have EnsureAvailable enabled.
// This allows per-provider strict availability enforcement instead of a global setting.
func ValidateAvailabilityByProvider(ctx context.Context, cfg storage.ProviderConfig) []error {
	if !cfg.Enabled {
		return nil
	}

	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) <= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, StorageValidationTimeout)
		defer cancel()
	}

	// In dev mode we don't strictly enforce availability
	if cfg.DevMode {
		return nil
	}

	var errs []error

	if cfg.Providers.Disk.Enabled && cfg.Providers.Disk.EnsureAvailable {
		if err := validateDiskProvider(ctx, cfg.Providers.Disk); err != nil {
			errs = append(errs, err)
		}
	}

	if cfg.Providers.S3.Enabled && cfg.Providers.S3.EnsureAvailable {
		if err := validateS3Provider(ctx, cfg.Providers.S3); err != nil {
			errs = append(errs, err)
		}
	}

	if cfg.Providers.CloudflareR2.Enabled && cfg.Providers.CloudflareR2.EnsureAvailable {
		if err := validateR2Provider(ctx, cfg.Providers.CloudflareR2); err != nil {
			errs = append(errs, err)
		}
	}

	if cfg.Providers.Database.Enabled && cfg.Providers.Database.EnsureAvailable {
		if err := validateDatabaseProvider(ctx, cfg.Providers.Database); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func validateDiskProvider(ctx context.Context, cfg storage.ProviderConfigs) error {
	if !cfg.Enabled {
		return nil
	}

	bucket := cfg.Bucket
	if bucket == "" {
		bucket = objects.DefaultDevStorageBucket
	}

	options := storage.NewProviderOptions(
		storage.WithBucket(bucket),
		storage.WithBasePath(bucket),
	)
	if cfg.Endpoint != "" {
		options.Apply(storage.WithLocalURL(cfg.Endpoint))
	}

	provider, err := disk.NewDiskBuilder().Build(ctx, cfg.Credentials, options)
	if err != nil {
		return fmt.Errorf("disk provider initialization: %w", err)
	}
	defer provider.Close()

	return validateBuckets("disk", provider, bucket)
}

func validateS3Provider(ctx context.Context, cfg storage.ProviderConfigs) error {
	if !cfg.Enabled {
		return nil
	}

	options := storage.NewProviderOptions()
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

	provider, err := s3provider.NewS3Builder().Build(ctx, cfg.Credentials, options)
	if err != nil {
		return fmt.Errorf("s3 provider initialization: %w", err)
	}
	defer provider.Close()

	return validateBuckets("s3", provider, cfg.Bucket)
}

func validateR2Provider(ctx context.Context, cfg storage.ProviderConfigs) error {
	if !cfg.Enabled {
		return nil
	}

	options := storage.NewProviderOptions()
	if cfg.Bucket != "" {
		options.Apply(storage.WithBucket(cfg.Bucket))
	}
	if cfg.Endpoint != "" {
		options.Apply(storage.WithEndpoint(cfg.Endpoint))
	}

	provider, err := r2provider.NewR2Builder().Build(ctx, cfg.Credentials, options)
	if err != nil {
		return fmt.Errorf("r2 provider initialization: %w", err)
	}
	defer provider.Close()

	return validateBuckets("r2", provider, cfg.Bucket)
}

func validateDatabaseProvider(ctx context.Context, cfg storage.ProviderConfigs) error {
	if !cfg.Enabled {
		return nil
	}

	entClient := ent.FromContext(ctx)
	if entClient == nil {
		// When no ent client is available (e.g. during early startup), skip validation.
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	if _, err := entClient.File.Query().Limit(1).Exist(allowCtx); err != nil {
		return fmt.Errorf("database provider validation: %w", err)
	}

	return nil
}

func validateBuckets(providerName string, provider storagetypes.Provider, expectedBucket string) error {
	buckets, err := provider.ListBuckets()
	if err != nil {
		return fmt.Errorf("%s list buckets: %w", providerName, err)
	}

	log.Info().Str("provider", providerName).Strs("available_buckets", buckets).Msg("storage provider connectivity verified")

	if expectedBucket != "" && !slices.Contains(buckets, expectedBucket) {
		return fmt.Errorf("%s bucket %s not found", providerName, expectedBucket)
	}

	return nil
}

func ensureDirectoryExists(path string) error {
	if path == "" {
		return nil
	}

	return os.MkdirAll(path, os.ModePerm)
}

func StorageAvailabilityCheck(cfgProvider func() storage.ProviderConfig) handlers.CheckFunc {
	return func(ctx context.Context) error {
		errs := ValidateConfiguredStorageProviders(ctx, cfgProvider())
		if len(errs) == 0 {
			return nil
		}

		return errors.Join(errs...)
	}
}
