package awssecurityhub

import (
	"context"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// AccountScopeAll indicates operations should run across all accessible accounts
	AccountScopeAll = "all"
	// AccountScopeSpecific indicates operations should be limited to explicitly listed accounts
	AccountScopeSpecific = "specific"
)

// buildAWSConfig constructs an AWS SDK config with assume-role credentials
func buildAWSConfig(ctx context.Context, assumeRoleCredential AssumeRoleCredentialSchema, opCfg Config) (awssdk.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(assumeRoleCredential.HomeRegion),
	}

	if opCfg.AccessKeyID != "" && opCfg.SecretAccessKey != "" {
		provider := credentials.NewStaticCredentialsProvider(opCfg.AccessKeyID, opCfg.SecretAccessKey, "")
		opts = append(opts, config.WithCredentialsProvider(provider))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, ErrAWSConfigBuildFailed
	}

	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, assumeRoleCredential.RoleARN, func(options *stscreds.AssumeRoleOptions) {
		options.RoleSessionName = assumeRoleCredential.SessionName
		if assumeRoleCredential.ExternalID != "" {
			options.ExternalID = awssdk.String(assumeRoleCredential.ExternalID)
		}
		if duration := parseDuration(assumeRoleCredential.SessionDuration); duration > 0 {
			options.Duration = duration
		}
	})
	cfg.Credentials = awssdk.NewCredentialsCache(provider)

	return cfg, nil
}

// buildAWSConfigFromStaticCreds constructs an AWS SDK config from static IAM credentials
func buildAWSConfigFromStaticCreds(ctx context.Context, cred ServiceAccountCredentialSchema) (awssdk.Config, error) {
	if cred.Region == "" {
		return awssdk.Config{}, ErrRegionMissing
	}

	provider := credentials.NewStaticCredentialsProvider(cred.AccessKeyID, cred.SecretAccessKey, cred.SessionToken)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(provider),
		config.WithRegion(cred.Region),
	)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error loading aws config from static credentials")
		return cfg, ErrAWSConfigBuildFailed
	}

	return cfg, nil
}

// parseDuration parses a duration string and returns zero on empty or invalid input
func parseDuration(value string) time.Duration {
	if value == "" {
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}

	return duration
}
