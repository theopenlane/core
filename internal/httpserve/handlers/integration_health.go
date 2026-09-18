package handlers

import (
	"context"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	integrationsruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// CheckIntegrationHealth runs the installation health assessment, recording connection and
// per-operation health, and returns the outcome
func (h *Handler) CheckIntegrationHealth(ctx echo.Context) error {
	req, err := BindAndValidate[IntegrationHealthRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser)
	}

	if req.IntegrationID == "" {
		return h.BadRequest(ctx, ErrIntegrationIDRequired)
	}

	installation, err := h.IntegrationsRuntime.ResolveIntegration(requestCtx, integrationsruntime.IntegrationLookup{
		IntegrationID: req.IntegrationID,
		OwnerID:       caller.OrganizationID,
	})
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("failed to resolve installation")

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(installation.DefinitionID)
	if !ok {
		logx.FromContext(requestCtx).Error().Str("definitionID", installation.DefinitionID).Msg("definition not found in registry")

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	if !def.Active {
		logx.FromContext(requestCtx).Error().Err(ErrProviderDisabled).Str("definitionID", installation.DefinitionID).Msg("integration provider is disabled, not running assessment")

		return h.BadRequest(ctx, ErrProviderDisabled)
	}

	// the assessment records health state, so it runs to completion even if the caller disconnects
	assessment, err := h.IntegrationsRuntime.RunHealthAssessment(context.WithoutCancel(requestCtx), installation)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Msg("health assessment failed")

		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, IntegrationHealthResponse{
		Reply:         rout.Reply{Success: true},
		Provider:      installation.DefinitionID,
		IntegrationID: installation.ID,
		Status:        assessment.Status.String(),
		Connection: IntegrationConnectionHealth{
			Healthy: assessment.Connection.Healthy,
			Reason:  assessment.Connection.Reason,
		},
		Operations: lo.Map(assessment.Operations, func(result integrationsruntime.OperationHealthResult, _ int) IntegrationOperationHealth {
			return IntegrationOperationHealth{Name: result.Name, Healthy: result.Healthy, Reason: result.Reason}
		}),
	})
}
