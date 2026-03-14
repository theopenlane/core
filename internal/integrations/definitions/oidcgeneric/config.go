package oidcgeneric

// Config holds operator-level credentials for the Generic OIDC definition
type Config struct {
	// ClientID is the OIDC application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the OIDC application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// DiscoveryURL is the OIDC issuer URL used for endpoint discovery
	DiscoveryURL string `json:"discoveryurl" koanf:"discoveryurl"`
	// RedirectURL is the OAuth callback URL registered with the OIDC provider
	RedirectURL string `json:"redirecturl" koanf:"redirecturl"`
}
