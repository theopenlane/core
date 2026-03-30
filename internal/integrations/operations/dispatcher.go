package operations

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Dispatch validates and enqueues one operation execution request
func Dispatch(ctx context.Context, reg *registry.Registry, db *ent.Client, runtime *gala.Gala, req DispatchRequest) (DispatchResult, error) {
	if req.IntegrationID == "" || req.Operation == "" {
		return DispatchResult{}, ErrDispatchInputInvalid
	}

	installationRecord, err := db.Integration.Get(ctx, req.IntegrationID)
	if err != nil {
		return DispatchResult{}, err
	}

	operation, err := reg.Operation(installationRecord.DefinitionID, req.Operation)
	if err != nil {
		return DispatchResult{}, err
	}

	if err := ValidateConfig(operation.ConfigSchema, req.Config); err != nil {
		if errors.Is(err, ErrOperationConfigInvalid) {
			return DispatchResult{}, ErrDispatchInputInvalid
		}

		return DispatchResult{}, err
	}

	runType := req.RunType
	if runType == "" {
		runType = enums.IntegrationRunTypeManual
	}

	runRecord, err := CreatePendingRun(ctx, db, installationRecord, DispatchRequest{
		IntegrationID:      req.IntegrationID,
		Operation:          req.Operation,
		Config:             jsonx.CloneRawMessage(req.Config),
		ForceClientRebuild: req.ForceClientRebuild,
		RunType:            runType,
	})
	if err != nil {
		return DispatchResult{}, err
	}

	metadata := types.ExecutionMetadata{
		OwnerID:       installationRecord.OwnerID,
		IntegrationID: installationRecord.ID,
		DefinitionID:  installationRecord.DefinitionID,
		Operation:     req.Operation,
		RunID:         runRecord.ID,
		RunType:       runType,
		Workflow:      req.Workflow,
	}

	// Inherit webhook/event context from the parent execution when dispatching
	// from a webhook handler so the envelope carries the triggering event identity
	if existing, ok := types.ExecutionMetadataFromContext(ctx); ok {
		if metadata.Webhook == "" {
			metadata.Webhook = existing.Webhook
		}

		if metadata.Event == "" {
			metadata.Event = existing.Event
		}

		if metadata.DeliveryID == "" {
			metadata.DeliveryID = existing.DeliveryID
		}
	}

	emitCtx := types.WithExecutionMetadata(ctx, metadata)
	receipt := runtime.EmitWithHeaders(emitCtx, operation.Topic, Envelope{
		ExecutionMetadata:  metadata,
		Config:             jsonx.CloneRawMessage(req.Config),
		ForceClientRebuild: req.ForceClientRebuild,
	}, gala.Headers{
		IdempotencyKey: runRecord.ID,
		Properties:     metadata.Properties(),
	})

	if receipt.Err != nil {
		if completeErr := CompleteRun(ctx, db, runRecord.ID, time.Now(), RunResult{
			Status:  enums.IntegrationRunStatusFailed,
			Summary: "dispatch failed",
			Error:   receipt.Err.Error(),
		}); completeErr != nil {
			return DispatchResult{}, errors.Join(receipt.Err, completeErr)
		}

		return DispatchResult{}, receipt.Err
	}

	return DispatchResult{
		RunID:   runRecord.ID,
		EventID: string(receipt.EventID),
		Status:  enums.IntegrationRunStatusPending,
	}, nil
}

// ValidateConfig validates one raw configuration payload against the operation schema
func ValidateConfig(schema json.RawMessage, value json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}

	var document any = map[string]any{}
	if err := jsonx.UnmarshalIfPresent(value, &document); err != nil {
		return ErrOperationConfigInvalid
	}

	result, err := jsonx.ValidateSchema(schema, document)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	return ErrOperationConfigInvalid
}
