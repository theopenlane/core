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
)

const (
	// AccountScopeAll indicates operations should run across all accessible accounts
	AccountScopeAll = "all"
	// AccountScopeSpecific indicates operations should be limited to explicitly listed accounts
	AccountScopeSpecific = "specific"
	// defaultSessionName is the STS session name used when no override is present in the credential metadata
	defaultSessionName = "openlane-aws"
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

	return decoded.applyDefaults(), nil
}

// resolveSourceCredential extracts the optional static source credential from the bindings
func resolveSourceCredential(bindings types.CredentialBindings) (*SourceCredentialSchema, error) {
	decoded, ok, err := awsSourceCredential.Resolve(bindings)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, nil
	}

	return &decoded, nil
}

// applyDefaults fills in zero-value fields with sensible defaults
func (c AssumeRoleCredentialSchema) applyDefaults() AssumeRoleCredentialSchema {
	if c.AccountScope == "" {
		c.AccountScope = AccountScopeAll
	}

	if c.SessionName == "" {
		c.SessionName = defaultSessionName
	}

	return c
}

// buildAWSConfig constructs an AWS SDK config with assume-role credentials and optional static source credentials
func buildAWSConfig(ctx context.Context, assumeRoleCredential AssumeRoleCredentialSchema, sourceCredential *SourceCredentialSchema) (awssdk.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(assumeRoleCredential.HomeRegion),
	}

	if sourceCredential != nil && sourceCredential.AccessKeyID != "" && sourceCredential.SecretAccessKey != "" {
		provider := credentials.NewStaticCredentialsProvider(sourceCredential.AccessKeyID, sourceCredential.SecretAccessKey, sourceCredential.SessionToken)
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

// buildAWSServiceClient resolves credentials from the request and constructs a typed AWS service client
func buildAWSServiceClient[T any](ctx context.Context, req types.ClientBuildRequest, build func(awssdk.Config) *T) (*T, error) {
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

	sourceCredential, err := resolveSourceCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	cfg, err := buildAWSConfig(ctx, assumeRoleCredential, sourceCredential)
	if err != nil {
		return nil, err
	}

	return build(cfg), nil
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
