package handlers

import (
	"context"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/operations"
	integrationstypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows/engine"
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
// Health checks are executed inline to return immediate validation status to callers.
func (h *Handler) RunIntegrationOperation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleIntegrationOperationPayload, IntegrationOperationResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	requestCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, err := h.resolveActiveDefinition(ctx, req.DefinitionID, openapiCtx)
	if err != nil {
		return err
	}

	operationName := req.Body.Operation
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}
	integrationID := req.IntegrationID

	logger := logx.FromContext(requestCtx).With().
		Str("definition_id", def.ID).
		Str("operation", operationName).
		Logger()

	operation, err := h.IntegrationsRuntime.Registry().Operation(def.ID, operationName)
	if err != nil {
		logger.Error().Err(err).Msg("operation not found")
		return h.BadRequest(ctx, operations.ErrDispatchInputInvalid, openapiCtx)
	}

	inlineExecution := operationName == integrationstypes.HealthDefaultOperation || operation.Policy.Inline
	queueCtx := context.WithoutCancel(requestCtx)
	configDoc := jsonx.CloneRawMessage(req.Body.Config)

	if inlineExecution {
		if err := operations.ValidateConfig(operation.ConfigSchema, configDoc); err != nil {
			logger.Error().Err(err).Msg("invalid operation config")
			return h.BadRequest(ctx, operations.ErrDispatchInputInvalid, openapiCtx)
		}
	}

	var installationRec *ent.Integration

	if integrationID != "" {
		installationRec, err = h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, integrationID, def.ID)
		if err != nil {
			logger.Error().Err(err).Str("integration_id", integrationID).Msg("failed to resolve installation")
			return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
		}

		integrationID = installationRec.ID
	}

	if inlineExecution {
		if installationRec == nil {
			installationRec, err = h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, "", def.ID)
			if err != nil {
				logger.Error().Err(err).Msg("failed to resolve installation")
				return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
			}
		}

		installLogger := logger.With().Str("installation_id", installationRec.ID).Logger()

		output, err := h.IntegrationsRuntime.ExecuteOperation(queueCtx, installationRec, operation, nil, configDoc)
		if err != nil {
			installLogger.Error().Err(err).Msg("operation execution failed")
			return h.BadRequest(ctx, err, openapiCtx)
		}

		summary := "Integration operation completed"
		if operationName == integrationstypes.HealthDefaultOperation {
			summary = "Health check completed"
		}

		return h.Success(ctx, IntegrationOperationResponse{
			Reply:     rout.Reply{Success: true},
			Provider:  def.Slug,
			Operation: operationName,
			Status:    "ok",
			Summary:   summary,
			Details:   output,
		})
	}

	var result engine.IntegrationQueueResult

	if h.WorkflowEngine != nil {
		result, err = h.WorkflowEngine.QueueIntegrationOperation(queueCtx, engine.IntegrationQueueRequest{
			OrgID:          caller.OrganizationID,
			DefinitionID:   def.ID,
			InstallationID: integrationID,
			Operation:      operationName,
			Config:         configDoc,
			RunType:        enums.IntegrationRunTypeManual,
		})
	} else {
		var dispatchRec *ent.Integration

		dispatchRec, err = h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, integrationID, def.ID)
		if err != nil {
			logger.Error().Err(err).Str("integration_id", integrationID).Msg("failed to resolve installation for dispatch")
			return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
		}

		dispatchResult, dispatchErr := h.IntegrationsRuntime.Dispatch(queueCtx, operations.DispatchRequest{
			InstallationID: dispatchRec.ID,
			Operation:      operationName,
			Config:         configDoc,
			RunType:        enums.IntegrationRunTypeManual,
		})
		if dispatchErr != nil {
			err = dispatchErr
		} else {
			result = engine.IntegrationQueueResult{
				RunID:   dispatchResult.RunID,
				EventID: dispatchResult.EventID,
				Status:  dispatchResult.Status,
			}
		}
	}

	if err != nil {
		logger.Error().Err(err).Str("integration_id", integrationID).Msg("failed to queue operation")
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
		Provider:  def.Slug,
		Operation: operationName,
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details:   queueDetails,
	})
}
