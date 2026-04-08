package slack

// Config holds operator-level credentials for the Slack definition
type Config struct {
	// ClientID is the Slack OAuth application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Slack OAuth application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Slack application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"https://api.theopenlane.io/v1/integrations/auth/callback"`
}
