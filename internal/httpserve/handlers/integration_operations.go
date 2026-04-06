package handlers

import (
	"context"
	"time"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/operations"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
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

// RunIntegrationOperation queues operation execution; operations have individual policies dictating when or how they run
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

	if req.IntegrationID == "" || req.Body.Operation == "" {
		// not terribly concerned about distinct error responses here since this isn't intended to be used as a primary execution method
		logx.FromContext(requestCtx).Error().Err(ErrIntegrationIDRequired).Msg("missing integrationID or Operation in request")

		return h.BadRequest(ctx, ErrIntegrationIDRequired, openapiCtx)
	}

	integrationRef, err := h.IntegrationsRuntime.ResolveIntegration(requestCtx, integrationsruntime.IntegrationLookup{
		IntegrationID: req.IntegrationID,
		OwnerID:       caller.OrganizationID,
	})
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("failed to resolve installation")

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(integrationRef.DefinitionID)
	if !ok {
		logx.FromContext(requestCtx).Error().Str("definitionID", integrationRef.DefinitionID).Msg("definition not found in registry")

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	if !def.Active {
		logx.FromContext(requestCtx).Error().Err(ErrProviderDisabled).Str("definitionID", integrationRef.DefinitionID).Msg("integration provider is disabled, not executing operation")

		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	operationName := req.Body.Operation
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

	if inlineExecution {
		output, err := h.IntegrationsRuntime.ExecuteOperation(queueCtx, integrationRef, operation, nil, configDoc)
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("operation execution failed")

			return h.BadRequest(ctx, err, openapiCtx)
		}

		meta := integrationRef.Metadata
		if meta == nil {
			meta = map[string]any{}
		}

		meta["lastSuccessfulHealthCheck"] = time.Now().UTC().Format(time.RFC3339)

		if err := h.IntegrationsRuntime.DB().Integration.UpdateOneID(integrationRef.ID).
			SetMetadata(meta).Exec(queueCtx); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("integrationID", integrationRef.ID).Msg("failed to persist health check timestamp")
			// not failing the request at this point since the operation itself succeeded and this is just a best-effort update to the integration record
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
		IntegrationID: integrationRef.ID,
		Operation:     operationName,
		Config:        configDoc,
		RunType:       enums.IntegrationRunTypeManual,
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
