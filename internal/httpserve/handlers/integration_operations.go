package handlers

import (
	"errors"
	"maps"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/logx"
)

// RunIntegrationOperation executes a provider-published operation using stored credentials
func (h *Handler) RunIntegrationOperation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleIntegrationOperationPayload, openapi.IntegrationOperationResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if h.IntegrationRegistry == nil {
		return h.InternalServerError(ctx, errIntegrationRegistryNotConfigured, openapiCtx)
	}
	if h.IntegrationOperations == nil {
		return h.InternalServerError(ctx, errIntegrationOperationsNotConfigured, openapiCtx)
	}

	requestCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(requestCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapiCtx)
	}

	providerType := types.ProviderTypeFromString(req.Provider)
	if providerType == types.ProviderUnknown {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	operationName := types.OperationName(strings.TrimSpace(req.Body.Operation))
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}

	operationConfig := cloneOperationConfig(req.Body.Config)
	if h.IntegrationStore != nil {
		if integrationRecord, err := h.IntegrationStore.EnsureIntegration(requestCtx, user.OrganizationID, providerType); err != nil {
			logx.FromContext(requestCtx).Warn().Err(err).Str("provider", string(providerType)).Msg("failed to load integration config")
		} else {
			merged, err := helpers.ResolveOperationConfig(&integrationRecord.Config, string(operationName), req.Body.Config)
			if err != nil {
				return h.BadRequest(ctx, err, openapiCtx)
			}
			operationConfig = merged
		}
	}

	result, runErr := h.IntegrationOperations.Run(requestCtx, types.OperationRequest{
		OrgID:    user.OrganizationID,
		Provider: providerType,
		Name:     operationName,
		Config:   operationConfig,
		Force:    req.Body.Force,
	})
	if runErr != nil {
		switch {
		case errors.Is(runErr, keystore.ErrCredentialNotFound):
			return h.NotFound(ctx, wrapIntegrationError("find", runErr), openapiCtx)
		case errors.Is(runErr, keystore.ErrOperationNotRegistered),
			errors.Is(runErr, keystore.ErrProviderNotRegistered),
			errors.Is(runErr, keystore.ErrProviderRequired),
			errors.Is(runErr, keystore.ErrOperationNameRequired):
			return h.BadRequest(ctx, runErr, openapiCtx)
		default:
			return h.InternalServerError(ctx, wrapIntegrationError("run operation", runErr), openapiCtx)
		}
	}

	out := openapi.IntegrationOperationResponse{
		Reply:     rout.Reply{Success: result.Status == types.OperationStatusOK},
		Provider:  string(providerType),
		Operation: string(operationName),
		Status:    string(result.Status),
		Summary:   result.Summary,
		Details:   result.Details,
	}

	return h.Success(ctx, out)
}

func cloneOperationConfig(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	return maps.Clone(input)
}
