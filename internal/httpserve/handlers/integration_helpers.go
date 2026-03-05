package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
)

// statePayloadParts is the number of parts in an encoded OAuth state payload.
const statePayloadParts = 3

var (
	errIntegrationBrokerNotConfigured         = errors.New("integration broker not configured")
	errIntegrationStoreNotConfigured          = errors.New("integration store not configured")
	errIntegrationRegistryNotConfigured       = errors.New("integration registry not configured")
	errIntegrationOperationsNotConfigured     = errors.New("integration operations manager not configured")
	errIntegrationWorkflowEngineNotConfigured = errors.New("integration workflow engine not configured")
	errKeymakerNotConfigured                  = errors.New("integration keymaker not configured")
	// errDBClientNotConfigured indicates the database client is missing.
	errDBClientNotConfigured = errors.New("database client not configured")
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
	ErrInvalidState = errors.New("invalid OAuth state parameter")
	// ErrMissingCode is returned when OAuth authorization code is missing
	ErrMissingCode = errors.New("missing OAuth authorization code")
	// ErrExchangeAuthCode is returned when OAuth code exchange fails
	ErrExchangeAuthCode = errors.New("failed to exchange authorization code")
	// ErrValidateToken is returned when OAuth token validation fails
	ErrValidateToken = errors.New("failed to validate OAuth token")
	// ErrInvalidStateFormat is returned when OAuth state format is invalid
	ErrInvalidStateFormat = errors.New("invalid state format")
	// ErrProviderRequired is returned when provider parameter is missing
	ErrProviderRequired = errors.New("provider parameter is required")
	// ErrIntegrationIDRequired is returned when integration ID is missing
	ErrIntegrationIDRequired = errors.New("integration ID is required")
	// ErrIntegrationNotFound is returned when integration is not found
	ErrIntegrationNotFound = errors.New("integration not found")
	// ErrDeleteSecrets is returned when deleting integration secrets fails
	ErrDeleteSecrets = errors.New("failed to delete integration secrets")
	// ErrUnsupportedAuthType indicates the provider does not support the requested flow
	ErrUnsupportedAuthType = errors.New("provider does not support this authentication flow")
	// ErrProviderHealthCheckFailed indicates the provider health check failed
	ErrProviderHealthCheckFailed = errors.New("provider health check failed")
)

// integrationHTTPStatus maps known integration errors to HTTP status codes.
// Returns http.StatusInternalServerError for unrecognized errors.
func integrationHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrIntegrationNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalidState),
		errors.Is(err, ErrInvalidStateFormat),
		errors.Is(err, ErrMissingCode),
		errors.Is(err, ErrUnsupportedAuthType),
		errors.Is(err, ErrExchangeAuthCode),
		errors.Is(err, ErrValidateToken),
		errors.Is(err, keystore.ErrOperationNotRegistered),
		errors.Is(err, keystore.ErrCredentialNotFound):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// buildStatePayload encodes the OAuth state payload for cookies and callbacks.
func buildStatePayload(orgID, provider string, randomBytes []byte) string {
	return orgID + ":" + provider + ":" + base64.URLEncoding.EncodeToString(randomBytes)
}

// parseStatePayload decodes the OAuth state payload and extracts the org and provider values.
func parseStatePayload(state string) (string, string, error) {
	if state == "" {
		return "", "", ErrInvalidStateFormat
	}

	decoded, err := decodeURLBase64(state)
	if err != nil {
		return "", "", ErrInvalidStateFormat
	}

	parts := strings.SplitN(string(decoded), ":", statePayloadParts)
	if len(parts) != statePayloadParts {
		return "", "", ErrInvalidStateFormat
	}

	orgID := parts[0]
	provider := parts[1]
	randomPart := parts[2]
	if orgID == "" || provider == "" || randomPart == "" {
		return "", "", ErrInvalidStateFormat
	}

	if _, err := decodeURLBase64(randomPart); err != nil {
		return "", "", ErrInvalidStateFormat
	}

	return orgID, provider, nil
}

func decodeURLBase64(value string) ([]byte, error) {
	if decoded, err := base64.URLEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}

	return base64.RawURLEncoding.DecodeString(value)
}

// stateFingerprint returns a non-reversible short fingerprint for state logging.
func stateFingerprint(state string) string {
	if state == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(state))

	return base64.RawURLEncoding.EncodeToString(sum[:8])
}

// OAuth error helpers reused across handlers to preserve consistent messaging.
func wrapIntegrationError(operation string, err error) error {
	return fmt.Errorf("failed to %s integration: %w", operation, err)
}

func wrapTokenError(operation, provider string, err error) error {
	return fmt.Errorf("failed to %s token for %s: %w", operation, provider, err)
}

func (h *Handler) updateIntegrationProviderMetadata(ctx context.Context, integrationID string, provider types.ProviderType) error {
	if h == nil || h.DBClient == nil || h.IntegrationRegistry == nil {
		return nil
	}

	spec, ok := h.IntegrationRegistry.Config(provider)
	if !ok || spec.Active == nil || !*spec.Active {
		return nil
	}

	meta, ok := h.IntegrationRegistry.ProviderMetadataCatalog()[provider]
	if !ok {
		meta = spec.ToProviderConfig()
	}

	entry := buildIntegrationProviderMetadata(provider, spec, meta, h.IntegrationRegistry)
	return h.DBClient.Integration.UpdateOneID(integrationID).
		SetProviderMetadata(entry).
		Exec(ctx)
}
