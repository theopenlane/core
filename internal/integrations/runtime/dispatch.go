package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// Dispatch validates and enqueues one operation execution request
func (r *Runtime) Dispatch(ctx context.Context, req types.DispatchRequest) (types.DispatchResult, error) {
	if req.Operation == "" || (!req.Runtime && req.IntegrationID == "") {
		return types.DispatchResult{}, operations.ErrDispatchInputInvalid
	}

	var (
		definitionID string
		ownerID      = req.OwnerID
		installation *ent.Integration
		input        json.RawMessage
	)

	switch {
	case req.Runtime:
		definitionID = req.DefinitionID
	default:
		record, err := r.ResolveIntegration(ctx, IntegrationLookup{IntegrationID: req.IntegrationID, OwnerID: req.OwnerID, DefinitionID: req.DefinitionID})
		if err != nil {
			return types.DispatchResult{}, err
		}

		if err := r.ensureCurrentVersion(auth.EnsureIntegrationCaller(ctx, record.OwnerID), record); err != nil {
			return types.DispatchResult{}, err
		}

		installation = record
		definitionID = record.DefinitionID
		ownerID = record.OwnerID

		ctx = auth.EnsureIntegrationCaller(ctx, record.OwnerID)
	}

	operation, err := r.Registry().Operation(definitionID, req.Operation)
	if err != nil {
		return types.DispatchResult{}, normalizeDispatchError(err)
	}

	if installation != nil {
		input = installation.OperationConfig.For(operation.Name)
	}

	if operation.DisabledFor(input) {
		logx.FromContext(ctx).Debug().Str(intobvs.FieldOperation, req.Operation).Msg("operation is disabled, skipping dispatch")

		return types.DispatchResult{Status: enums.IntegrationRunStatusCancelled}, nil
	}

	if err := operations.ValidateInput(ctx, types.InstallationRequest{Integration: installation}, operation.Input.Schema, nil, req.Config, types.ErrOperationConfigInvalid); err != nil {
		if errors.Is(err, types.ErrOperationConfigInvalid) {
			return types.DispatchResult{}, operations.ErrDispatchInputInvalid
		}

		return types.DispatchResult{}, err
	}

	runType := lo.CoalesceOrEmpty(req.RunType, enums.IntegrationRunTypeManual)

	src := types.IntegrationSource{
		IntegrationID: req.IntegrationID,
		DefinitionID:  definitionID,
		RunType:       runType,
		Workflow:      req.Workflow,
		Runtime:       req.Runtime,
	}

	var runID string

	if installation != nil && !operation.Policy.SkipRunRecord {
		runRecord, err := operations.CreatePendingRun(ctx, r.DB(), installation, operation, runType, req.Config)
		if err != nil {
			return types.DispatchResult{}, err
		}

		runID = runRecord.ID
		src.RunID = runID
	}

	inheritWebhookContext(ctx, &src)

	oc := types.NewOperationContext(ownerID, req.Operation, src)

	emitCtx, headers := intobvs.EmitContext(ctx, oc)
	headers.ScheduledAt = req.ScheduledAt

	if req.UniqueKey != "" {
		headers.UniqueKey = req.UniqueKey
		headers.UniqueOnce = true
	}

	eventID, err := r.Gala().EmitWithHeaders(emitCtx, operation.Topic, operations.Envelope{
		OperationContext:   oc,
		Config:             req.Config,
		ForceClientRebuild: req.ForceClientRebuild,
	}, headers, gala.WithEventID(gala.EventID(runID)))
	if err != nil {
		if runID != "" {
			if completeErr := operations.CompleteRun(ctx, r.DB(), runID, time.Now(), operations.RunResult{
				Status:  enums.IntegrationRunStatusFailed,
				Summary: "dispatch failed",
				Error:   err.Error(),
			}); completeErr != nil {
				return types.DispatchResult{}, errors.Join(err, completeErr)
			}
		}

		return types.DispatchResult{}, err
	}

	return types.DispatchResult{
		RunID:   runID,
		EventID: string(eventID),
		Status:  enums.IntegrationRunStatusPending,
	}, nil
}

// normalizeDispatchError translates registry lookup errors into runtime sentinel errors
func normalizeDispatchError(err error) error {
	switch {
	case errors.Is(err, registry.ErrDefinitionNotFound):
		return ErrDefinitionNotFound
	case errors.Is(err, registry.ErrOperationNotFound):
		return ErrOperationNotFound
	default:
		return err
	}
}

// inheritWebhookContext propagates webhook and event context from a parent execution
func inheritWebhookContext(ctx context.Context, src *types.IntegrationSource) {
	oc, ok := gala.OperationContextFromContext(ctx)
	if !ok {
		return
	}

	existing := types.IntegrationSourceFrom(oc)

	if src.Webhook == "" {
		src.Webhook = existing.Webhook
	}

	if src.Event == "" {
		src.Event = existing.Event
	}

	if src.DeliveryID == "" {
		src.DeliveryID = existing.DeliveryID
	}
}
