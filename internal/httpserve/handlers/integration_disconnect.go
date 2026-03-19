package handlers

import (
	"fmt"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/logx"
)

// DisconnectIntegration removes the stored integration configuration and secrets for a provider.
func (h *Handler) DisconnectIntegration(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleDisconnectIntegrationRequest, models.DeleteIntegrationResponse{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}
	if err := h.requireIntegrationsRuntime(ctx, openapi); err != nil {
		return err
	}

	userCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(userCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, ErrUnauthorized, openapi)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(in.DefinitionID)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapi)
	}

	record, err := h.IntegrationsRuntime.ResolveInstallation(userCtx, caller.OrganizationID, in.IntegrationID, def.ID)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Str("integration_id", in.IntegrationID).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapi)
	}

	integrationID := record.ID
	logger := logx.FromContext(userCtx).With().Str("integration_id", integrationID).Logger()

	if err := h.IntegrationsRuntime.DeleteCredential(userCtx, integrationID); err != nil {
		logger.Error().Err(err).Msg("failed to delete credential")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if err := h.IntegrationsRuntime.DeleteInstallation(userCtx, integrationID); err != nil {
		if ent.IsNotFound(err) {
			return h.BadRequest(ctx, ErrIntegrationNotFound, openapi)
		}

		logger.Error().Err(err).Msg("failed to delete integration")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	displayName := def.DisplayName
	if displayName == "" {
		displayName = def.Slug
	}

	return h.Success(ctx, models.DeleteIntegrationResponse{
		Reply:     rout.Reply{Success: true},
		Message:   fmt.Sprintf("%s integration disconnected", displayName),
		DeletedID: integrationID,
	})
}
