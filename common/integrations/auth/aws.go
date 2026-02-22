package auth

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// AWSAccountScopeAll indicates operations should run across all accessible accounts.
	AWSAccountScopeAll = "all"
	// AWSAccountScopeSpecific indicates operations should be limited to explicitly listed accounts.
	AWSAccountScopeSpecific = "specific"
)

// awsProviderData holds AWS provider data fields used for authentication
type awsProviderData struct {
	// Region is the AWS region for API calls
	Region types.TrimmedString `json:"region"`
	// HomeRegion is the Security Hub home region for aggregated queries
	HomeRegion types.TrimmedString `json:"homeRegion"`
	// LinkedRegions optionally limits queries to the listed regions
	LinkedRegions []string `json:"linkedRegions"`
	// OrganizationID is the AWS Organizations identifier associated with this integration
	OrganizationID types.TrimmedString `json:"organizationId"`
	// AccountScope controls whether queries should use all accounts or a provided subset
	AccountScope types.LowerString `json:"accountScope"`
	// AccountIDs optionally scopes collection to specific AWS account IDs
	AccountIDs []string `json:"accountIds"`
	// RoleARN is the ARN of the role to assume
	RoleARN types.TrimmedString `json:"roleArn"`
	// AccountID is the AWS account ID
	AccountID types.TrimmedString `json:"accountId"`
	// ExternalID is the external ID for role assumption
	ExternalID types.TrimmedString `json:"externalId"`
	// SessionName is the name for the session
	SessionName types.TrimmedString `json:"sessionName"`
	// SessionDuration is the duration for the session
	SessionDuration types.TrimmedString `json:"sessionDuration"`
	// AccessKeyID is the AWS access key ID
	AccessKeyID types.TrimmedString `json:"accessKeyId"`
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey types.TrimmedString `json:"secretAccessKey"`
	// SessionToken is the AWS session token
	SessionToken types.TrimmedString `json:"sessionToken"`
}

// AWSMetadata captures common AWS configuration fields stored in provider metadata
type AWSMetadata struct {
	// Region is the AWS region for API calls
	Region string
	// HomeRegion is the Security Hub home region for aggregated queries
	HomeRegion string
	// LinkedRegions optionally limits queries to the listed regions
	LinkedRegions []string
	// OrganizationID is the AWS Organizations identifier associated with this integration
	OrganizationID string
	// AccountScope controls whether queries should use all accounts or a provided subset
	AccountScope string
	// AccountIDs optionally scopes collection to specific AWS account IDs
	AccountIDs []string
	// RoleARN is the ARN of the role to assume
	RoleARN string
	// AccountID is the AWS account ID
	AccountID string
	// ExternalID is the external ID for role assumption
	ExternalID string
	// SessionName is the name for the session
	SessionName string
	// SessionDuration is the duration for the session
	SessionDuration time.Duration
}

// AWSAssumeRole captures the optional STS assume-role settings
type AWSAssumeRole struct {
	// RoleARN is the ARN of the role to assume
	RoleARN string
	// ExternalID is the external ID for role assumption
	ExternalID string
	// SessionName is the name for the session
	SessionName string
	// SessionDuration is the duration for the session
	SessionDuration time.Duration
}

// AWSCredentials captures static AWS access key credentials
type AWSCredentials struct {
	// AccessKeyID is the AWS access key ID
	AccessKeyID string
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string
	// SessionToken is the AWS session token
	SessionToken string
}

// AWSMetadataFromProviderData normalizes AWS metadata with a default session name
func AWSMetadataFromProviderData(meta map[string]any, defaultSessionName string) (AWSMetadata, error) {
	var decoded awsProviderData
	if err := DecodeProviderData(meta, &decoded); err != nil {
		return AWSMetadata{}, err
	}

	region := lo.CoalesceOrEmpty(decoded.Region.String(), decoded.HomeRegion.String())
	homeRegion := lo.CoalesceOrEmpty(decoded.HomeRegion.String(), region)
	sessionName := lo.CoalesceOrEmpty(decoded.SessionName.String(), defaultSessionName)
	accountScope := lo.CoalesceOrEmpty(decoded.AccountScope.String(), AWSAccountScopeAll)

	linkedRegions := types.NormalizeStringSlice(decoded.LinkedRegions)
	accountIDs := types.NormalizeStringSlice(decoded.AccountIDs)

	return AWSMetadata{
		Region:          region,
		HomeRegion:      homeRegion,
		LinkedRegions:   linkedRegions,
		OrganizationID:  decoded.OrganizationID.String(),
		AccountScope:    accountScope,
		AccountIDs:      accountIDs,
		RoleARN:         decoded.RoleARN.String(),
		AccountID:       decoded.AccountID.String(),
		ExternalID:      decoded.ExternalID.String(),
		SessionName:     sessionName,
		SessionDuration: ParseDuration(decoded.SessionDuration.String()),
	}, nil
}

// AWSCredentialsFromPayload extracts access keys from payload credentials with metadata fallback
func AWSCredentialsFromPayload(payload types.CredentialPayload) AWSCredentials {
	accessKey := payload.Data.AccessKeyID
	secretKey := payload.Data.SecretAccessKey
	sessionToken := payload.Data.SessionToken

	var decoded awsProviderData
	if err := DecodeProviderData(payload.Data.ProviderData, &decoded); err == nil {
		accessKey = lo.CoalesceOrEmpty(accessKey, decoded.AccessKeyID.String())
		secretKey = lo.CoalesceOrEmpty(secretKey, decoded.SecretAccessKey.String())
		sessionToken = lo.CoalesceOrEmpty(sessionToken, decoded.SessionToken.String())
	}

	return AWSCredentials{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		SessionToken:    sessionToken,
	}
}

// BuildAWSConfig constructs an AWS SDK config with optional static and assumed credentials
func BuildAWSConfig(ctx context.Context, region string, creds AWSCredentials, assume AWSAssumeRole) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if creds.AccessKeyID != "" && creds.SecretAccessKey != "" {
		provider := credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)
		opts = append(opts, config.WithCredentialsProvider(provider))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, err
	}

	if assume.RoleARN == "" {
		return cfg, nil
	}

	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, assume.RoleARN, func(options *stscreds.AssumeRoleOptions) {
		options.RoleSessionName = assume.SessionName
		if assume.ExternalID != "" {
			options.ExternalID = aws.String(assume.ExternalID)
		}
		if assume.SessionDuration > 0 {
			options.Duration = assume.SessionDuration
		}
	})
	cfg.Credentials = aws.NewCredentialsCache(provider)

	return cfg, nil
}
