package awssts

import (
	"context"
	"maps"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/helpers"
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

			clients := helpers.SanitizeClientDescriptors(provider, cfg.clients)
			return &Provider{
				BaseProvider: providers.NewBaseProvider(
					provider,
					types.ProviderCapabilities{
						SupportsRefreshTokens: false,
						SupportsClientPooling: len(clients) > 0,
						SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
					},
					helpers.SanitizeOperationDescriptors(provider, cfg.operations),
					clients,
				),
			}, nil
		},
	}
}

// Provider persists AWS STS metadata and exposes it via CredentialSet.
type Provider struct {
	providers.BaseProvider
}

// BeginAuth is not supported for AWS STS metadata flows.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint validates the stored AWS metadata and persists structured credential fields.
func (p *Provider) Mint(_ context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	if p == nil {
		return types.CredentialPayload{}, ErrProviderNotInitialized
	}

	meta := cloneProviderData(subject.Credential.Data.ProviderData)
	if len(meta) == 0 {
		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	roleArn, err := helpers.RequiredString(meta, "roleArn", ErrRoleARNRequired)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	region, err := helpers.RequiredString(meta, "region", ErrRegionRequired)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	sanitized := maps.Clone(meta)
	sanitized["roleArn"] = roleArn
	sanitized["region"] = region

	creds := helpers.AWSCredentialsFromProviderData(meta)

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
			ProviderData:    sanitized,
		}),
	)

	return builder.Build()
}

// cloneProviderData returns a shallow copy of provider data
func cloneProviderData(data map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}

	return maps.Clone(data)
}
