package githubapp

// Config holds operator-level credentials for the GitHub App definition
type Config struct {
	// AppID is the GitHub App identifier
	AppID string `json:"appid" koanf:"appid" sensitive:"true"`
	// PrivateKey is the PEM-encoded RSA private key for the GitHub App
	PrivateKey string `json:"privatekey" koanf:"privatekey" sensitive:"true"`
	// WebhookSecret is the GitHub App webhook secret for signature verification
	WebhookSecret string `json:"webhooksecret" koanf:"webhooksecret" sensitive:"true"`
	// AppSlug is the GitHub App slug identifier
	AppSlug string `json:"appslug" koanf:"appslug"`
	// APIURL overrides the GitHub API host for local tests
	APIURL string `json:"-" koanf:"-"`
}
