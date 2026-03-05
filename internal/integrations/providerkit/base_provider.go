package providerkit

import (
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

// BaseProviderConfig controls shared capability and descriptor wiring for base providers.
type BaseProviderConfig struct {
	// SupportsRefreshTokens indicates whether provider Mint refresh/exchange is expected.
	SupportsRefreshTokens bool
	// EnvironmentCredentials indicates provider credentials are derived from environment/install context.
	EnvironmentCredentials bool
	// Operations are provider operation descriptors to register.
	Operations []types.OperationDescriptor
	// Clients are provider client descriptors to register.
	Clients []types.ClientDescriptor
}

// NewBaseProvider constructs a providers.BaseProvider with sanitized descriptors and default capabilities.
func NewBaseProvider(provider types.ProviderType, spec config.ProviderSpec, cfg BaseProviderConfig) providers.BaseProvider {
	clients := SanitizeClientDescriptors(provider, cfg.Clients)
	ops := SanitizeOperationDescriptors(provider, cfg.Operations)

	caps := types.ProviderCapabilities{
		SupportsRefreshTokens:  cfg.SupportsRefreshTokens,
		SupportsClientPooling:  len(clients) > 0,
		SupportsMetadataForm:   len(spec.CredentialsSchema) > 0,
		EnvironmentCredentials: cfg.EnvironmentCredentials,
	}

	return providers.NewBaseProvider(provider, caps, ops, clients)
}
