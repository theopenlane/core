package handlers

import (
	"context"
	"errors"
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
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

	requestCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(req.Provider)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !def.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	operationName := req.Body.Operation
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}
	integrationID := req.IntegrationID

	operation, err := h.IntegrationsRuntime.Registry().Operation(def.ID, operationName)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", def.ID).Str("operation", operationName).Msg("operation not registered")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	inlineExecution := operationName == "health.default" || operation.Policy.Inline
	if h.WorkflowEngine == nil && !inlineExecution {
		return h.InternalServerError(ctx, errIntegrationWorkflowEngineNotConfigured, openapiCtx)
	}

	queueCtx := context.WithoutCancel(requestCtx)
	configDoc := jsonx.CloneRawMessage(req.Body.Config)

	if integrationID != "" {
		installationRec, err := h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, integrationID, def.ID)
		if err != nil {
			if errors.Is(err, integrationsruntime.ErrInstallationDefinitionMismatch) {
				return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
			}
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		integrationID = installationRec.ID
	}

	if inlineExecution {
		var (
			installationRec *ent.Integration
			err             error
		)

		if integrationID != "" {
			installationRec, err = h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, integrationID, def.ID)
		} else {
			installationRec, err = h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, "", def.ID)
		}
		if err != nil {
			if errors.Is(err, integrationsruntime.ErrInstallationIDRequired) {
				return h.BadRequest(ctx, ErrIntegrationIDRequired, openapiCtx)
			}
			if errors.Is(err, integrationsruntime.ErrInstallationDefinitionMismatch) {
				return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
			}

			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		credential, _, err := h.IntegrationsRuntime.LoadCredential(requestCtx, installationRec)
		if err != nil {
			return h.InternalServerError(ctx, wrapIntegrationError("load credential for", err), openapiCtx)
		}

		output, err := h.IntegrationsRuntime.ExecuteOperation(queueCtx, installationRec, operation, credential, configDoc)
		if err != nil {
			if operationName == "health.default" {
				return h.BadRequest(ctx, ErrProviderHealthCheckFailed, openapiCtx)
			}

			return h.BadRequest(ctx, wrapIntegrationError("execute", err), openapiCtx)
		}

		summary := "Integration operation completed"
		if operationName == "health.default" {
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

	result, err := h.WorkflowEngine.QueueIntegrationOperation(queueCtx, engine.IntegrationQueueRequest{
		OrgID:          caller.OrganizationID,
		DefinitionID:   def.ID,
		InstallationID: integrationID,
		Operation:      operationName,
		Config:         configDoc,
		Force:          req.Body.Force,
		RunType:        enums.IntegrationRunTypeManual,
	})
	if err != nil {
		switch integrationHTTPStatus(err) {
		case http.StatusBadRequest:
			return h.BadRequest(ctx, err, openapiCtx)
		case http.StatusNotFound:
			if errors.Is(err, engine.ErrInstallationNotFound) {
				return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
			}
			return h.NotFound(ctx, err, openapiCtx)
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

	out := IntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  def.Slug,
		Operation: operationName,
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details:   queueDetails,
	}

	return h.Success(ctx, out)
}
