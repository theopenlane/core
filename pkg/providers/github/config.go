package github

// ProviderConfig represents the configuration settings for a Github Oauth Provider
type ProviderConfig struct {
	// ClientID is the public identifier for the GitHub oauth2 client
	ClientID string `json:"clientId" koanf:"clientId" jsonschema:"required"`
	// ClientSecret is the secret for the GitHub oauth2 client
	ClientSecret string `json:"clientSecret" koanf:"clientSecret" jsonschema:"required"`
	// ClientEndpoint is the endpoint for the GitHub oauth2 client
	ClientEndpoint string `json:"clientEndpoint" koanf:"clientEndpoint" default:"http://localhost:17608"`
	// Scopes are the scopes that the GitHub oauth2 client will request
	Scopes []string `json:"scopes" koanf:"scopes" jsonschema:"required"`
	// RedirectURL is the URL that the GitHub oauth2 client will redirect to after authentication with Github
	RedirectURL string `json:"redirectUrl" koanf:"redirectUrl" jsonschema:"required" default:"/v1/github/callback"`
}
