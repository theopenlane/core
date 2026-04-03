package oidclocal

// Config holds operator-level configuration for the local Dex-backed OIDC definition
type Config struct {
	// Enabled controls whether the local OIDC test definition is exposed in the catalog
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// ClientID is the OIDC application client identifier configured in Dex
	ClientID string `json:"clientid" koanf:"clientid" default:"local-core-oidc"`
	// ClientSecret is the OIDC application client secret configured in Dex
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true" default:"local-core-oidc-secret"`
	// DiscoveryURL is the OIDC issuer URL used for endpoint discovery
	DiscoveryURL string `json:"discoveryurl" koanf:"discoveryurl" default:"http://localhost:5557/dex"`
	// RedirectURL is the OAuth callback URL registered with the OIDC provider
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"http://localhost:17608/v1/integrations/auth/callback"`
}
