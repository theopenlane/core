package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
	slackOAuth2 "golang.org/x/oauth2/slack"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// IntegrationOauthProviderConfig represents the configuration for OAuth providers used for integrations
type IntegrationOauthProviderConfig struct {
	// RedirectURL is the base URL for integration OAuth callbacks
	RedirectURL string `json:"redirecturl" koanf:"redirecturl"`
	// SuccessRedirectURL is the URL to redirect to after successful OAuth integration
	SuccessRedirectURL string `json:"successredirecturl" koanf:"successredirecturl" domain:"inherit" domainPrefix:"https://console" domainSuffix:"/organization-settings/integrations"`
	// Github contains the configuration settings for GitHub integrations
	Github IntegrationProviderConfig `json:"github" koanf:"github"`
	// Slack contains the configuration settings for Slack integrations
	Slack IntegrationProviderConfig `json:"slack" koanf:"slack"`
}

// IntegrationProviderConfig contains OAuth configuration for a specific integration provider
type IntegrationProviderConfig struct {
	// ClientID is the OAuth2 client ID
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the OAuth2 client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// ClientEndpoint is the base URL for the OAuth endpoints
	ClientEndpoint string `json:"clientendpoint" koanf:"clientendpoint" domain:"inherit" domainPrefix:"https://api"`
	// Scopes are the OAuth2 scopes to request
	Scopes []string `json:"scopes" koanf:"scopes"`
}

// IntegrationProvider represents a supported OAuth provider for integrations
type IntegrationProvider struct {
	Name     string
	Config   *oauth2.Config
	Validate func(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error)
}

// IntegrationUserInfo contains user information from OAuth provider
type IntegrationUserInfo struct {
	ID       string
	Username string
	Email    string
}

// GitHubUser represents GitHub user data from API
type GitHubUser struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Blog     string `json:"blog"`
	Bio      string `json:"bio"`
}

// GitHubEmail represents GitHub email data from API
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// SlackUser represents Slack user data from API
type SlackUser struct {
	OK    bool          `json:"ok"`
	User  SlackUserInfo `json:"user"`
	Error string        `json:"error,omitempty"`
}

// SlackUserInfo contains Slack user profile information
type SlackUserInfo struct {
	ID       string           `json:"id"`
	TeamID   string           `json:"team_id"`
	Name     string           `json:"name"`
	Deleted  bool             `json:"deleted"`
	RealName string           `json:"real_name"`
	Profile  SlackUserProfile `json:"profile"`
}

// SlackUserProfile contains Slack user profile details
type SlackUserProfile struct {
	Email                 string `json:"email"`
	DisplayName           string `json:"display_name"`
	DisplayNameNormalized string `json:"display_name_normalized"`
	RealName              string `json:"real_name"`
	RealNameNormalized    string `json:"real_name_normalized"`
}

var (
	// ErrInvalidState is returned when OAuth state validation fails
	ErrInvalidState = fmt.Errorf("invalid OAuth state parameter")
	// ErrMissingCode is returned when OAuth authorization code is missing
	ErrMissingCode = fmt.Errorf("missing OAuth authorization code")
	// ErrExchangeAuthCode is returned when OAuth code exchange fails
	ErrExchangeAuthCode = fmt.Errorf("failed to exchange authorization code")
	// ErrValidateToken is returned when OAuth token validation fails
	ErrValidateToken = fmt.Errorf("failed to validate OAuth token")
	// ErrInvalidStateFormat is returned when OAuth state format is invalid
	ErrInvalidStateFormat = fmt.Errorf("invalid state format")
	// ErrProviderRequired is returned when provider parameter is missing
	ErrProviderRequired = fmt.Errorf("provider parameter is required")
	// ErrIntegrationIDRequired is returned when integration ID is missing
	ErrIntegrationIDRequired = fmt.Errorf("integration ID is required")
	// ErrIntegrationNotFound is returned when integration is not found
	ErrIntegrationNotFound = fmt.Errorf("integration not found")
	// ErrDeleteSecrets is returned when deleting integration secrets fails
	ErrDeleteSecrets = fmt.Errorf("failed to delete integration secrets")
)

// OAuth Integration Constants
const (
	// URL patterns
	oauthCallbackPath = "/v1/integrations/oauth/callback"

	// Integration naming patterns
	integrationNameSuffix = " Integration"
	integrationDescPrefix = "OAuth integration with "

	// Secret naming patterns
	secretNameSeparator = "_"
	secretDescSeparator = " "
	secretDescSuffix    = " for %s integration" // nolint:gosec

	// Token field names
	accessTokenField      = "access_token"
	refreshTokenField     = "refresh_token"
	expiresAtField        = "expires_at"
	providerUserIDField   = "provider_user_id"
	providerUsernameField = "provider_username"
	providerEmailField    = "provider_email"

	// Status messages
	statusConnected    = "connected"
	statusInvalid      = "invalid"
	statusExpired      = "expired"
	statusNotConnected = "not_connected"

	// OAuth token kind
	oauthTokenKind = "oauth_token" // nolint:gosec
)

// IntegrationHelper provides helper methods for integration operations
type IntegrationHelper struct {
	provider string
	username string
}

