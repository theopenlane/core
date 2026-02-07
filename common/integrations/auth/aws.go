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

type awsProviderData struct {
	Region          string `json:"region"`
	RoleARN         string `json:"roleArn"`
	AccountID       string `json:"accountId"`
	ExternalID      string `json:"externalId"`
	SessionName     string `json:"sessionName"`
	SessionDuration string `json:"sessionDuration"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
}

// AWSMetadata captures common AWS configuration fields stored in provider metadata.
type AWSMetadata struct {
	Region          string
	RoleARN         string
	AccountID       string
	ExternalID      string
	SessionName     string
	SessionDuration time.Duration
}

// AWSAssumeRole captures the optional STS assume-role settings.
type AWSAssumeRole struct {
	RoleARN         string
	ExternalID      string
	SessionName     string
	SessionDuration time.Duration
}

// AWSCredentials captures static AWS access key credentials.
type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// AWSMetadataFromProviderData normalizes AWS metadata with a default session name.
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

// AWSCredentialsFromPayload extracts access keys from payload credentials with metadata fallback.
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

// BuildAWSConfig constructs an AWS SDK config with optional static and assumed credentials.
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
