package handlers

import (
	"context"
	"errors"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// RunIntegrationOperation queues a provider-published operation for async execution
func (h *Handler) RunIntegrationOperation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleIntegrationOperationPayload, openapi.IntegrationOperationResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
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

	if h.WorkflowEngine == nil {
		return h.InternalServerError(ctx, errIntegrationWorkflowEngineNotConfigured, openapiCtx)
	}

	queueCtx := context.WithoutCancel(requestCtx)
	result, err := h.WorkflowEngine.QueueIntegrationOperation(queueCtx, engine.IntegrationQueueRequest{
		OrgID:    user.OrganizationID,
		Provider: providerType,
		Operation: operationName,
		Config:   req.Body.Config,
		Force:    req.Body.Force,
		RunType:  enums.IntegrationRunTypeManual,
	})
	if err != nil {
		if errors.Is(err, keystore.ErrOperationNotRegistered) {
			return h.BadRequest(ctx, err, openapiCtx)
		}
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	out := openapi.IntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  string(providerType),
		Operation: string(operationName),
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details: map[string]any{
			"run_id":   result.RunID,
			"event_id": result.EventID,
			"status":   result.Status.String(),
		},
	}

	return h.Success(ctx, out)
}
