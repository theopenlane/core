package validators

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/common/storagetypes"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects/storage"
	disk "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
)

const (
	StorageValidationTimeout     = 10 * time.Second
	StorageCredentialSyncTimeout = 10 * time.Second
)

var (
	// ErrBucketNotFound is returned when a provider does not contain the expected bucket
	ErrBucketNotFound = errors.New("bucket not found")
	// ErrUnableToCreateBucket is returned when the local bucket could not be created and the directory does not exist
	ErrUnableToCreateBucket = errors.New("unable to create bucket directory")
	// ErrProviderTypeMismatch is returned when provider's reported type doesn't match expected type
	ErrProviderTypeMismatch = errors.New("provider type mismatch")
)

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

	var errs []error

	if cfg.DevMode {
		if err := ensureDirectoryExists(objects.DefaultDevStorageBucket); err != nil {
			errs = append(errs, ErrUnableToCreateBucket)
		}

		logx.FromContext(ctx).Info().Msg("storage provider validation skipped in dev mode - disable dev mode to configure individual providers")

		return errs
	}

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

	if cfg.Providers.R2.Enabled && cfg.Providers.R2.EnsureAvailable {
		if err := validateR2Provider(ctx, cfg.Providers.R2); err != nil {
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

// validateDiskProvider checks connectivity to the disk provider and the existence of the specified bucket (directory)
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

	if err := validateProviderType(storage.DiskProvider, provider); err != nil {
		return err
	}

	return validateBuckets("disk", provider, bucket)
}

// validateS3Provider checks connectivity to the S3 provider and the existence of the specified bucket
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

	if err := validateProviderType(storage.S3Provider, provider); err != nil {
		return err
	}

	return validateBuckets("s3", provider, cfg.Bucket)
}

// validateR2Provider checks connectivity to Cloudflare R2 and the existence of the specified bucket
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

	if err := validateProviderType(storage.R2Provider, provider); err != nil {
		return err
	}

	return validateBuckets("r2", provider, cfg.Bucket)
}

// validateDatabaseProvider checks that the database provider can access the File table
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

// validateProviderType checks that the provider's reported type matches the expected provider constant
func validateProviderType(expected storagetypes.ProviderType, provider storagetypes.Provider) error {
	actual := provider.ProviderType()
	if actual != expected {
		return fmt.Errorf("%w: expected %s but provider reports %s", ErrProviderTypeMismatch, expected, actual)
	}
	return nil
}

// validateBuckets checks that the expected bucket exists in the provider's list of buckets
func validateBuckets(providerName string, provider storagetypes.Provider, expectedBucket string) error {
	buckets, err := provider.ListBuckets()
	if err != nil {
		return fmt.Errorf("%s list buckets: %w", providerName, err)
	}

	log.Info().Str("provider", providerName).Strs("available_buckets", buckets).Msg("storage provider connectivity verified")

	if expectedBucket != "" && !slices.Contains(buckets, expectedBucket) {
		return fmt.Errorf("%w: provider %s bucket %s", ErrBucketNotFound, providerName, expectedBucket)
	}

	return nil
}

// ensureDirectoryExists creates the directory at path if it does not already exist
func ensureDirectoryExists(path string) error {
	if path == "" {
		return nil
	}

	return os.MkdirAll(path, os.ModePerm)
}

// StorageAvailabilityCheck returns a handlers.CheckFunc that validates storage provider availability
func StorageAvailabilityCheck(cfgProvider func() storage.ProviderConfig) handlers.CheckFunc {
	return func(ctx context.Context) error {
		errs := ValidateAvailabilityByProvider(ctx, cfgProvider())
		if len(errs) == 0 {
			return nil
		}

		return errors.Join(errs...)
	}
}
