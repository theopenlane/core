package corejobs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/riverqueue/river"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects/storage"
)

const (
	maxClearCacheRetries = 2
)

// ClearTrustCenterCacheArgs for the worker to clear trust center cache
type ClearTrustCenterCacheArgs struct {
	// CustomDomain is the custom domain for the trust center
	// If provided, will clear cache for this custom domain
	CustomDomain string `json:"custom_domain,omitempty"`

	// TrustCenterSlug is the slug for the trust center
	// Used with default domain: trust.theopenlane.net/<trust center slug>
	// If CustomDomain is not provided, this will be used
	TrustCenterSlug string `json:"trust_center_slug,omitempty"`
}

// Kind satisfies the river.Job interface
func (ClearTrustCenterCacheArgs) Kind() string { return "clear_trust_center_cache" }

// InsertOpts provides the insertion options for the clear trust center cache job
func (ClearTrustCenterCacheArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: maxClearCacheRetries}
}

// ClearTrustCenterCacheWorkerConfig contains the configuration for the clear trust center cache worker
type ClearTrustCenterCacheWorkerConfig struct {
	ObjectStorage storage.ProviderConfig `koanf:"objectstorage" json:"objectstorage" jsonschema:"description=the object storage configuration for R2"`

	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the clear trust center cache worker is enabled"`
}

// ClearTrustCenterCacheWorker is the worker to clear trust center cache from R2
type ClearTrustCenterCacheWorker struct {
	river.WorkerDefaults[ClearTrustCenterCacheArgs]

	Config ClearTrustCenterCacheWorkerConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for clearing trust center cache"`
}

// Work satisfies the river.Worker interface for the clear trust center cache worker
func (w *ClearTrustCenterCacheWorker) Work(ctx context.Context, job *river.Job[ClearTrustCenterCacheArgs]) error {
	logger := logx.FromContext(ctx)

	if job.Args.CustomDomain == "" && job.Args.TrustCenterSlug == "" {
		return errors.New("either the custom domain or trustcenter slug must be available") //nolint:err113
	}

	var prefix string

	if job.Args.TrustCenterSlug != "" {
		prefix = fmt.Sprintf("trust.theopenlane.net/%s/", strings.TrimSuffix(job.Args.TrustCenterSlug, "/"))
	}

	if job.Args.CustomDomain != "" {
		prefix = strings.TrimSuffix(job.Args.CustomDomain, "/")
	}

	r2Config := w.Config.ObjectStorage.Providers.CloudflareR2
	if !r2Config.Enabled {
		return fmt.Errorf("R2 provider is not enabled") //nolint:err113
	}

	if r2Config.Bucket == "" {
		return fmt.Errorf("R2 bucket is not configured") //nolint:err113
	}

	if r2Config.Endpoint == "" && r2Config.Credentials.AccountID == "" {
		return fmt.Errorf("R2 endpoint or account ID is not configured") //nolint:err113
	}

	endpoint := r2Config.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2Config.Credentials.AccountID)
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(r2Config.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			r2Config.Credentials.AccessKeyID,
			r2Config.Credentials.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err) //nolint:err113
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
		o.Region = r2Config.Region
		o.Credentials = awsCfg.Credentials
	})

	deletedCount := 0
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(r2Config.Bucket),
		Prefix: aws.String(prefix),
	}

	paginator := s3.NewListObjectsV2Paginator(s3Client, listInput)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			logger.Error().Err(err).Str("prefix", prefix).Msg("failed to list objects")
			return fmt.Errorf("failed to list objects: %w", err) //nolint:err113
		}

		if len(page.Contents) == 0 {
			continue
		}

		itemsToDelete := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			itemsToDelete = append(itemsToDelete, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(r2Config.Bucket),
			Delete: &types.Delete{
				Objects: itemsToDelete,
				Quiet:   aws.Bool(true),
			},
		}

		result, err := s3Client.DeleteObjects(ctx, deleteInput)
		if err != nil {
			logger.Error().Err(err).
				Str("prefix", prefix).
				Msg("failed to delete objects")
			return fmt.Errorf("failed to delete objects: %w", err) //nolint:err113
		}

		deletedCount += len(result.Deleted)
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				logger.Err(errors.New("could not delete r2 object")).
					Str("key", aws.ToString(err.Key)).
					Str("code", aws.ToString(err.Code)).
					Str("message", aws.ToString(err.Message)).
					Msg("failed to delete object")
			}
		}
	}

	return nil
}
