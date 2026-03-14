package oidcgeneric

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeOIDCGeneric identifies the generic OIDC provider
const TypeOIDCGeneric = types.ProviderType("oidcgeneric")

const (
	// ClientOIDCAPI identifies the OIDC HTTP client used for userinfo calls
	ClientOIDCAPI types.ClientName = "api"
)

// Builder returns the generic OIDC provider builder with the supplied operator config applied.
func Builder(cfg Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeOIDCGeneric,
		SpecFunc:     oidcGenericSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if s.OAuth != nil && cfg.ClientID != "" {
				s.OAuth.ClientID = cfg.ClientID
				s.OAuth.ClientSecret = cfg.ClientSecret
			}

			if s.OAuth != nil && cfg.DiscoveryURL != "" {
				s.OAuth.OIDCDiscovery = cfg.DiscoveryURL
			}

			ops := oidcOperations(userInfoURL(s))
			return oauth.New(s, oauth.WithOperations(ops), oauth.WithClientDescriptors(oidcClientDescriptors()))
		},
	}
}

// oidcGenericSpec returns the static provider specification for the generic OIDC provider.
func oidcGenericSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:             "oidcgeneric",
		DisplayName:      "Generic OIDC",
		Category:         "sso",
		AuthType:         types.AuthKindOIDC,
		AuthStartPath:    "/v1/integrations/oauth/start",
		AuthCallbackPath: "/v1/integrations/oauth/callback",
		Active:           lo.ToPtr(true),
		Visible:          lo.ToPtr(false),
		LogoURL:          "https://static-00.iconduck.com/assets.00/openid-icon-512x512-5t0u9s21.png",
		DocsURL:          "https://docs.theopenlane.io/docs/platform/integrations/oidc_generic/overview",
		OAuth: &spec.OAuthSpec{
			Scopes:      []string{"openid", "profile", "email", "offline_access"},
			RedirectURI: "https://api.theopenlane.io/v1/integrations/oauth/callback",
		},
		UserInfo: &spec.UserInfoSpec{
			URL:       "https://openidconnect.googleapis.com/v1/userinfo",
			Method:    "GET",
			AuthStyle: "Bearer",
			IDPath:    "sub",
			EmailPath: "email",
			LoginPath: "name",
		},
		Persistence: &spec.PersistenceSpec{
			StoreRefreshToken: true,
		},
		Description: "Connect Openlane to standards-based OpenID Connect identity providers for federated authentication and claim inspection.",
	}
}

// userInfoURL returns the configured userinfo endpoint when present
func userInfoURL(s spec.ProviderSpec) string {
	if s.UserInfo != nil {
		return s.UserInfo.URL
	}

	return ""
}

// oidcClientDescriptors returns the client descriptors published by the generic OIDC provider
func oidcClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeOIDCGeneric, ClientOIDCAPI, "OIDC userinfo HTTP client", providerkit.TokenClientBuilder(providerkit.OAuthTokenFromCredential, nil))
}

// oidcOperations returns OIDC operation descriptors
func oidcOperations(infoURL string) []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(types.OperationHealthDefault, "Call the configured userinfo endpoint (when available) to validate the OIDC token.", ClientOIDCAPI, runOIDCHealth(infoURL)),
		{
			Name:        types.OperationName("claims.inspect"),
			Kind:        types.OperationKindScanSettings,
			Description: "Expose stored ID token claims for downstream checks.",
			Run:         runOIDCClaims,
		},
	}
}

// runOIDCHealth builds a health check function for OIDC tokens
func runOIDCHealth(infoURL string) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		client, err := providerkit.ResolveAuthenticatedClient(input, providerkit.OAuthTokenFromCredential, "", nil)
		if err != nil {
			return types.OperationResult{}, err
		}

		if infoURL == "" {
			return types.OperationResult{
				Status:  types.OperationStatusOK,
				Summary: "OIDC token present (no userinfo endpoint configured)",
			}, nil
		}

		var resp map[string]any
		if err := client.GetJSON(ctx, infoURL, &resp); err != nil {
			return providerkit.OperationFailure("OIDC userinfo call failed", err, nil)
		}

		summary := "OIDC userinfo call succeeded"
		if subject, ok := resp["sub"].(string); ok {
			summary = fmt.Sprintf("OIDC userinfo succeeded for %s", subject)
		}

		return providerkit.OperationSuccess(summary, resp), nil
	}
}

// runOIDCClaims returns stored OIDC claims for inspection
func runOIDCClaims(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
	claims := input.Credential.Claims
	if claims == nil {
		return types.OperationResult{
			Status:  types.OperationStatusUnknown,
			Summary: "No OIDC claims stored",
		}, nil
	}

	return providerkit.OperationSuccess("OIDC claims available", struct {
		Claims map[string]any `json:"claims"`
	}{
		Claims: claims,
	}), nil
}
