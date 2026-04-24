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

// Dispatch validates and enqueues one operation execution request. When
// DispatchRequest.Runtime is true, no DB integration lookup is performed and
// the client is resolved from the registry at execution time
func Dispatch(ctx context.Context, reg *registry.Registry, db *ent.Client, runtime *gala.Gala, req DispatchRequest) (DispatchResult, error) {
	if req.Operation == "" || (!req.Runtime && req.IntegrationID == "") {
		return DispatchResult{}, ErrDispatchInputInvalid
	}

	var (
		definitionID string
		ownerID      = req.OwnerID
		installation *ent.Integration
	)

	switch {
	case req.Runtime:
		definitionID = req.DefinitionID
	default:
		record, err := db.Integration.Get(ctx, req.IntegrationID)
		if err != nil {
			return DispatchResult{}, err
		}

		installation = record
		definitionID = record.DefinitionID
		ownerID = record.OwnerID
	}

	operation, err := reg.Operation(definitionID, req.Operation)
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

	metadata := types.ExecutionMetadata{
		OwnerID:       ownerID,
		IntegrationID: req.IntegrationID,
		DefinitionID:  definitionID,
		Operation:     req.Operation,
		RunType:       runType,
		Workflow:      req.Workflow,
		Runtime:       req.Runtime,
	}

	if installation != nil && !operation.Policy.SkipRunRecord {
		runRecord, err := CreatePendingRun(ctx, db, installation, DispatchRequest{
			IntegrationID:      req.IntegrationID,
			Operation:          req.Operation,
			Config:             jsonx.CloneRawMessage(req.Config),
			ForceClientRebuild: req.ForceClientRebuild,
			RunType:            runType,
		})
		if err != nil {
			return DispatchResult{}, err
		}

		metadata.RunID = runRecord.ID
	}

	inheritWebhookContext(&metadata, ctx)

	emitCtx := types.WithExecutionMetadata(ctx, metadata)
	receipt := runtime.EmitWithHeaders(emitCtx, operation.Topic, Envelope{
		ExecutionMetadata:  metadata,
		Config:             jsonx.CloneRawMessage(req.Config),
		ForceClientRebuild: req.ForceClientRebuild,
	}, gala.Headers{
		IdempotencyKey: metadata.RunID,
		Properties:     metadata.Properties(),
		ScheduledAt:    req.ScheduledAt,
	})

	if receipt.Err != nil {
		if metadata.RunID != "" {
			if completeErr := CompleteRun(ctx, db, metadata.RunID, time.Now(), RunResult{
				Status:  enums.IntegrationRunStatusFailed,
				Summary: "dispatch failed",
				Error:   receipt.Err.Error(),
			}); completeErr != nil {
				return DispatchResult{}, errors.Join(receipt.Err, completeErr)
			}
		}

		return DispatchResult{}, receipt.Err
	}

	return DispatchResult{
		RunID:   metadata.RunID,
		EventID: string(receipt.EventID),
		Status:  enums.IntegrationRunStatusPending,
	}, nil
}

// inheritWebhookContext propagates webhook/event context from a parent execution
// so the envelope carries the triggering event identity
func inheritWebhookContext(metadata *types.ExecutionMetadata, ctx context.Context) {
	existing, ok := types.ExecutionMetadataFromContext(ctx)
	if !ok {
		return
	}

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
