package scim

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Provider implements types.Provider for SCIM 2.0 push-based integrations.
// SCIM is push-based: the external IdP calls Openlane's SCIM endpoint; no outbound
// credentials are stored and auth kind is AuthKindNone.
type Provider struct {
	providers.BaseProvider
}

// scimCapabilities are the capability flags for the SCIM provider
var scimCapabilities = types.ProviderCapabilities{
	SupportsRefreshTokens:  false,
	SupportsClientPooling:  false,
	SupportsMetadataForm:   false,
	EnvironmentCredentials: false,
}

// Builder returns a providers.BuilderFunc that constructs the SCIM provider from a spec
func Builder() providers.BuilderFunc {
	return providers.BuilderFunc{
		ProviderType: types.ProviderSCIM,
		SpecFunc:     scimSpec,
		BuildFunc: func(_ context.Context, _ spec.ProviderSpec) (types.Provider, error) {
			return &Provider{
				BaseProvider: providers.NewBaseProvider(types.ProviderSCIM, scimCapabilities, nil, nil),
			}, nil
		},
	}
}

// scimSpec returns the static provider specification for the SCIM 2.0 provider.
func scimSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "scim",
		DisplayName: "SCIM 2.0",
		Category:    "identity",
		AuthType:    types.AuthKindNone,
		Active:      lo.ToPtr(true),
		Visible:     lo.ToPtr(true),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/scim/overview",
		Tags:        []string{"scim", "provisioning", "identity"},
		Labels: map[string]string{
			"protocol": "scim2",
		},
		Description: "Enable SCIM 2.0 push-based provisioning so external identity providers can sync users and groups into Openlane.",
	}
}

// BeginAuth returns ErrBeginAuthNotSupported because SCIM is push-based
func (p *Provider) BeginAuth(_ context.Context, _ types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint returns ErrMintNotSupported because SCIM credentials are org API tokens provisioned externally
func (p *Provider) Mint(_ context.Context, _ types.CredentialMintRequest) (types.CredentialSet, error) {
	return types.CredentialSet{}, ErrMintNotSupported
}
