package handlers

import (
	"fmt"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/logx"
)

// DisconnectIntegration executes the definition-driven teardown flow for one installed integration
func (h *Handler) DisconnectIntegration(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleDisconnectIntegrationRequest, models.DeleteIntegrationResponse{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
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
		logx.FromContext(userCtx).Error().Err(err).Interface("request", in).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapi)
	}

	result, err := h.IntegrationsRuntime.Disconnect(userCtx, record)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Interface("request", in).Msg("disconnect failed")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	resp := models.DeleteIntegrationResponse{
		Reply:   rout.Reply{Success: true},
		Message: lo.CoalesceOrEmpty(result.Message, fmt.Sprintf("%s integration disconnected", lo.CoalesceOrEmpty(def.DisplayName, def.Slug))),
	}

	resp.RedirectURL = result.RedirectURL
	resp.Details = result.Details

	if !result.SkipLocalCleanup {
		resp.DeletedID = record.ID
	}

	return h.Success(ctx, resp)
}
