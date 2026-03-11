package github

import "time"

// Config holds operator-level credentials for the GitHub OAuth provider.
type Config struct {
	// ClientID is the GitHub OAuth application client identifier.
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the GitHub OAuth application client secret.
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
}

// AppConfig holds operator-level credentials for the GitHub App provider.
type AppConfig struct {
	// AppID is the GitHub App identifier.
	AppID string `json:"appid" koanf:"appid" sensitive:"true"`
	// PrivateKey is the PEM-encoded RSA private key for the GitHub App.
	PrivateKey string `json:"privatekey" koanf:"privatekey" sensitive:"true"`
	// WebhookSecret is the GitHub App webhook secret for signature verification.
	WebhookSecret string `json:"webhooksecret" koanf:"webhooksecret" sensitive:"true"`
	// InstallationID is the GitHub App installation identifier.
	InstallationID string `json:"installationid" koanf:"installationid"`
	// BaseURL overrides the GitHub API base URL (for GitHub Enterprise).
	BaseURL string `json:"baseurl" koanf:"baseurl"`
	// TokenTTL optionally overrides the installation token lifetime.
	TokenTTL time.Duration `json:"tokenttl" koanf:"tokenttl"`
	// AppSlug is the GitHub App slug identifier.
	AppSlug string `json:"appslug" koanf:"appslug"`
}
