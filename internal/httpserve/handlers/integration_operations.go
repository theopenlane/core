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
	"github.com/theopenlane/core/internal/ent/generated/integration"
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
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleIntegrationOperationPayload, IntegrationOperationResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	requestCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(types.DefinitionID(req.Provider))
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !def.Spec.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	operationName := types.OperationName(req.Body.Operation)
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}
	integrationID := req.IntegrationID

	if h.WorkflowEngine == nil && operationName != "health.default" {
		return h.InternalServerError(ctx, errIntegrationWorkflowEngineNotConfigured, openapiCtx)
	}

	queueCtx := context.WithoutCancel(requestCtx)
	configDoc := jsonx.CloneRawMessage(req.Body.Config)

	if integrationID != "" {
		installationRec, err := h.DBClient.Integration.Query().
			Where(
				integration.OwnerIDEQ(caller.OrganizationID),
				integration.IDEQ(integrationID),
			).
			Only(requestCtx)
		if err != nil {
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		if installationRec.DefinitionID != string(def.Spec.ID) {
			return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
		}

		integrationID = installationRec.ID
	}

	if operationName == "health.default" {
		var (
			installationRec *ent.Integration
			err             error
		)

		if integrationID != "" {
			installationRec, err = h.DBClient.Integration.Query().
				Where(
					integration.OwnerIDEQ(caller.OrganizationID),
					integration.IDEQ(integrationID),
				).
				Only(requestCtx)
		} else {
			installationRec, err = h.DBClient.Integration.Query().
				Where(
					integration.OwnerIDEQ(caller.OrganizationID),
					integration.DefinitionIDEQ(string(def.Spec.ID)),
				).
				Only(requestCtx)
		}
		if err != nil {
			if ent.IsNotSingular(err) {
				return h.BadRequest(ctx, ErrIntegrationIDRequired, openapiCtx)
			}

			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		if installationRec.DefinitionID != string(def.Spec.ID) {
			return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
		}

		credential, _, err := h.IntegrationsRuntime.CredentialStore().LoadCredential(requestCtx, installationRec)
		if err != nil {
			return h.InternalServerError(ctx, wrapIntegrationError("load credential for", err), openapiCtx)
		}

		output, err := h.IntegrationsRuntime.Executor().ExecuteOperation(queueCtx, installationRec, operationName, credential, configDoc)
		if err != nil {
			return h.BadRequest(ctx, ErrProviderHealthCheckFailed, openapiCtx)
		}

		return h.Success(ctx, IntegrationOperationResponse{
			Reply:     rout.Reply{Success: true},
			Provider:  def.Spec.Slug,
			Operation: string(operationName),
			Status:    "ok",
			Summary:   "Health check completed",
			Details:   output,
		})
	}

	result, err := h.WorkflowEngine.QueueIntegrationOperation(queueCtx, engine.IntegrationQueueRequest{
		OrgID:          caller.OrganizationID,
		DefinitionID:   string(def.Spec.ID),
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
		Provider:  def.Spec.Slug,
		Operation: string(operationName),
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details:   queueDetails,
	}

	return h.Success(ctx, out)
}
