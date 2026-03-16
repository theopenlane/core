package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// statePayloadParts is the number of parts in an encoded OAuth state payload.
const statePayloadParts = 3

var (
	errIntegrationWorkflowEngineNotConfigured = errors.New("integration workflow engine not configured")
	// errDBClientNotConfigured indicates the database client is missing.
	errDBClientNotConfigured = errors.New("database client not configured")
	// errGitHubAppNotConfigured indicates required GitHub App operator credentials are absent from the provider spec.
	errGitHubAppNotConfigured = errors.New("github app integration not configured: required credentials missing from provider spec")
)

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
	case errors.Is(err, ErrInvalidState):
		return http.StatusBadRequest
	case errors.Is(err, ErrInvalidStateFormat):
		return http.StatusBadRequest
	case errors.Is(err, ErrMissingCode):
		return http.StatusBadRequest
	case errors.Is(err, ErrIntegrationIDRequired):
		return http.StatusBadRequest
	case errors.Is(err, ErrInvalidProvider):
		return http.StatusBadRequest
	case errors.Is(err, ErrProviderDisabled):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnsupportedAuthType):
		return http.StatusBadRequest
	case errors.Is(err, ErrExchangeAuthCode):
		return http.StatusBadRequest
	case errors.Is(err, ErrValidateToken):
		return http.StatusBadRequest
	case errors.Is(err, keystore.ErrCredentialNotFound):
		return http.StatusBadRequest
	case errors.Is(err, integrationsruntime.ErrInstallationNotFound):
		return http.StatusNotFound
	case errors.Is(err, integrationsruntime.ErrInstallationIDRequired):
		return http.StatusBadRequest
	case errors.Is(err, integrationsruntime.ErrInstallationDefinitionMismatch):
		return http.StatusBadRequest
	case errors.Is(err, engine.ErrInstallationNotFound):
		return http.StatusNotFound
	case errors.Is(err, engine.ErrInstallationIDRequired):
		return http.StatusBadRequest
	case errors.Is(err, engine.ErrIntegrationProviderRequired):
		return http.StatusBadRequest
	case errors.Is(err, engine.ErrIntegrationOperationCriteriaRequired):
		return http.StatusBadRequest
	case errors.Is(err, engine.ErrIntegrationScopeConditionFalse):
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
