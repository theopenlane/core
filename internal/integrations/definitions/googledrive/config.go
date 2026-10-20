package googledrive

// AppConfig holds the credentials and callback of one Google OAuth application
type AppConfig struct {
	// ClientID is the Google OAuth application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Google OAuth application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Google application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"https://api.theopenlane.io/v1/integrations/auth/callback"`
}

// Config holds operator-level credentials for the Google Drive definition
type Config struct {
	AppConfig `koanf:",squash"`
	// PickerAPIKey is the browser API key used by the Google Picker
	PickerAPIKey string `json:"pickerapikey" koanf:"pickerapikey"`
	// Legacy is the retired drive.readonly OAuth app that installations created before the drive.file migration still use
	Legacy AppConfig `json:"legacy" koanf:"legacy"`
}
