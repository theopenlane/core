package apikey

import (
	"context"
	"maps"
	"strings"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
)

// ProviderOption customizes API key providers.
type ProviderOption func(*providerConfig)

type providerConfig struct {
	tokenField string
	operations []types.OperationDescriptor
	clients    []types.ClientDescriptor
}

// WithTokenField overrides the metadata field used to extract the API token.
func WithTokenField(field string) ProviderOption {
	return func(cfg *providerConfig) {
		field = strings.TrimSpace(field)
		if field != "" {
			cfg.tokenField = field
		}
	}
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

// Builder returns a providers.Builder that constructs API key providers.
func Builder(provider types.ProviderType, opts ...ProviderOption) providers.Builder {
	cfg := providerConfig{
		tokenField: "apiToken",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return providers.BuilderFunc{
		ProviderType: provider,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.AuthType != "" && spec.AuthType != types.AuthKindAPIKey {
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
				tokenField: cfg.tokenField,
			}, nil
		},
	}
}

// Provider implements API key based integrations.
type Provider struct {
	providers.BaseProvider
	tokenField string
}

// BeginAuth is not supported for API key providers.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint materializes a stored API key configuration into a credential payload.
func (p *Provider) Mint(_ context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	if p == nil {
		return types.CredentialPayload{}, ErrProviderNotInitialized
	}

	providerData := subject.Credential.Data.ProviderData
	if len(providerData) == 0 {
		if token := strings.TrimSpace(subject.Credential.Data.APIToken); token != "" {
			return subject.Credential, nil
		}

		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	token := helpers.StringFromAny(providerData[p.tokenField])
	if token == "" {
		return types.CredentialPayload{}, ErrTokenFieldRequired
	}

	cloned := maps.Clone(providerData)

	builder := types.NewCredentialBuilder(p.Type()).With(
		types.WithCredentialKind(types.CredentialKindAPIKey),
		types.WithCredentialSet(models.CredentialSet{
			APIToken:     token,
			ProviderData: cloned,
		}),
	)

	return builder.Build()
}
