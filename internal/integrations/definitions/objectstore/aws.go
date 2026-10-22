package objectstore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/objects/resolver"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/s3"
)

// loadAWSConfig loads the AWS config for a region, using the static keys when set and the default credential chain otherwise
func loadAWSConfig(ctx context.Context, region string, keys storage.AccessKeyCredentials) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}

	if keys.AccessKeyID != "" && keys.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(keys.AccessKeyID, keys.SecretAccessKey, "")))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	return cfg, nil
}

// runtimeBuilderOptions selects the AWS default credential chain for a runtime S3 bucket configured without keys
func runtimeBuilderOptions(ctx context.Context, provider storage.ProviderType, cfg RuntimeConfig) ([]resolver.BuilderOption, error) {
	if provider != storage.S3Provider || cfg.Providers.S3.Credentials.AccessKeyID != "" {
		return nil, nil
	}

	awsCfg, err := loadAWSConfig(ctx, cfg.Providers.S3.Region, storage.AccessKeyCredentials{})
	if err != nil {
		return nil, err
	}

	return []resolver.BuilderOption{resolver.WithS3Options(s3.WithAWSConfig(awsCfg))}, nil
}

// assumeRoleClient returns the builder of the S3 storage client through the cross-account role bound to the installation, assumed from the platform source identity in cfg
func assumeRoleClient(cfg Config) func(context.Context, types.ConnectionRequest[AWSAssumeRoleCredentialSchema]) (*Client, error) {
	return func(ctx context.Context, req types.ConnectionRequest[AWSAssumeRoleCredentialSchema]) (*Client, error) {
		cred := req.Credential

		awsCfg, err := loadAWSConfig(ctx, cred.Region, storage.AccessKeyCredentials{AccessKeyID: cfg.AccessKeyID, SecretAccessKey: cfg.SecretAccessKey})
		if err != nil {
			return nil, err
		}

		provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(awsCfg), cred.RoleARN, func(options *stscreds.AssumeRoleOptions) {
			options.RoleSessionName = cred.SessionName
			if cred.ExternalID != "" {
				options.ExternalID = aws.String(cred.ExternalID)
			}
		})
		awsCfg.Credentials = aws.NewCredentialsCache(provider)

		return s3Client(ctx, cred.Bucket, cred.Region, storage.AccessKeyCredentials{}, resolver.WithS3Options(s3.WithAWSConfig(awsCfg)))
	}
}

// accessKeyClient builds the S3 storage client with the static IAM keys bound to the installation
func accessKeyClient(ctx context.Context, req types.ConnectionRequest[AWSAccessKeyCredentialSchema]) (*Client, error) {
	cred := req.Credential

	return s3Client(ctx, cred.Bucket, cred.Region, storage.AccessKeyCredentials{AccessKeyID: cred.AccessKeyID, SecretAccessKey: cred.SecretAccessKey})
}

// s3Client builds the S3 storage client for one bucket with the supplied provider options
func s3Client(ctx context.Context, bucket, region string, keys storage.AccessKeyCredentials, opts ...resolver.BuilderOption) (*Client, error) {
	provider, err := newProvider(ctx, storage.S3Provider, storage.Providers{S3: storage.S3Config{
		ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: bucket},
		Region:         region,
		Credentials:    keys,
	}}, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{Provider: provider}, nil
}
