package apikey

import (
	"context"
	"strings"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Option customizes API key providers
type Option func(*providerConfig)

type providerConfig struct {
	tokenField string
	operations []types.OperationDescriptor
	clients    []types.ClientDescriptor
}

// WithTokenField overrides the metadata field used to extract the API token
func WithTokenField(field string) Option {
	return func(cfg *providerConfig) {
		if field != "" {
			cfg.tokenField = field
		}
	}
}

// WithOperations registers provider-published operations
func WithOperations(descriptors []types.OperationDescriptor) Option {
	return func(cfg *providerConfig) {
		cfg.operations = descriptors
	}
}

// WithClientDescriptors registers client descriptors for pooling
func WithClientDescriptors(descriptors []types.ClientDescriptor) Option {
	return func(cfg *providerConfig) {
		cfg.clients = descriptors
	}
}

// Builder returns a providers.Builder that constructs API key providers
func Builder(provider types.ProviderType, opts ...Option) providers.Builder {
	cfg := providerConfig{
		tokenField: "apiToken",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return providerkit.Builder(provider, func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
		if err := providerkit.ValidateAuthType(s, types.AuthKindAPIKey, ErrAuthTypeMismatch); err != nil {
			return nil, err
		}

		return &Provider{
			BaseProvider: providers.NewBaseProvider(
				provider,
				types.ProviderCapabilities{SupportsRefreshTokens: false},
				cfg.operations,
				cfg.clients,
			),
			tokenField: cfg.tokenField,
		}, nil
	})
}

// Provider implements API key based integrations
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
	tokenField string
}

// BeginAuth is not supported for API key providers
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint validates and returns the stored API key credential set.
// The API token must be present in ProviderData under the configured token field.
// The returned CredentialSet preserves the full ProviderData (including the token)
// so that APITokenFromCredential can extract it for downstream operations.
func (p *Provider) Mint(_ context.Context, subject types.CredentialMintRequest) (types.CredentialSet, error) {
	metadata, err := jsonx.ToRawMap(subject.Credential.ProviderData)
	if err != nil {
		return types.CredentialSet{}, err
	}

	var token string
	if raw := metadata[p.tokenField]; len(raw) > 0 {
		if roundTripErr := jsonx.RoundTrip(raw, &token); roundTripErr == nil {
			token = strings.TrimSpace(token)
		}
	}

	if token == "" {
		if len(metadata) == 0 {
			return types.CredentialSet{}, ErrProviderMetadataRequired
		}

		return types.CredentialSet{}, ErrTokenFieldRequired
	}

	// Ensure the token is stored under "apiToken" in ProviderData so APITokenFromCredential
	// can locate it regardless of the configured tokenField name
	if p.tokenField != "apiToken" {
		delete(metadata, p.tokenField)

		tokenRaw, encErr := jsonx.ToRawMessage(token)
		if encErr != nil {
			return types.CredentialSet{}, encErr
		}

		metadata["apiToken"] = tokenRaw
	}

	providerDataRaw, err := jsonx.ToRawMessage(metadata)
	if err != nil {
		return types.CredentialSet{}, err
	}

	return types.CredentialSet{ProviderData: providerDataRaw}, nil
}
