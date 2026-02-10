package awssts

import (
	"context"
	"maps"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
)

// ProviderOption customizes AWS STS providers.
type ProviderOption func(*providerConfig)

type providerConfig struct {
	operations []types.OperationDescriptor
	clients    []types.ClientDescriptor
}

// WithOperations registers provider-published operations.
func WithOperations(descriptors []types.OperationDescriptor) ProviderOption {
	return func(cfg *providerConfig) {
		cfg.operations = descriptors
	}
}

// WithClientDescriptors registers client descriptors for pooling.
func WithClientDescriptors(descriptors []types.ClientDescriptor) ProviderOption {
	return func(cfg *providerConfig) {
		cfg.clients = descriptors
	}
}

// Builder returns a providers.Builder that materializes AWS federation metadata.
func Builder(provider types.ProviderType, opts ...ProviderOption) providers.Builder {
	cfg := &providerConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	return providers.BuilderFunc{
		ProviderType: provider,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.AuthType != "" && spec.AuthType != types.AuthKindAWSFederation {
				return nil, ErrAuthTypeMismatch
			}

			clients := operations.SanitizeClientDescriptors(provider, cfg.clients)
			return &Provider{
				BaseProvider: providers.NewBaseProvider(
					provider,
					types.ProviderCapabilities{
						SupportsRefreshTokens: false,
						SupportsClientPooling: len(clients) > 0,
						SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
					},
					operations.SanitizeOperationDescriptors(provider, cfg.operations),
					clients,
				),
			}, nil
		},
	}
}

// Provider persists AWS STS metadata and exposes it via CredentialSet.
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
}

// BeginAuth is not supported for AWS STS metadata flows.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint validates the stored AWS metadata and persists structured credential fields.
func (p *Provider) Mint(_ context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	meta := subject.Credential.Data.ProviderData
	if len(meta) == 0 {
		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	var decoded awsSTSMetadata
	if err := auth.DecodeProviderData(meta, &decoded); err != nil {
		return types.CredentialPayload{}, err
	}

	if decoded.RoleARN == "" {
		return types.CredentialPayload{}, ErrRoleARNRequired
	}
	if decoded.Region == "" {
		return types.CredentialPayload{}, ErrRegionRequired
	}

	sanitized := maps.Clone(meta)
	sanitized["roleArn"] = decoded.RoleARN
	sanitized["region"] = decoded.Region

	creds := auth.AWSCredentials{
		AccessKeyID:     decoded.AccessKeyID,
		SecretAccessKey: decoded.SecretAccessKey,
		SessionToken:    decoded.SessionToken,
	}

	if decoded.ExternalID != "" {
		sanitized["externalId"] = decoded.ExternalID
	}
	if decoded.SessionName != "" {
		sanitized["sessionName"] = decoded.SessionName
	}
	if decoded.SessionDuration != "" {
		sanitized["sessionDuration"] = decoded.SessionDuration
	}
	if decoded.AccountID != "" {
		sanitized["accountId"] = decoded.AccountID
	}
	if creds.AccessKeyID != "" {
		sanitized["accessKeyId"] = creds.AccessKeyID
	}
	if creds.SecretAccessKey != "" {
		sanitized["secretAccessKey"] = creds.SecretAccessKey
	}
	if creds.SessionToken != "" {
		sanitized["sessionToken"] = creds.SessionToken
	}

	builder := types.NewCredentialBuilder(p.Type()).With(
		types.WithCredentialKind(types.CredentialKindMetadata),
		types.WithCredentialSet(models.CredentialSet{
			AccessKeyID:     creds.AccessKeyID,
			SecretAccessKey: creds.SecretAccessKey,
			SessionToken:    creds.SessionToken,
			ProviderData:    sanitized,
		}),
	)

	return builder.Build()
}

type awsSTSMetadata struct {
	// RoleARN is the role ARN to assume
	RoleARN string `json:"roleArn"`
	// Region is the AWS region for API calls
	Region string `json:"region"`
	// ExternalID is the optional external ID for role assumption
	ExternalID string `json:"externalId"`
	// SessionName is the optional session name for STS
	SessionName string `json:"sessionName"`
	// SessionDuration is the optional session duration string
	SessionDuration string `json:"sessionDuration"`
	// AccountID is the AWS account identifier
	AccountID string `json:"accountId"`
	// AccessKeyID is the AWS access key ID
	AccessKeyID string `json:"accessKeyId"`
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string `json:"secretAccessKey"`
	// SessionToken is the AWS session token
	SessionToken string `json:"sessionToken"`
}
