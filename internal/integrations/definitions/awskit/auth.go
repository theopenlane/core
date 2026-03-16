package awskit

import (
	"context"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/samber/lo"

	"github.com/theopenlane/core/pkg/jsonx"
)

// MetadataFromProviderData normalizes AWS metadata from a CredentialSet's ProviderData field,
// applying a default session name when one is not present in the stored data
func MetadataFromProviderData(providerData []byte, defaultSessionName string) (Metadata, error) {
	var decoded ProviderData
	if err := jsonx.UnmarshalIfPresent(providerData, &decoded); err != nil {
		return Metadata{}, ErrMetadataDecode
	}

	region := lo.CoalesceOrEmpty(decoded.Region, decoded.HomeRegion)
	homeRegion := lo.CoalesceOrEmpty(decoded.HomeRegion, region)
	sessionName := lo.CoalesceOrEmpty(decoded.SessionName, defaultSessionName)
	accountScope := lo.CoalesceOrEmpty(decoded.AccountScope, AccountScopeAll)

	return Metadata{
		Region:          region,
		HomeRegion:      homeRegion,
		LinkedRegions:   decoded.LinkedRegions,
		OrganizationID:  decoded.OrganizationID,
		AccountScope:    accountScope,
		AccountIDs:      decoded.AccountIDs,
		RoleARN:         decoded.RoleARN,
		AccountID:       decoded.AccountID,
		ExternalID:      decoded.ExternalID,
		SessionName:     sessionName,
		SessionDuration: ParseDuration(decoded.SessionDuration),
		AccessKeyID:     decoded.AccessKeyID,
		SecretAccessKey: decoded.SecretAccessKey,
		SessionToken:    decoded.SessionToken,
	}, nil
}

// CredentialsFromMetadata extracts static credential fields from a parsed Metadata struct
func CredentialsFromMetadata(meta Metadata) Credentials {
	return Credentials{
		AccessKeyID:     meta.AccessKeyID,
		SecretAccessKey: meta.SecretAccessKey,
		SessionToken:    meta.SessionToken,
	}
}

// BuildAWSConfig constructs an AWS SDK config with optional static and assumed credentials
func BuildAWSConfig(ctx context.Context, region string, creds Credentials, assume AssumeRole) (awssdk.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if creds.AccessKeyID != "" && creds.SecretAccessKey != "" {
		provider := credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)
		opts = append(opts, config.WithCredentialsProvider(provider))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, ErrAWSConfigBuildFailed
	}

	if assume.RoleARN == "" {
		return cfg, nil
	}

	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, assume.RoleARN, func(options *stscreds.AssumeRoleOptions) {
		options.RoleSessionName = assume.SessionName
		if assume.ExternalID != "" {
			options.ExternalID = awssdk.String(assume.ExternalID)
		}
		if assume.SessionDuration > 0 {
			options.Duration = assume.SessionDuration
		}
	})
	cfg.Credentials = awssdk.NewCredentialsCache(provider)

	return cfg, nil
}

// ParseDuration parses a duration string and returns 0 when empty or invalid
func ParseDuration(value string) time.Duration {
	if value == "" {
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}

	return duration
}
