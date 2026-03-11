package slack

// Config holds operator-level credentials for the Slack OAuth provider.
type Config struct {
	// ClientID is the Slack OAuth application client identifier.
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Slack OAuth application client secret.
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
}
