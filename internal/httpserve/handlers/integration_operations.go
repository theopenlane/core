package handlers

import (
	"context"
	"errors"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	integrationsruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
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
func (h *Handler) RunIntegrationOperation(ctx echo.Context) error {
	req, err := BindAndValidate[RunIntegrationOperationRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser)
	}

	if req.IntegrationID == "" || req.Body.Operation == "" {
		// not terribly concerned about distinct error responses here since this isn't intended to be used as a primary execution method
		logx.FromContext(requestCtx).Error().Err(ErrIntegrationIDRequired).Msg("missing integrationID or Operation in request")

		return h.BadRequest(ctx, ErrIntegrationIDRequired)
	}

	integrationRef, def, operation, err := h.resolveInvocableOperation(requestCtx, caller.OrganizationID, req)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	if operation.RequiresPaymentMethod {
		if err := rule.RequirePaymentMethod()(requestCtx, nil); err != nil && !errors.Is(err, privacy.Skip) {
			return h.Forbidden(ctx, err)
		}
	}

	if operation.Policy.Inline {
		return h.runInlineOperation(ctx, req, integrationRef, def, operation)
	}

	result, err := h.IntegrationsRuntime.Dispatch(context.WithoutCancel(requestCtx), types.DispatchRequest{
		IntegrationID: integrationRef.ID,
		Operation:     req.Body.Operation,
		Config:        jsonx.CloneRawMessage(req.Body.Config),
		RunType:       enums.IntegrationRunTypeManual,
	})
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("failed to queue operation")

		return h.BadRequest(ctx, err)
	}

	queueDetails, err := jsonx.ToRawMessage(integrationOperationQueueDetails{
		RunID:   result.RunID,
		EventID: result.EventID,
		Status:  result.Status.String(),
	})
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	return h.Success(ctx, RunIntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  def.ID,
		Operation: req.Body.Operation,
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details:   queueDetails,
	})
}

// resolveInvocableOperation resolves the caller's installation, its active definition, and the requested operation, returning the error to answer with when the operation cannot be invoked directly
func (h *Handler) resolveInvocableOperation(ctx context.Context, ownerID string, req *RunIntegrationOperationRequest) (*ent.Integration, types.Definition, types.OperationRegistration, error) {
	integrationRef, err := h.IntegrationsRuntime.ResolveIntegration(ctx, integrationsruntime.IntegrationLookup{
		IntegrationID: req.IntegrationID,
		OwnerID:       ownerID,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Interface("request", req).Msg("failed to resolve installation")

		return nil, types.Definition{}, types.OperationRegistration{}, ErrIntegrationNotFound
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(integrationRef.DefinitionID)
	if !ok {
		logx.FromContext(ctx).Error().Str("definitionID", integrationRef.DefinitionID).Msg("definition not found in registry")

		return nil, types.Definition{}, types.OperationRegistration{}, ErrIntegrationNotFound
	}

	if !def.Active {
		logx.FromContext(ctx).Error().Err(ErrProviderDisabled).Str("definitionID", integrationRef.DefinitionID).Msg("integration provider is disabled, not executing operation")

		return nil, types.Definition{}, types.OperationRegistration{}, ErrProviderDisabled
	}

	operation, err := h.IntegrationsRuntime.Registry().Operation(def.ID, req.Body.Operation)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Interface("request", req).Msg("operation not found")

		return nil, types.Definition{}, types.OperationRegistration{}, operations.ErrDispatchInputInvalid
	}

	if operation.Internal {
		logx.FromContext(ctx).Error().Interface("request", req).Msg("operation is internal and cannot be invoked directly")

		return nil, types.Definition{}, types.OperationRegistration{}, operations.ErrDispatchInputInvalid
	}

	return integrationRef, def, operation, nil
}

// runInlineOperation validates the supplied config and executes the operation synchronously, answering rate limiting with a too-many-requests response
func (h *Handler) runInlineOperation(ctx echo.Context, req *RunIntegrationOperationRequest, integrationRef *ent.Integration, def types.Definition, operation types.OperationRegistration) error {
	requestCtx := ctx.Request().Context()
	configDoc := jsonx.CloneRawMessage(req.Body.Config)

	if err := operations.ValidateInput(requestCtx, types.InstallationRequest{Integration: integrationRef}, operation.Input.Schema, nil, configDoc, types.ErrOperationConfigInvalid); err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Msg("invalid operation config")

		return h.BadRequest(ctx, operations.ErrDispatchInputInvalid)
	}

	output, err := h.IntegrationsRuntime.ExecuteOperation(context.WithoutCancel(requestCtx), integrationRef, operation, configDoc)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", req).Msg("operation execution failed")

		if errors.Is(err, integrationsruntime.ErrOperationRateLimited) {
			return h.TooManyRequests(ctx, err)
		}

		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, RunIntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  def.ID,
		Operation: req.Body.Operation,
		Status:    "ok",
		Summary:   "Integration operation completed",
		Details:   output,
	})
}
