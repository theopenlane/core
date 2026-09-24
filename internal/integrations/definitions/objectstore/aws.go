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

// assumeRoleProvider builds S3 through the cross-account role bound to the installation, assumed from the platform source identity
func (c clientBuilder) assumeRoleProvider(ctx context.Context, bindings types.CredentialBindings) (storage.Provider, error) {
	cred, err := decodeCredential(bindings, awsAssumeRoleCredential)
	if err != nil {
		return nil, err
	}

	cfg, err := loadAWSConfig(ctx, cred.Region, storage.AccessKeyCredentials{AccessKeyID: c.aws.AccessKeyID, SecretAccessKey: c.aws.SecretAccessKey})
	if err != nil {
		return nil, err
	}

	provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), cred.RoleARN, func(options *stscreds.AssumeRoleOptions) {
		options.RoleSessionName = cred.SessionName
		if cred.ExternalID != "" {
			options.ExternalID = aws.String(cred.ExternalID)
		}
	})
	cfg.Credentials = aws.NewCredentialsCache(provider)

	return newProvider(ctx, storage.S3Provider, storage.Providers{S3: storage.S3Config{
		ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: cred.Bucket},
		Region:         cred.Region,
	}}, resolver.WithS3Options(s3.WithAWSConfig(cfg)))
}

// accessKeyProvider builds S3 with the static IAM keys bound to the installation
func accessKeyProvider(ctx context.Context, bindings types.CredentialBindings) (storage.Provider, error) {
	cred, err := decodeCredential(bindings, awsAccessKeyCredential)
	if err != nil {
		return nil, err
	}

	return newProvider(ctx, storage.S3Provider, storage.Providers{S3: storage.S3Config{
		ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: cred.Bucket},
		Region:         cred.Region,
		Credentials:    storage.AccessKeyCredentials{AccessKeyID: cred.AccessKeyID, SecretAccessKey: cred.SecretAccessKey},
	}})
}
