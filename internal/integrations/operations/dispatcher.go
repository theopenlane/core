package operations

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// Dispatch validates and enqueues one operation execution request
func Dispatch(ctx context.Context, reg *registry.Registry, db *ent.Client, runtime *gala.Gala, req types.DispatchRequest) (types.DispatchResult, error) {
	if req.Operation == "" || (!req.Runtime && req.IntegrationID == "") {
		return types.DispatchResult{}, ErrDispatchInputInvalid
	}

	var (
		definitionID string
		ownerID      = req.OwnerID
		installation *ent.Integration
		userInput    json.RawMessage
	)

	switch {
	case req.Runtime:
		definitionID = req.DefinitionID
	default:
		record, err := ResolveIntegration(ctx, db, req.IntegrationID, req.OwnerID, req.DefinitionID)
		if err != nil {
			return types.DispatchResult{}, err
		}

		installation = record
		userInput = record.Config.ClientConfig
		definitionID = record.DefinitionID
		ownerID = record.OwnerID

		ctx = auth.EnsureIntegrationCaller(ctx, record.OwnerID)
	}

	operation, err := reg.Operation(definitionID, req.Operation)
	if err != nil {
		return types.DispatchResult{}, err
	}

	if operation.DisabledFor(userInput) {
		logx.FromContext(ctx).Debug().Str(intobvs.FieldOperation, req.Operation).Msg("operation is disabled, skipping dispatch")

		return types.DispatchResult{Status: enums.IntegrationRunStatusCancelled}, nil
	}

	if err := ValidateConfig(operation.ConfigSchema, req.Config); err != nil {
		if errors.Is(err, ErrOperationConfigInvalid) {
			return types.DispatchResult{}, ErrDispatchInputInvalid
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
		runRecord, err := CreatePendingRun(ctx, db, installation, operation, runType, req.Config)
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

	eventID, err := runtime.EmitWithHeaders(emitCtx, operation.Topic, Envelope{
		OperationContext:   oc,
		Config:             req.Config,
		ForceClientRebuild: req.ForceClientRebuild,
	}, headers, gala.WithEventID(gala.EventID(runID)))
	if err != nil {
		if runID != "" {
			if completeErr := CompleteRun(ctx, db, runID, time.Now(), RunResult{
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

// ResolveIntegration resolves one integration by ID with optional owner and definition checks
func ResolveIntegration(ctx context.Context, db *ent.Client, integrationID, ownerID, definitionID string) (*ent.Integration, error) {
	if integrationID == "" {
		return nil, ErrIntegrationIDRequired
	}

	query := db.Integration.Query().Where(integration.IDEQ(integrationID))
	if ownerID != "" {
		query = query.Where(integration.OwnerIDEQ(ownerID))
	}

	record, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	if definitionID != "" && record.DefinitionID != definitionID {
		return nil, ErrInstallationDefinitionMismatch
	}

	return record, nil
}

// ResolveOwnerIntegration finds an operational integration for the given definition and owner
func ResolveOwnerIntegration(ctx context.Context, db *ent.Client, definitionID, ownerID string, prefer ...func(*ent.Integration) bool) (string, error) {
	integrations, err := db.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(definitionID),
			integration.StatusIn(enums.IntegrationOperationalStatuses...),
		).All(ctx)
	if err != nil {
		return "", err
	}

	switch {
	case len(integrations) == 1:
		return integrations[0].ID, nil
	case len(prefer) == 0:
		return "", nil
	}

	preferred, found := lo.Find(integrations, prefer[0])
	if !found {
		return "", nil
	}

	return preferred.ID, nil
}

// inheritWebhookContext propagates webhook/event context from a parent execution
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
