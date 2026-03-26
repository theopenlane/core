package handlers

import (
	"context"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// integrationOperationQueueDetails captures queue response details for integration operation requests
type integrationOperationQueueDetails struct {
	// RunID is the queued integration run identifier
	RunID string `json:"run_id"`
	// EventID is the emitted event identifier
	EventID string `json:"event_id"`
	// Status is the queued integration run status
	Status string `json:"status"`
}

// RunIntegrationOperation queues provider operations for async execution.
// Health checks are executed inline to return immediate validation status to callers
func (h *Handler) RunIntegrationOperation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleIntegrationOperationPayload, IntegrationOperationResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(req.DefinitionID)
	if !ok || !def.Active {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	operationName := req.Body.Operation
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}

	operation, err := h.IntegrationsRuntime.Registry().Operation(def.ID, operationName)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("operation not found")

		return h.BadRequest(ctx, operations.ErrDispatchInputInvalid, openapiCtx)
	}

	inlineExecution := operation.Policy.Inline

	queueCtx := context.WithoutCancel(requestCtx)
	configDoc := jsonx.CloneRawMessage(req.Body.Config)

	if inlineExecution {
		if err := operations.ValidateConfig(operation.ConfigSchema, configDoc); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Msg("invalid operation config")

			return h.BadRequest(ctx, operations.ErrDispatchInputInvalid, openapiCtx)
		}
	}

	integrationID := req.IntegrationID

	installationRec, err := h.IntegrationsRuntime.ResolveIntegration(requestCtx, caller.OrganizationID, integrationID, def.ID)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	if inlineExecution {
		output, err := h.IntegrationsRuntime.ExecuteOperation(queueCtx, installationRec, operation, nil, configDoc)
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("operation execution failed")
			return h.BadRequest(ctx, err, openapiCtx)
		}

		return h.Success(ctx, IntegrationOperationResponse{
			Reply:     rout.Reply{Success: true},
			Provider:  def.ID,
			Operation: operationName,
			Status:    "ok",
			Summary:   "Integration operation completed",
			Details:   output,
		})
	}

	result, err := h.IntegrationsRuntime.Dispatch(queueCtx, operations.DispatchRequest{
		InstallationID: installationRec.ID,
		Operation:      operationName,
		Config:         configDoc,
		RunType:        enums.IntegrationRunTypeManual,
	})
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("failed to queue operation")
		return h.BadRequest(ctx, err, openapiCtx)
	}

	queueDetails, err := jsonx.ToRawMessage(integrationOperationQueueDetails{
		RunID:   result.RunID,
		EventID: result.EventID,
		Status:  result.Status.String(),
	})
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	return h.Success(ctx, IntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  def.ID,
		Operation: operationName,
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details:   queueDetails,
	})
}
