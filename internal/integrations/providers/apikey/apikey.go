package apikey

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
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

	return providerkit.Builder(provider, func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
		if err := providerkit.ValidateAuthType(spec, types.AuthKindAPIKey, ErrAuthTypeMismatch); err != nil {
			return nil, err
		}

		return &Provider{
			BaseProvider: providerkit.NewBaseProvider(provider, spec, providerkit.BaseProviderConfig{
				SupportsRefreshTokens: false,
				Operations:            cfg.operations,
				Clients:               cfg.clients,
			}),
			tokenField: cfg.tokenField,
		}, nil
	})
}

// Provider implements API key based integrations.
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
	tokenField string
}

// BeginAuth is not supported for API key providers.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint materializes a stored API key configuration into a credential set.
func (p *Provider) Mint(_ context.Context, subject types.CredentialMintRequest) (models.CredentialSet, error) {
	var metadata map[string]any
	if len(subject.Credential.ProviderData) > 0 {
		if err := json.Unmarshal(subject.Credential.ProviderData, &metadata); err != nil {
			return models.CredentialSet{}, err
		}
	}

	token, _ := metadata[p.tokenField].(string)
	token = strings.TrimSpace(token)

	if token == "" {
		token = strings.TrimSpace(subject.Credential.APIToken)
		if token == "" {
			if len(metadata) == 0 {
				return models.CredentialSet{}, ErrProviderMetadataRequired
			}

			return models.CredentialSet{}, ErrTokenFieldRequired
		}
	}

	delete(metadata, p.tokenField)

	var providerDataRaw json.RawMessage
	if len(metadata) > 0 {
		raw, err := json.Marshal(metadata)
		if err != nil {
			return models.CredentialSet{}, err
		}
		providerDataRaw = raw
	}

	return auth.BuildAPITokenCredentialSet(token, providerDataRaw), nil
}
