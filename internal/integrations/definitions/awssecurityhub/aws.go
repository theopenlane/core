package awssecurityhub

import (
	"context"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// AccountScopeAll indicates operations should run across all accessible accounts
	AccountScopeAll = "all"
	// AccountScopeSpecific indicates operations should be limited to explicitly listed accounts
	AccountScopeSpecific = "specific"
)

// resolveAssumeRoleCredential extracts and validates the assume-role credential from the bindings
func resolveAssumeRoleCredential(bindings types.CredentialBindings) (AssumeRoleCredentialSchema, error) {
	decoded, ok, err := awsAssumeRoleCredential.Resolve(bindings)
	if err != nil {
		return AssumeRoleCredentialSchema{}, ErrCredentialMetadataInvalid
	}

	if !ok {
		return AssumeRoleCredentialSchema{}, ErrCredentialMetadataRequired
	}

	return decoded, nil
}

// buildAWSConfig constructs an AWS SDK config with assume-role credentials sourced from the operator config
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

// buildAWSServiceClient resolves credentials from the request and constructs a typed AWS service client.
// It uses the assume-role path when an assume-role credential is bound, otherwise falls back to static credentials.
func buildAWSServiceClient[T any](ctx context.Context, cfg Config, req types.ClientBuildRequest, build func(awssdk.Config) *T) (*T, error) {
	_, hasAssumeRole := req.Credentials.Resolve(awsAssumeRoleCredential.ID())
	if hasAssumeRole {
		return buildAWSServiceClientViaAssumeRole(ctx, cfg, req, build)
	}

	return buildAWSServiceClientViaStaticCreds(ctx, req, build)
}

// buildAWSServiceClientViaAssumeRole constructs a client using STS cross-account assume-role
func buildAWSServiceClientViaAssumeRole[T any](ctx context.Context, opCfg Config, req types.ClientBuildRequest, build func(awssdk.Config) *T) (*T, error) {
	assumeRoleCredential, err := resolveAssumeRoleCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if assumeRoleCredential.RoleARN == "" {
		return nil, ErrRoleARNMissing
	}

	if assumeRoleCredential.HomeRegion == "" {
		return nil, ErrRegionMissing
	}

	cfg, err := buildAWSConfig(ctx, assumeRoleCredential, opCfg)
	if err != nil {
		return nil, err
	}

	return build(cfg), nil
}

// buildAWSServiceClientViaStaticCreds constructs a client using static IAM credentials directly
func buildAWSServiceClientViaStaticCreds[T any](ctx context.Context, req types.ClientBuildRequest, build func(awssdk.Config) *T) (*T, error) {
	serviceAccount, ok, err := awsServiceAccountCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error resolving aws credentials")
		return nil, ErrCredentialMetadataInvalid
	}

	if !ok {
		return nil, ErrCredentialMetadataRequired
	}

	cfg, err := buildAWSConfigFromStaticCreds(ctx, serviceAccount)
	if err != nil {
		return nil, err
	}

	return build(cfg), nil
}

// buildAWSConfigFromStaticCreds constructs an AWS SDK config using static IAM credentials and a region
func buildAWSConfigFromStaticCreds(ctx context.Context, cred ServiceAccountCredentialSchema) (awssdk.Config, error) {
	provider := credentials.NewStaticCredentialsProvider(cred.AccessKeyID, cred.SecretAccessKey, cred.SessionToken)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(provider),
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
