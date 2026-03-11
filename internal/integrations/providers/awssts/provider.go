package awssts

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// AWSAccountScopeAll indicates operations should run across all accessible accounts
	AWSAccountScopeAll = "all"
	// AWSAccountScopeSpecific indicates operations should be limited to explicitly listed accounts
	AWSAccountScopeSpecific = "specific"
)

// ProviderOption customizes AWS STS providers
type ProviderOption func(*providerConfig)

type providerConfig struct {
	operations []types.OperationDescriptor
	clients    []types.ClientDescriptor
}

// WithOperations registers provider-published operations
func WithOperations(descriptors []types.OperationDescriptor) ProviderOption {
	return func(cfg *providerConfig) {
		cfg.operations = descriptors
	}
}

// WithClientDescriptors registers client descriptors for pooling
func WithClientDescriptors(descriptors []types.ClientDescriptor) ProviderOption {
	return func(cfg *providerConfig) {
		cfg.clients = descriptors
	}
}

// Builder returns a providers.Builder that materializes AWS federation metadata
func Builder(provider types.ProviderType, opts ...ProviderOption) providers.Builder {
	cfg := &providerConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	caps := types.ProviderCapabilities{
		SupportsRefreshTokens:  false,
		SupportsClientPooling:  true,
		SupportsMetadataForm:   true,
		EnvironmentCredentials: false,
	}

	return providerkit.Builder(provider, func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
		if err := providerkit.ValidateAuthType(s, types.AuthKindAWSFederation, ErrAuthTypeMismatch); err != nil {
			return nil, err
		}

		return &Provider{
			BaseProvider: providers.NewBaseProvider(provider, caps, cfg.operations, cfg.clients),
		}, nil
	})
}

// Provider persists AWS STS metadata and exposes it via CredentialSet
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
}

// BeginAuth is not supported for AWS STS metadata flows
func (p *Provider) BeginAuth(_ context.Context, _ types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint validates the stored AWS metadata and persists structured credential fields
func (p *Provider) Mint(_ context.Context, subject types.CredentialMintRequest) (types.CredentialSet, error) {
	decoded, err := awsSTSMetadataFromPayload(subject.Credential)
	if err != nil {
		return types.CredentialSet{}, err
	}

	combined := combinedProviderData{
		awsSTSProviderData: decoded.providerData(),
		AccessKeyID:        decoded.AccessKeyID,
		SecretAccessKey:    decoded.SecretAccessKey,
		SessionToken:       decoded.SessionToken,
	}

	providerDataRaw, err := jsonx.ToRawMessage(combined)
	if err != nil {
		return types.CredentialSet{}, err
	}

	return types.CredentialSet{
		ProviderData: providerDataRaw,
	}, nil
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
	AccessKeyID string `json:"accessKeyId,omitempty"`
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	// SessionToken is the AWS session token
	SessionToken string `json:"sessionToken,omitempty"`
}

func awsSTSMetadataFromPayload(payload types.CredentialSet) (awsSTSMetadata, error) {
	if len(payload.ProviderData) == 0 {
		return awsSTSMetadata{}, ErrProviderMetadataRequired
	}

	var decoded awsSTSMetadata
	if err := jsonx.UnmarshalIfPresent(payload.ProviderData, &decoded); err != nil {
		return awsSTSMetadata{}, err
	}

	decoded.applyDefaults()

	return decoded, nil
}

// applyDefaults fills in fallback values and deduplicates slice fields
func (m *awsSTSMetadata) applyDefaults() {
	m.Region = lo.CoalesceOrEmpty(m.Region, m.HomeRegion)
	m.HomeRegion = lo.CoalesceOrEmpty(m.HomeRegion, m.Region)
	m.AccountScope = lo.CoalesceOrEmpty(m.AccountScope, types.LowerString(AWSAccountScopeAll))

	m.AccountIDs = types.NormalizeStringSlice(m.AccountIDs)
	m.LinkedRegions = types.NormalizeStringSlice(m.LinkedRegions)
}

// awsSTSProviderData holds the non-credential AWS metadata fields stored in ProviderData
type awsSTSProviderData struct {
	RoleARN         string   `json:"roleArn,omitempty"`
	Region          string   `json:"region,omitempty"`
	HomeRegion      string   `json:"homeRegion,omitempty"`
	LinkedRegions   []string `json:"linkedRegions,omitempty"`
	OrganizationID  string   `json:"organizationId,omitempty"`
	AccountScope    string   `json:"accountScope,omitempty"`
	AccountIDs      []string `json:"accountIds,omitempty"`
	ExternalID      string   `json:"externalId,omitempty"`
	SessionName     string   `json:"sessionName,omitempty"`
	SessionDuration string   `json:"sessionDuration,omitempty"`
	AccountID       string   `json:"accountId,omitempty"`
}

func (m awsSTSMetadata) providerData() awsSTSProviderData {
	return awsSTSProviderData{
		RoleARN:         m.RoleARN.String(),
		Region:          m.Region.String(),
		HomeRegion:      m.HomeRegion.String(),
		LinkedRegions:   m.LinkedRegions,
		OrganizationID:  m.OrganizationID.String(),
		AccountScope:    m.AccountScope.String(),
		AccountIDs:      m.AccountIDs,
		ExternalID:      m.ExternalID.String(),
		SessionName:     m.SessionName.String(),
		SessionDuration: m.SessionDuration.String(),
		AccountID:       m.AccountID.String(),
	}
}

// combinedProviderData embeds awsSTSProviderData together with the credential fields
// so that both metadata and access keys are serialized into a single ProviderData blob
type combinedProviderData struct {
	awsSTSProviderData
	// AccessKeyID is the AWS access key ID
	AccessKeyID string `json:"accessKeyId,omitempty"`
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	// SessionToken is the AWS session token
	SessionToken string `json:"sessionToken,omitempty"`
}
