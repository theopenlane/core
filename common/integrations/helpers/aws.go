package helpers

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
func AWSMetadataFromProviderData(meta map[string]any, defaultSessionName string) AWSMetadata {
	sessionName := StringValue(meta, "sessionName")
	if sessionName == "" {
		sessionName = defaultSessionName
	}

	return AWSMetadata{
		Region:          StringValue(meta, "region"),
		RoleARN:         StringValue(meta, "roleArn"),
		AccountID:       StringValue(meta, "accountId"),
		ExternalID:      StringValue(meta, "externalId"),
		SessionName:     sessionName,
		SessionDuration: ParseDuration(StringValue(meta, "sessionDuration")),
	}
}

// AWSCredentialsFromProviderData extracts access keys from a provider metadata map.
func AWSCredentialsFromProviderData(meta map[string]any) AWSCredentials {
	return AWSCredentials{
		AccessKeyID:     StringValue(meta, "accessKeyId"),
		SecretAccessKey: StringValue(meta, "secretAccessKey"),
		SessionToken:    StringValue(meta, "sessionToken"),
	}
}

// AWSCredentialsFromPayload extracts access keys from payload credentials with metadata fallback.
func AWSCredentialsFromPayload(payload types.CredentialPayload) AWSCredentials {
	accessKey := strings.TrimSpace(payload.Data.AccessKeyID)
	secretKey := strings.TrimSpace(payload.Data.SecretAccessKey)

	if accessKey == "" {
		accessKey = StringValue(payload.Data.ProviderData, "accessKeyId")
	}
	if secretKey == "" {
		secretKey = StringValue(payload.Data.ProviderData, "secretAccessKey")
	}

	return AWSCredentials{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		SessionToken:    StringValue(payload.Data.ProviderData, "sessionToken"),
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
