package handlers

import (
	"context"
	"errors"
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/jsonx"
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
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleIntegrationOperationPayload, openapi.IntegrationOperationResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	requestCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	providerType := types.ProviderTypeFromString(req.Provider)
	if providerType == types.ProviderUnknown {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}
	if err := h.validateIntegrationProvider(providerType); err != nil {
		if errors.Is(err, errIntegrationRuntimeNotConfigured) {
			return h.InternalServerError(ctx, err, openapiCtx)
		}

		return h.BadRequest(ctx, err, openapiCtx)
	}

	operationName := types.OperationName(req.Body.Operation)
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}
	integrationID := req.IntegrationID

	if h.WorkflowEngine == nil && operationName != types.OperationHealthDefault {
		return h.InternalServerError(ctx, errIntegrationWorkflowEngineNotConfigured, openapiCtx)
	}

	queueCtx := context.WithoutCancel(requestCtx)
	configDoc := jsonx.CloneRawMessage(req.Body.Config)

	if operationName == types.OperationHealthDefault {
		result, err := h.IntegrationRuntime.Operations().Run(queueCtx, types.OperationRequest{
			OrgID:         caller.OrganizationID,
			IntegrationID: integrationID,
			Provider:      providerType,
			Name:          operationName,
			Config:        configDoc,
			Force:         req.Body.Force,
		})
		if err != nil {
			return h.BadRequest(ctx, err, openapiCtx)
		}

		out := openapi.IntegrationOperationResponse{
			Reply:     rout.Reply{Success: result.Status == types.OperationStatusOK},
			Provider:  string(providerType),
			Operation: string(operationName),
			Status:    string(result.Status),
			Summary:   result.Summary,
			Details:   jsonx.CloneRawMessage(result.Details),
		}

		if out.Status == "" {
			out.Status = string(types.OperationStatusUnknown)
		}
		if out.Summary == "" {
			out.Summary = "Health check completed"
		}

		if result.Status != types.OperationStatusOK {
			return h.BadRequest(ctx, ErrProviderHealthCheckFailed, openapiCtx)
		}

		return h.Success(ctx, out)
	}

	result, err := h.WorkflowEngine.QueueIntegrationOperation(queueCtx, engine.IntegrationQueueRequest{
		OrgID:         caller.OrganizationID,
		Provider:      providerType,
		IntegrationID: integrationID,
		Operation:     operationName,
		Config:        configDoc,
		Force:         req.Body.Force,
		RunType:       enums.IntegrationRunTypeManual,
	})
	if err != nil {
		switch integrationHTTPStatus(err) {
		case http.StatusBadRequest:
			return h.BadRequest(ctx, err, openapiCtx)
		default:
			return h.InternalServerError(ctx, err, openapiCtx)
		}
	}

	queueDetails, err := jsonx.ToRawMessage(integrationOperationQueueDetails{
		RunID:   result.RunID,
		EventID: result.EventID,
		Status:  result.Status.String(),
	})
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	out := openapi.IntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  string(providerType),
		Operation: string(operationName),
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details:   queueDetails,
	}

	return h.Success(ctx, out)
}
