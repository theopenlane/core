package apikey

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/helpers"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/models"
)

// ProviderOption customizes API key providers.
type ProviderOption func(*providerConfig)

type providerConfig struct {
	tokenField string
	operations []types.OperationDescriptor
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
				return nil, fmt.Errorf("%w (provider %s expects %s, found %s)", ErrAuthTypeMismatch, provider, types.AuthKindAPIKey, spec.AuthType)
			}

			return &Provider{
				provider:   provider,
				tokenField: cfg.tokenField,
				caps: types.ProviderCapabilities{
					SupportsRefreshTokens: false,
					SupportsClientPooling: false,
					SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
				},
				operations: helpers.SanitizeOperationDescriptors(provider, cfg.operations),
			}, nil
		},
	}
}

// Provider implements API key based integrations.
type Provider struct {
	provider   types.ProviderType
	tokenField string
	caps       types.ProviderCapabilities
	operations []types.OperationDescriptor
}

// Type returns the provider identifier.
func (p *Provider) Type() types.ProviderType {
	if p == nil {
		return types.ProviderUnknown
	}
	return p.provider
}

// Capabilities exposes optional provider behaviour flags.
func (p *Provider) Capabilities() types.ProviderCapabilities {
	if p == nil {
		return types.ProviderCapabilities{}
	}
	return p.caps
}

// Operations returns provider-published operations when configured.
func (p *Provider) Operations() []types.OperationDescriptor {
	if p == nil || len(p.operations) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(p.operations))
	copy(out, p.operations)
	return out
}

// BeginAuth is not supported for API key providers.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, fmt.Errorf("%w (provider %s)", ErrBeginAuthNotSupported, p.provider)
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

	token := strings.TrimSpace(fmt.Sprint(providerData[p.tokenField]))
	if token == "" {
		return types.CredentialPayload{}, fmt.Errorf("%w: %s", ErrTokenFieldRequired, p.tokenField)
	}

	cloned := maps.Clone(providerData)

	builder := types.NewCredentialBuilder(p.provider).With(
		types.WithCredentialKind(types.CredentialKindAPIKey),
		types.WithCredentialSet(models.CredentialSet{
			APIToken:     token,
			ProviderData: cloned,
		}),
	)

	return builder.Build()
}

func sanitizeOperationDescriptors(provider types.ProviderType, descriptors []types.OperationDescriptor) []types.OperationDescriptor {
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.Run == nil {
			continue
		}

		if descriptor.Name == "" {
			continue
		}

		if descriptor.Provider == types.ProviderUnknown {
			descriptor.Provider = provider
		}

		out = append(out, descriptor)
	}

	return out
}
