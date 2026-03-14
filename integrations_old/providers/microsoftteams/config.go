package microsoftteams

// Config holds operator-level credentials for the Microsoft Teams OAuth provider.
type Config struct {
	// ClientID is the Microsoft OAuth application client identifier.
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Microsoft OAuth application client secret.
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
}
