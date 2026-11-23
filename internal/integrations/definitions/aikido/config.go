package aikido

// Config holds operator-level credentials for the Openlane Aikido OAuth app
type Config struct {
	// ClientID is the Aikido OAuth application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Aikido OAuth application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Aikido application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"https://api.theopenlane.io/v1/integrations/auth/callback"`
}
