package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// statePayloadParts is the number of parts in an encoded OAuth state payload
	statePayloadParts = 3
)

var (
	errIntegrationsRuntimeNotConfigured = errors.New("integrations runtime not configured")
	// errGitHubAppNotConfigured indicates required GitHub App operator credentials are absent from the provider spec
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
	// ErrUnsupportedAuthType indicates the provider does not support the requested flow
	ErrUnsupportedAuthType = errors.New("provider does not support this authentication flow")
	// ErrProviderHealthCheckFailed indicates the provider health check failed
	ErrProviderHealthCheckFailed = errors.New("provider health check failed")
)

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

// resolveActiveDefinition looks up a definition by ID and rejects inactive providers.
// Returns the resolved definition on success; returns and writes the HTTP error response on failure.
func (h *Handler) resolveActiveDefinition(ctx echo.Context, defID string, openapiCtx *OpenAPIContext) (types.Definition, error) {
	def, ok := h.IntegrationsRuntime.Registry().Definition(defID)
	if !ok {
		return types.Definition{}, h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !def.Active {
		return types.Definition{}, h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	return def, nil
}

// persistInstallationUserInput writes userInput into the installation's client config and persists it.
// If the JSON contains a "name" key, the installation display name is updated as well.
func (h *Handler) persistInstallationUserInput(ctx context.Context, installationRec *ent.Integration, userInput json.RawMessage) error {
	config := installationRec.Config
	config.ClientConfig = userInput

	update := h.DBClient.Integration.UpdateOneID(installationRec.ID).SetConfig(config)

	var inputMap map[string]any
	if err := json.Unmarshal(userInput, &inputMap); err == nil {
		if name, ok := inputMap["name"].(string); ok && name != "" {
			update.SetName(name)
			installationRec.Name = name
		}
	}

	if err := update.Exec(ctx); err != nil {
		return err
	}

	installationRec.Config = config

	return nil
}

// finalizeIntegrationConnection runs the post-credential sequence shared by all install paths:
// health check, credential save, installation metadata resolve and save, status update to Connected.
// callbackInput is passed to the installation metadata resolver and may be nil.
func (h *Handler) finalizeIntegrationConnection(
	ctx echo.Context,
	openapiCtx *OpenAPIContext,
	installationRec *ent.Integration,
	def types.Definition,
	credentialRegistration types.CredentialRegistration,
	credential types.CredentialSet,
	callbackInput json.RawMessage,
) error {
	requestCtx := ctx.Request().Context()
	logger := logx.FromContext(requestCtx).With().
		Str("definition_id", def.ID).
		Str("installation_id", installationRec.ID).
		Logger()

	healthOperation, err := h.IntegrationsRuntime.Registry().Operation(def.ID, types.HealthDefaultOperation)
	if err != nil {
		logger.Error().Err(err).Msg("health operation not registered")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	credentialOverrides := types.CredentialBindings{
		{Ref: credentialRegistration.Ref, Credential: credential},
	}

	if _, err := h.IntegrationsRuntime.ExecuteOperation(requestCtx, installationRec, healthOperation, credentialOverrides, nil); err != nil {
		logger.Error().Err(err).Msg("provider health check failed")
		return h.BadRequest(ctx, ErrProviderHealthCheckFailed, openapiCtx)
	}

	if err := h.IntegrationsRuntime.SaveCredential(requestCtx, installationRec, credentialRegistration.Ref, credential); err != nil {
		logger.Error().Err(err).Msg("failed to save credential")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if def.Installation != nil {
		metadata, ok, err := def.Installation.Resolve(requestCtx, types.InstallationRequest{
			Installation: installationRec,
			Credential:   credential,
			Config:       installationRec.Config,
			Input:        callbackInput,
		})
		if err != nil {
			logger.Error().Err(err).Msg("failed to resolve installation metadata")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		if ok {
			if err := integrationsruntime.SaveInstallationMetadata(requestCtx, installationRec, metadata); err != nil {
				logger.Error().Err(err).Msg("failed to save installation metadata")
				return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
			}
		}
	}

	if err := h.IntegrationsRuntime.MarkConnected(requestCtx, installationRec); err != nil {
		logger.Error().Err(err).Msg("failed to update integration status")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	return nil
}

func resolveCredentialRegistration(def types.Definition, credentialRef types.CredentialRef) (types.CredentialRegistration, error) {
	if !credentialRef.Valid() {
		return types.CredentialRegistration{}, rout.MissingField("credentialRef")
	}

	registration, ok := lo.Find(def.CredentialRegistrations, func(registration types.CredentialRegistration) bool {
		return registration.Ref.String() == credentialRef.String()
	})
	if ok {
		return registration, nil
	}

	return types.CredentialRegistration{}, ErrInvalidInput
}

func validateDefinitionUserInput(def types.Definition, input IntegrationConfigBody) error {
	if len(input) == 0 || def.UserInput == nil || len(def.UserInput.Schema) == 0 {
		return nil
	}

	userInputValidation, err := jsonx.ValidateSchema(def.UserInput.Schema, input.RawMessage())
	if err != nil {
		return err
	}
	if !userInputValidation.Valid() {
		return ErrInvalidInput
	}

	return nil
}

