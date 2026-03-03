package handlers

import (
	"context"
	"encoding/json"
	"errors"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/jsonx"
)

const defaultHealthOperation types.OperationName = "health.default"

// integrationOperationQueueDetails captures queue response details for integration operation requests
type integrationOperationQueueDetails struct {
	// RunID is the queued integration run identifier
	RunID string `json:"run_id"`
	// EventID is the emitted event identifier
	EventID string `json:"event_id"`
	// Status is the queued integration run status
	Status string `json:"status"`
}

// integrationOperationQueueDetailsDoc converts queue details into a JSON object document.
func integrationOperationQueueDetailsDoc(details integrationOperationQueueDetails) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := jsonx.RoundTrip(details, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}

	return raw, nil
}

// copyRawJSON clones a raw JSON document to avoid accidental aliasing.
func copyRawJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}

	return append(json.RawMessage(nil), raw...)
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

	operationName := types.OperationName(req.Body.Operation)
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}

	if h.WorkflowEngine == nil {
		if operationName != defaultHealthOperation {
			return h.InternalServerError(ctx, errIntegrationWorkflowEngineNotConfigured, openapiCtx)
		}
	}

	queueCtx := context.WithoutCancel(requestCtx)

	configDoc := copyRawJSON(req.Body.Config)

	if operationName == defaultHealthOperation {
		if h.IntegrationOperations == nil {
			return h.InternalServerError(ctx, errIntegrationOperationsNotConfigured, openapiCtx)
		}

		result, err := h.IntegrationOperations.Run(queueCtx, types.OperationRequest{
			OrgID:    caller.OrganizationID,
			Provider: providerType,
			Name:     operationName,
			Config:   configDoc,
			Force:    req.Body.Force,
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
			Details:   copyRawJSON(result.Details),
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
		OrgID:     caller.OrganizationID,
		Provider:  providerType,
		Operation: operationName,
		Config:    configDoc,
		Force:     req.Body.Force,
		RunType:   enums.IntegrationRunTypeManual,
	})
	if err != nil {
		if errors.Is(err, keystore.ErrOperationNotRegistered) {
			return h.BadRequest(ctx, err, openapiCtx)
		}
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	queueDetails, err := integrationOperationQueueDetailsDoc(integrationOperationQueueDetails{
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
