package githuboauth

// Config holds operator-level credentials for the GitHub OAuth definition
type Config struct {
	// ClientID is the GitHub OAuth application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the GitHub OAuth application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// BaseURL overrides the GitHub API base URL (for GitHub Enterprise)
	BaseURL string `json:"baseurl" koanf:"baseurl"`
	// RedirectURL is the OAuth callback URL registered with the GitHub application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl"`
}
