package awssts

import (
	"context"
	"fmt"
	"maps"

	"github.com/theopenlane/common/integrations/config"
	"github.com/theopenlane/common/integrations/helpers"
	"github.com/theopenlane/common/integrations/types"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/pkg/integrations/providers"
)

// ProviderOption customizes AWS STS providers.
type ProviderOption func(*providerConfig)

type providerConfig struct {
	operations []types.OperationDescriptor
}

// WithOperations registers provider-published operations.
func WithOperations(descriptors []types.OperationDescriptor) ProviderOption {
	return func(cfg *providerConfig) {
		cfg.operations = descriptors
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
				return nil, fmt.Errorf("%w (provider %s expects %s, found %s)", ErrAuthTypeMismatch, provider, types.AuthKindAWSFederation, spec.AuthType)
			}

			return &Provider{
				provider:   provider,
				operations: helpers.SanitizeOperationDescriptors(provider, cfg.operations),
				caps: types.ProviderCapabilities{
					SupportsRefreshTokens: false,
					SupportsClientPooling: false,
					SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
				},
			}, nil
		},
	}
}

// Provider persists AWS STS metadata and exposes it via CredentialSet.
type Provider struct {
	provider   types.ProviderType
	operations []types.OperationDescriptor
	caps       types.ProviderCapabilities
}

// Type returns the provider identifier.
func (p *Provider) Type() types.ProviderType {
	if p == nil {
		return types.ProviderUnknown
	}
	return p.provider
}

// Capabilities returns optional capability flags.
func (p *Provider) Capabilities() types.ProviderCapabilities {
	if p == nil {
		return types.ProviderCapabilities{}
	}
	return p.caps
}

// Operations returns provider-published operations.
func (p *Provider) Operations() []types.OperationDescriptor {
	if p == nil || len(p.operations) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(p.operations))
	copy(out, p.operations)
	return out
}

// BeginAuth is not supported for AWS STS metadata flows.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, fmt.Errorf("%w (provider %s)", ErrBeginAuthNotSupported, p.provider)
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

	roleArn := helpers.StringValue(meta, "roleArn")
	if roleArn == "" {
		return types.CredentialPayload{}, ErrRoleARNRequired
	}

	region := helpers.StringValue(meta, "region")
	if region == "" {
		return types.CredentialPayload{}, ErrRegionRequired
	}

	sanitized := maps.Clone(meta)

	accessKey := helpers.StringValue(meta, "accessKeyId")
	secretKey := helpers.StringValue(meta, "secretAccessKey")
	sessionToken := helpers.StringValue(meta, "sessionToken")

	if accessKey != "" {
		sanitized["accessKeyId"] = accessKey
	}
	if secretKey != "" {
		sanitized["secretAccessKey"] = secretKey
	}
	if sessionToken != "" {
		sanitized["sessionToken"] = sessionToken
	}

	builder := types.NewCredentialBuilder(p.provider).With(
		types.WithCredentialKind(types.CredentialKindMetadata),
		types.WithCredentialSet(models.CredentialSet{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
			ProviderData:    sanitized,
		}),
	)

	return builder.Build()
}

func cloneProviderData(data map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}

	return maps.Clone(data)
}
