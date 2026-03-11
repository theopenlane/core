package azuresecuritycenter

// Config holds operator-level credentials for the Azure Security Center OAuth provider.
type Config struct {
	// ClientID is the Azure OAuth application client identifier.
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Azure OAuth application client secret.
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
}
