package microsoftteams

// Config holds operator-level credentials for the Microsoft Teams definition
type Config struct {
	// ClientID is the Azure OAuth application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Azure OAuth application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Azure application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"https://api.theopenlane.io/v1/integrations/auth/callback"`
	// ApplicationID is the application ID registered in azure, used in the well-known configuration for domain validation
	ApplicationID string `json:"applicationid" koanf:"applicationid"`
}
