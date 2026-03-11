package oidcgeneric

// Config holds operator-level credentials for the generic OIDC provider.
type Config struct {
	// ClientID is the OIDC application client identifier.
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the OIDC application client secret.
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// DiscoveryURL is the OIDC discovery endpoint for the provider.
	DiscoveryURL string `json:"discoveryurl" koanf:"discoveryurl"`
}