// NewIntegrationHelper creates a new integration helper
func NewIntegrationHelper(provider, username string) *IntegrationHelper {
	return &IntegrationHelper{
		provider: provider,
		username: username,
	}
}

// Name returns the integration name
func (ih *IntegrationHelper) Name() string {
	name := cases.Title(language.English).String(ih.provider) + integrationNameSuffix
	if ih.username != "" {
		return name + " (" + ih.username + ")"
	}

	return name
}

// Description returns the integration description
func (ih *IntegrationHelper) Description() string {
	desc := integrationDescPrefix + ih.provider
	if ih.username != "" {
		desc += " for " + ih.username
	}

	return desc
}

// SecretName returns the secret name for a given field
func (ih *IntegrationHelper) SecretName(fieldName string) string {
	return ih.provider + secretNameSeparator + fieldName
}

// SecretDisplayName returns the display name for a secret
func (ih *IntegrationHelper) SecretDisplayName(integrationName, fieldName string) string {
	return integrationName + secretDescSeparator + strings.ReplaceAll(fieldName, secretNameSeparator, secretDescSeparator)
}

// SecretDescription returns the description for a secret
func (ih *IntegrationHelper) SecretDescription(fieldName string) string {
	displayName := cases.Title(language.English).String(strings.ReplaceAll(fieldName, secretNameSeparator, secretDescSeparator))

	return fmt.Sprintf(displayName+secretDescSuffix, ih.provider)
}

// CallbackURL returns the OAuth callback URL
func (ih *IntegrationHelper) CallbackURL(baseURL string) string {
	return baseURL + oauthCallbackPath
}

// RedirectURL returns the success redirect URL with parameters
func (ih *IntegrationHelper) RedirectURL(baseURL string) string {
	message := "Successfully connected " + cases.Title(language.English).String(ih.provider) + integrationNameSuffix

	return baseURL + "?provider=" + ih.provider + "&status=success&message=" + message
}

// StateData returns the OAuth state data
func (ih *IntegrationHelper) StateData(orgID string, randomBytes []byte) string {
	return orgID + ":" + ih.provider + ":" + base64.URLEncoding.EncodeToString(randomBytes)
}

// StatusMessage returns status messages for integration status
func (ih *IntegrationHelper) StatusMessage(status string) string {
	providerTitle := cases.Title(language.English).String(ih.provider)

	switch status {
	case statusConnected:
		return providerTitle + " integration is connected and active"
	case statusInvalid:
		return providerTitle + " integration exists but has invalid tokens"
	case statusExpired:
		return providerTitle + " integration tokens have expired"
	case statusNotConnected:
		return "No " + ih.provider + " integration found"
	default:
		return providerTitle + " integration status unknown"
	}
}

// AuthHeader returns the appropriate authorization header for the provider
func (ih *IntegrationHelper) AuthHeader(token string) (string, string) {
	switch ih.provider {
	case "github":
		return "Authorization", "token " + token
	case "slack":
		return "Authorization", "Bearer " + token
	default:
		return "Authorization", "Bearer " + token
	}
}

// getIntegrationProviders returns available OAuth providers for integrations
func (h *Handler) getIntegrationProviders() map[string]IntegrationProvider {
	providers := map[string]IntegrationProvider{
		"github": {
			Name:   "github",
			Config: h.getGithubIntegrationConfig(),
			Validate: func(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
				return h.validateGithubIntegrationToken(ctx, token)
			},
		},
	}

	// Only add Slack if configuration is available
	if slackConfig := h.getSlackIntegrationConfig(); slackConfig != nil {
		providers["slack"] = IntegrationProvider{
			Name:   "slack",
			Config: slackConfig,
			Validate: func(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
				return h.validateSlackIntegrationToken(ctx, token)
			},
		}
	}

	return providers
}

// getGithubIntegrationConfig returns OAuth2 config for GitHub integrations
func (h *Handler) getGithubIntegrationConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.IntegrationOauthProvider.Github.ClientID,
		ClientSecret: h.IntegrationOauthProvider.Github.ClientSecret,
		RedirectURL:  NewIntegrationHelper("github", "").CallbackURL(h.IntegrationOauthProvider.Github.ClientEndpoint),
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       []string{"read:user", "user:email", "repo"}, // Extended scopes for API access
	}
}

// getSlackIntegrationConfig returns OAuth2 config for Slack integrations
func (h *Handler) getSlackIntegrationConfig() *oauth2.Config {
	// Check if Slack configuration is available
	if h.IntegrationOauthProvider.Slack.ClientID == "" || h.IntegrationOauthProvider.Slack.ClientSecret == "" {
		return nil
	}

	helper := NewIntegrationHelper("slack", "")

	return &oauth2.Config{
		ClientID:     h.IntegrationOauthProvider.Slack.ClientID,
		ClientSecret: h.IntegrationOauthProvider.Slack.ClientSecret,
		RedirectURL:  helper.CallbackURL(h.IntegrationOauthProvider.Slack.ClientEndpoint),
		Endpoint:     slackOAuth2.Endpoint,
		Scopes:       h.IntegrationOauthProvider.Slack.Scopes, // Default Slack scopes for integrations
	}
}
