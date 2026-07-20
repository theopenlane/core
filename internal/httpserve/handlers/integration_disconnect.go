package handlers

import (
	"fmt"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/logx"
)

// DisconnectIntegration executes the definition-driven teardown flow for one installed integration
func (h *Handler) DisconnectIntegration(ctx echo.Context) error {
	in, err := BindAndValidate[models.DisconnectIntegrationRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	userCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(userCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, ErrUnauthorized)
	}

	if in.IntegrationID == "" {
		logx.FromContext(userCtx).Error().Err(ErrIntegrationIDRequired).Msg("missing integrationID in request")

		return h.BadRequest(ctx, ErrIntegrationIDRequired)
	}

	record, err := h.IntegrationsRuntime.ResolveIntegration(userCtx, integrationsruntime.IntegrationLookup{
		IntegrationID: in.IntegrationID,
		OwnerID:       caller.OrganizationID,
	})
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Interface("request", in).Msg("failed to resolve integration record")

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(record.DefinitionID)
	if !ok {
		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	result, err := h.IntegrationsRuntime.Disconnect(userCtx, record)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Interface("request", in).Msg("disconnect failed")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	resp := models.DeleteIntegrationResponse{
		Reply:   rout.Reply{Success: true},
		Message: lo.CoalesceOrEmpty(result.Message, fmt.Sprintf("%s integration disconnected", def.DisplayName)),
	}

	resp.RedirectURL = result.RedirectURL
	resp.Details = result.Details

	resp.DeletedID = record.ID

	return h.Success(ctx, resp)
}
