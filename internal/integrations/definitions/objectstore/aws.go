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

// runtimeBuilderOptions selects the AWS default credential chain for a runtime S3 bucket configured without keys
func runtimeBuilderOptions(ctx context.Context, provider storage.ProviderType, cfg storage.ProviderConfigs) ([]resolver.BuilderOption, error) {
	if provider != storage.S3Provider || cfg.Credentials.AccessKeyID != "" {
		return nil, nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	return []resolver.BuilderOption{resolver.WithS3Options(s3.WithAWSConfig(awsCfg))}, nil
}

// assumeRoleConfig builds the AWS config that assumes the customer role from the platform source identity,
// using the operator's static keys when set and the default credential chain otherwise
func assumeRoleConfig(ctx context.Context, cred AWSAssumeRoleCredentialSchema, source Config) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cred.Region),
	}

	if source.AccessKeyID != "" && source.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(source.AccessKeyID, source.SecretAccessKey, "")))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), cred.RoleARN, func(options *stscreds.AssumeRoleOptions) {
		options.RoleSessionName = cred.SessionName
		if cred.ExternalID != "" {
			options.ExternalID = aws.String(cred.ExternalID)
		}
	})
	cfg.Credentials = aws.NewCredentialsCache(provider)

	return cfg, nil
}

// assumeRoleSpec selects S3 through the cross-account role bound to the installation
func (c clientBuilder) assumeRoleSpec(ctx context.Context, bindings types.CredentialBindings) (providerSpec, error) {
	cred, err := decodeCredential(bindings, awsAssumeRoleCredential)
	if err != nil {
		return providerSpec{}, err
	}

	cfg, err := assumeRoleConfig(ctx, cred, c.aws)
	if err != nil {
		return providerSpec{}, err
	}

	return providerSpec{
		provider: storage.S3Provider,
		config: storage.ProviderConfigs{
			Enabled: true,
			Bucket:  cred.Bucket,
			Region:  cred.Region,
		},
		options: []resolver.BuilderOption{resolver.WithS3Options(s3.WithAWSConfig(cfg))},
	}, nil
}

// accessKeySpec selects S3 with the static IAM keys bound to the installation
func accessKeySpec(bindings types.CredentialBindings) (providerSpec, error) {
	cred, err := decodeCredential(bindings, awsAccessKeyCredential)
	if err != nil {
		return providerSpec{}, err
	}

	return providerSpec{
		provider: storage.S3Provider,
		config: storage.ProviderConfigs{
			Enabled: true,
			Bucket:  cred.Bucket,
			Region:  cred.Region,
			Credentials: storage.ProviderCredentials{
				AccessKeyID:     cred.AccessKeyID,
				SecretAccessKey: cred.SecretAccessKey,
			},
		},
	}, nil
}
