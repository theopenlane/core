package oidclocal

import (
	"maps"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the local Dex-backed OIDC definition builder
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		active := cfg.Enabled &&
			cfg.ClientID != "" &&
			cfg.ClientSecret != "" &&
			cfg.DiscoveryURL != "" &&
			cfg.RedirectURL != ""

		return types.Definition{
			ID:          definitionID.ID(),
			Family:      "oidc",
			DisplayName: "Local OIDC (Dex)",
			Description: "Dex-backed local OpenID Connect provider used to exercise integration OAuth start, callback, validation, and credential persistence flows end to end.",
			Category:    "identity",
			Tags:        []string{"oidc", "oauth", "local", "testing"},
			Active:      active,
			Visible:     active,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				oidcCredential.Registration(types.CredentialRegistration{
					Name:        "Local OIDC Credential",
					Description: "Auth-managed OIDC credential issued by the local Dex development provider.",
				}),
			},
			HealthCheck:  types.CredentialHealthCheck(checkHealth),
			Installation: installation.Registration(),
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: oidcCredential.ID(),
					Name:          "Local Dex OIDC",
					Description:   "Connect through the local Dex development provider to test integration OAuth callback flows end to end.",
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[oidcLocalCred]{
						CredentialRef: oidcCredential,
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							DiscoveryURL: cfg.DiscoveryURL,
							RedirectURL:  cfg.RedirectURL,
							Scopes: []string{
								"openid",
								"profile",
								"email",
								"offline_access",
							},
						},
						Material: func(material auth.OAuthMaterial) (oidcLocalCred, error) {
							claims := new(oidc.IDTokenClaims)
							if err := jsonx.RoundTrip(material.Claims, claims); err != nil {
								return oidcLocalCred{}, err
							}

							var groups []string
							if rawGroups, ok := claims.Claims["groups"]; ok {
								_ = jsonx.RoundTrip(rawGroups, &groups)
							}

							return oidcLocalCred{
								AccessToken:       material.AccessToken,
								RefreshToken:      material.RefreshToken,
								Expiry:            material.Expiry,
								Issuer:            claims.Issuer,
								Subject:           claims.Subject,
								Email:             claims.Email,
								Name:              claims.Name,
								PreferredUsername: claims.PreferredUsername,
								Groups:            groups,
								Claims:            maps.Clone(material.Claims),
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: oidcCredential.ID(),
						Description:   "Removes the stored local Dex credential from Openlane.",
					},
				},
			},
			Operations: []types.OperationRegistration{
				types.OperationPayloadOf[ClaimsInspect]().
					HandlesRequest(inspectClaims).
					Policy(types.ExecutionPolicy{Inline: true}).
					Description("Return the raw OIDC ID token claims stored with the auth-managed credential.").
					Registration(definitionID),
			},
		}, nil
	})
}
