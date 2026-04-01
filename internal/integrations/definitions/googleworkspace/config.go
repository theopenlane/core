package googleworkspace

// Config holds operator-level credentials for the Google Workspace definition
type Config struct {
	// ClientID is the Google OAuth application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Google OAuth application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Google application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl"`
}
