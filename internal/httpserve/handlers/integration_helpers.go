package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	entsql "entgo.io/ent/dialect/sql"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	entintegration "github.com/theopenlane/core/internal/ent/generated/integration"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/jsonx"
)

// statePayloadParts is the number of parts in an encoded OAuth state payload.
const statePayloadParts = 3

var (
	errIntegrationWorkflowEngineNotConfigured = errors.New("integration workflow engine not configured")
	errIntegrationsRuntimeNotConfigured       = errors.New("integrations runtime not configured")
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
	case errors.Is(err, integrationsruntime.ErrDefinitionNotFound):
		return http.StatusNotFound
	case errors.Is(err, integrationsruntime.ErrOperationNotFound):
		return http.StatusNotFound
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
	case errors.Is(err, engine.ErrIntegrationDefinitionNotFound):
		return http.StatusNotFound
	case errors.Is(err, engine.ErrIntegrationOperationNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) requireIntegrationsRuntime(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if h.IntegrationsRuntime != nil {
		return nil
	}

	return h.InternalServerError(ctx, errIntegrationsRuntimeNotConfigured, openapiCtx)
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

func validateDefinitionUserInput(def types.Definition, input IntegrationConfigBody) error {
	if len(input) == 0 || def.UserInput == nil || len(def.UserInput.Schema) == 0 {
		return nil
	}

	userInputValidation, err := jsonx.ValidateSchema(def.UserInput.Schema, input.ToMap())
	if err != nil {
		return err
	}
	if !userInputValidation.Valid() {
		return ErrInvalidInput
	}

	return nil
}

func (h *Handler) resolveOrCreateDefinitionIntegration(ctx context.Context, ownerID, installationID string, def types.Definition) (*ent.Integration, bool, error) {
	if installationID != "" {
		record, err := h.IntegrationsRuntime.ResolveInstallation(ctx, ownerID, installationID, def.ID)
		if err != nil {
			return nil, false, err
		}

		if err := h.refreshDefinitionIntegration(ctx, record, def); err != nil {
			return nil, false, err
		}

		return record, false, nil
	}

	record, err := h.findLatestDefinitionIntegration(ctx, ownerID, def.ID)
	if err != nil {
		return nil, false, err
	}

	if record != nil {
		if err := h.refreshDefinitionIntegration(ctx, record, def); err != nil {
			return nil, false, err
		}

		return record, false, nil
	}

	record, err = h.DBClient.Integration.Create().
		SetOwnerID(ownerID).
		SetName(def.DisplayName).
		SetDefinitionID(def.ID).
		SetDefinitionSlug(def.Slug).
		SetFamily(def.Family).
		SetStatus(enums.IntegrationStatusPending).
		Save(ctx)
	if err != nil {
		return nil, false, err
	}

	return record, true, nil
}

func (h *Handler) findLatestDefinitionIntegration(ctx context.Context, ownerID, definitionID string) (*ent.Integration, error) {
	record, err := h.DBClient.Integration.Query().
		Where(
			entintegration.OwnerIDEQ(ownerID),
			entintegration.DefinitionIDEQ(definitionID),
		).
		Order(
			entintegration.ByUpdatedAt(entsql.OrderDesc()),
			entintegration.ByCreatedAt(entsql.OrderDesc()),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return record, nil
}

func (h *Handler) refreshDefinitionIntegration(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	if installation == nil {
		return nil
	}

	if installation.Name == def.DisplayName &&
		installation.DefinitionID == def.ID &&
		installation.DefinitionSlug == def.Slug &&
		installation.Family == def.Family {
		return nil
	}

	if err := h.DBClient.Integration.UpdateOneID(installation.ID).
		SetName(def.DisplayName).
		SetDefinitionID(def.ID).
		SetDefinitionSlug(def.Slug).
		SetFamily(def.Family).
		Exec(ctx); err != nil {
		return err
	}

	installation.Name = def.DisplayName
	installation.DefinitionID = def.ID
	installation.DefinitionSlug = def.Slug
	installation.Family = def.Family

	return nil
}
