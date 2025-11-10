package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/keystore"
)

var (
	errIntegrationBrokerNotConfigured = errors.New("integration broker not configured")
	errIntegrationStoreNotConfigured  = errors.New("integration store not configured")
)

// IntegrationOauthProviderConfig represents the configuration for OAuth providers used for integrations.
type IntegrationOauthProviderConfig struct {
	// SuccessRedirectURL is the URL to redirect to after successful OAuth integration.
	SuccessRedirectURL string `json:"successRedirectUrl" koanf:"successRedirectUrl" domain:"inherit" domainPrefix:"https://console" domainSuffix:"/organization-settings/integrations"`
	// ProviderSpecPath is the path to the declarative provider spec configuration file.
	ProviderSpecPath string `json:"providerSpecPath" koanf:"providerSpecPath" default:"internal/keystore/config/providers"`
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
	// ErrUnsupportedAuthType indicates the provider does not support the requested flow
	ErrUnsupportedAuthType = fmt.Errorf("provider does not support this authentication flow")
)

// OAuth Integration Constants
const (
	// URL patterns
	oauthCallbackPath = "/v1/integrations/oauth/callback"

	// Status messages
	statusConnected    = "connected"
	statusInvalid      = "invalid"
	statusExpired      = "expired"
	statusNotConnected = "not_connected"
)

// IntegrationHelper wraps the shared helper with handler-specific utilities.
type IntegrationHelper struct {
	inner *keystore.Helper
}

// NewIntegrationHelper creates a new integration helper.
func NewIntegrationHelper(provider, username string) *IntegrationHelper {
	return &IntegrationHelper{inner: keystore.NewHelper(provider, username)}
}

// Provider returns the normalized provider.
func (ih *IntegrationHelper) Provider() string { return ih.inner.Provider() }

// Name returns the integration name.
func (ih *IntegrationHelper) Name() string { return ih.inner.Name() }

// Description returns the integration description.
func (ih *IntegrationHelper) Description() string { return ih.inner.Description() }

// SecretName returns the hush secret name.
func (ih *IntegrationHelper) SecretName(fieldName string) string {
	return ih.inner.SecretName(fieldName)
}

// SecretDisplayName returns the display name for a secret.
func (ih *IntegrationHelper) SecretDisplayName(integrationName, fieldName string) string {
	return ih.inner.SecretDisplayName(integrationName, fieldName)
}

// SecretDescription returns the description for a secret.
func (ih *IntegrationHelper) SecretDescription(fieldName string) string {
	return ih.inner.SecretDescription(fieldName)
}

// CallbackURL returns the OAuth callback URL.
func (ih *IntegrationHelper) CallbackURL(baseURL string) string {
	return baseURL + oauthCallbackPath
}

// RedirectURL returns the success redirect URL with parameters.
func (ih *IntegrationHelper) RedirectURL(baseURL string) string {
	message := fmt.Sprintf("Successfully connected %s%s",
		keystore.TitleCase(ih.Provider()), keystore.IntegrationNameSuffix)
	return baseURL + "?provider=" + ih.Provider() + "&status=success&message=" + message
}

// StateData returns the OAuth state data.
func (ih *IntegrationHelper) StateData(orgID string, randomBytes []byte) string {
	return orgID + ":" + ih.Provider() + ":" + base64.URLEncoding.EncodeToString(randomBytes)
}

// StatusMessage returns status messages for integration status.
func (ih *IntegrationHelper) StatusMessage(status string) string {
	providerTitle := keystore.TitleCase(ih.Provider())
	switch status {
	case statusConnected:
		return providerTitle + " integration is connected and active"
	case statusInvalid:
		return providerTitle + " integration exists but has invalid tokens"
	case statusExpired:
		return providerTitle + " integration tokens have expired"
	case statusNotConnected:
		return "No " + ih.Provider() + " integration found"
	default:
		return providerTitle + " integration status unknown"
	}
}

// OAuth error helpers reused across handlers to preserve consistent messaging.
func wrapIntegrationError(operation string, err error) error {
	return fmt.Errorf("failed to %s integration: %w", operation, err)
}

func wrapTokenError(operation, provider string, err error) error {
	return fmt.Errorf("failed to %s token for %s: %w", operation, provider, err)
}
