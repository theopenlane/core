package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
)

var (
	errIntegrationBrokerNotConfigured     = errors.New("integration broker not configured")
	errIntegrationStoreNotConfigured      = errors.New("integration store not configured")
	errIntegrationRegistryNotConfigured   = errors.New("integration registry not configured")
	errIntegrationOperationsNotConfigured = errors.New("integration operations manager not configured")
	errKeymakerNotConfigured              = errors.New("integration keymaker service not configured")
)

// IntegrationOauthProviderConfig represents the configuration for OAuth providers used for integrations.
type IntegrationOauthProviderConfig struct {
	// Enabled toggles initialization of the integration provider registry.
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// SuccessRedirectURL is the URL to redirect to after successful OAuth integration.
	SuccessRedirectURL string `json:"successredirecturl" koanf:"successredirecturl" domain:"inherit" domainPrefix:"https://console" domainSuffix:"/organization-settings/integrations"`
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
	// ErrProviderHealthCheckFailed indicates the provider health check failed
	ErrProviderHealthCheckFailed = errors.New("provider health check failed")
)

// buildStatePayload encodes the OAuth state payload for cookies and callbacks.
func buildStatePayload(orgID, provider string, randomBytes []byte) string {
	return orgID + ":" + provider + ":" + base64.URLEncoding.EncodeToString(randomBytes)
}

// OAuth error helpers reused across handlers to preserve consistent messaging.
func wrapIntegrationError(operation string, err error) error {
	return fmt.Errorf("failed to %s integration: %w", operation, err)
}

func wrapTokenError(operation, provider string, err error) error {
	return fmt.Errorf("failed to %s token for %s: %w", operation, provider, err)
}
