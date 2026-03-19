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
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// AccountScopeAll indicates operations should run across all accessible accounts.
	AccountScopeAll = "all"
	// AccountScopeSpecific indicates operations should be limited to explicitly listed accounts.
	AccountScopeSpecific = "specific"
	// defaultSessionName is the STS session name used when no override is present in the credential metadata.
	defaultSessionName = "openlane-aws"
)

// credentialSchemaFromSet decodes one persisted credential set into the typed AWS credential schema.
func credentialSchemaFromSet(credential types.CredentialSet) (CredentialSchema, error) {
	if len(credential.ProviderData) == 0 {
		return CredentialSchema{}, ErrCredentialMetadataRequired
	}

	return credentialSchemaFromProviderData(credential.ProviderData)
}

// credentialSchemaFromProviderData decodes persisted AWS provider data into the typed schema and applies defaults.
func credentialSchemaFromProviderData(raw []byte) (CredentialSchema, error) {
	var credential CredentialSchema
	if err := jsonx.UnmarshalIfPresent(raw, &credential); err != nil {
		return CredentialSchema{}, ErrCredentialMetadataInvalid
	}

	return credential.applyDefaults(), nil
}

func (c CredentialSchema) applyDefaults() CredentialSchema {
	if c.AccountScope == "" {
		c.AccountScope = AccountScopeAll
	}

	if c.SessionName == "" {
		c.SessionName = defaultSessionName
	}

	return c
}

func buildAWSConfig(ctx context.Context, credential CredentialSchema) (awssdk.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(credential.HomeRegion),
	}

	if credential.AccessKeyID != "" && credential.SecretAccessKey != "" {
		provider := credentials.NewStaticCredentialsProvider(credential.AccessKeyID, credential.SecretAccessKey, credential.SessionToken)
		opts = append(opts, config.WithCredentialsProvider(provider))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, ErrAWSConfigBuildFailed
	}

	if credential.RoleARN == "" {
		return cfg, nil
	}

	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, credential.RoleARN, func(options *stscreds.AssumeRoleOptions) {
		options.RoleSessionName = credential.SessionName
		if credential.ExternalID != "" {
			options.ExternalID = awssdk.String(credential.ExternalID)
		}
		if duration := parseDuration(credential.SessionDuration); duration > 0 {
			options.Duration = duration
		}
	})
	cfg.Credentials = awssdk.NewCredentialsCache(provider)

	return cfg, nil
}

func buildAWSServiceClient[T any](ctx context.Context, req types.ClientBuildRequest, build func(awssdk.Config) *T) (*T, error) {
	credential, err := credentialSchemaFromSet(req.Credential)
	if err != nil {
		return nil, err
	}

	if credential.RoleARN == "" {
		return nil, ErrRoleARNMissing
	}

	if credential.HomeRegion == "" {
		return nil, ErrRegionMissing
	}

	cfg, err := buildAWSConfig(ctx, credential)
	if err != nil {
		return nil, err
	}

	return build(cfg), nil
}

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
