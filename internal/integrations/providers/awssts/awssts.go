package awssts

import (
	"context"

	"github.com/samber/lo"

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

	decoded, err := awsSTSMetadataFromMap(meta)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	creds := decoded.credentials()

	sanitized, err := auth.PersistMetadata(meta, decoded)
	if err != nil {
		return types.CredentialPayload{}, err
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
	RoleARN types.TrimmedString `json:"roleArn,omitempty"`
	// Region is the AWS region for API calls
	Region types.TrimmedString `json:"region,omitempty"`
	// HomeRegion is the Security Hub home region used for aggregated findings
	HomeRegion types.TrimmedString `json:"homeRegion,omitempty"`
	// LinkedRegions optionally limits collection to the provided regions
	LinkedRegions []string `json:"linkedRegions,omitempty"`
	// OrganizationID is the AWS Organizations identifier associated with this integration
	OrganizationID types.TrimmedString `json:"organizationId,omitempty"`
	// AccountScope indicates whether queries should run for all accessible accounts or a specific set
	AccountScope types.LowerString `json:"accountScope,omitempty"`
	// AccountIDs identifies the account IDs to query when accountScope is specific
	AccountIDs []string `json:"accountIds,omitempty"`
	// ExternalID is the optional external ID for role assumption
	ExternalID types.TrimmedString `json:"externalId,omitempty"`
	// SessionName is the optional session name for STS
	SessionName types.TrimmedString `json:"sessionName,omitempty"`
	// SessionDuration is the optional session duration string
	SessionDuration types.TrimmedString `json:"sessionDuration,omitempty"`
	// AccountID is the AWS account identifier
	AccountID types.TrimmedString `json:"accountId,omitempty"`
	// AccessKeyID is the AWS access key ID
	AccessKeyID types.TrimmedString `json:"accessKeyId,omitempty"`
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey types.TrimmedString `json:"secretAccessKey,omitempty"`
	// SessionToken is the AWS session token
	SessionToken types.TrimmedString `json:"sessionToken,omitempty"`
}

func awsSTSMetadataFromMap(meta map[string]any) (awsSTSMetadata, error) {
	var decoded awsSTSMetadata
	if err := auth.DecodeProviderData(meta, &decoded); err != nil {
		return awsSTSMetadata{}, err
	}

	decoded.applyDefaults()

	return decoded, nil
}

// applyDefaults fills in fallback values and deduplicates slice fields.
func (m *awsSTSMetadata) applyDefaults() {
	m.Region = lo.CoalesceOrEmpty(m.Region, m.HomeRegion)
	m.HomeRegion = lo.CoalesceOrEmpty(m.HomeRegion, m.Region)
	m.AccountScope = lo.CoalesceOrEmpty(m.AccountScope, types.LowerString(auth.AWSAccountScopeAll))

	m.AccountIDs = types.NormalizeStringSlice(m.AccountIDs)
	m.LinkedRegions = types.NormalizeStringSlice(m.LinkedRegions)
}

func (m awsSTSMetadata) credentials() auth.AWSCredentials {
	return auth.AWSCredentials{
		AccessKeyID:     m.AccessKeyID.String(),
		SecretAccessKey: m.SecretAccessKey.String(),
		SessionToken:    m.SessionToken.String(),
	}
}
