package auth

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/common/integrations/types"
)

// awsProviderData holds AWS provider data fields used for authentication
type awsProviderData struct {
	// Region is the AWS region for API calls
	Region string `json:"region"`
	// RoleARN is the ARN of the role to assume
	RoleARN string `json:"roleArn"`
	// AccountID is the AWS account ID
	AccountID string `json:"accountId"`
	// ExternalID is the external ID for role assumption
	ExternalID string `json:"externalId"`
	// SessionName is the name for the session
	SessionName string `json:"sessionName"`
	// SessionDuration is the duration for the session
	SessionDuration string `json:"sessionDuration"`
	// AccessKeyID is the AWS access key ID
	AccessKeyID string `json:"accessKeyId"`
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string `json:"secretAccessKey"`
	// SessionToken is the AWS session token
	SessionToken string `json:"sessionToken"`
}

// AWSMetadata captures common AWS configuration fields stored in provider metadata
type AWSMetadata struct {
	// Region is the AWS region for API calls
	Region string
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

	sessionName := decoded.SessionName
	if sessionName == "" {
		sessionName = defaultSessionName
	}

	return AWSMetadata{
		Region:          decoded.Region,
		RoleARN:         decoded.RoleARN,
		AccountID:       decoded.AccountID,
		ExternalID:      decoded.ExternalID,
		SessionName:     sessionName,
		SessionDuration: ParseDuration(decoded.SessionDuration),
	}, nil
}

// AWSCredentialsFromPayload extracts access keys from payload credentials with metadata fallback
func AWSCredentialsFromPayload(payload types.CredentialPayload) AWSCredentials {
	accessKey := strings.TrimSpace(payload.Data.AccessKeyID)
	secretKey := strings.TrimSpace(payload.Data.SecretAccessKey)
	sessionToken := strings.TrimSpace(payload.Data.SessionToken)

	var decoded awsProviderData
	if err := DecodeProviderData(payload.Data.ProviderData, &decoded); err == nil {
		if accessKey == "" {
			accessKey = decoded.AccessKeyID
		}
		if secretKey == "" {
			secretKey = decoded.SecretAccessKey
		}
		if sessionToken == "" {
			sessionToken = decoded.SessionToken
		}
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

	if strings.TrimSpace(assume.RoleARN) == "" {
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
