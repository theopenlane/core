package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/ingest"
	"github.com/theopenlane/core/internal/integrations/operations"
	integrationscope "github.com/theopenlane/core/internal/integrations/scope"
	"github.com/theopenlane/core/internal/integrations/targetresolver"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// IntegrationRegistry exposes provider operation descriptors to the engine
type IntegrationRegistry interface {
	ResolveOperation(provider types.ProviderType, operationName types.OperationName, operationKind types.OperationKind) (types.OperationDescriptor, error)
}

// IntegrationStore ensures integration records exist for providers
type IntegrationStore interface {
	EnsureIntegration(ctx context.Context, orgID string, provider types.ProviderType) (*ent.Integration, error)
}

// IntegrationOperations executes provider operations
type IntegrationOperations interface {
	Run(ctx context.Context, req types.OperationRequest) (types.OperationResult, error)
}

// IntegrationDeps wires integration-specific dependencies into the workflow engine
type IntegrationDeps struct {
	// Registry provides access to integration operation descriptors
	Registry IntegrationRegistry
	// Store provides integration persistence for providers
	Store IntegrationStore
	// Operations executes provider operations through keystore
	Operations IntegrationOperations
	// MappingIndex resolves provider default mappings during ingest.
	MappingIndex types.MappingIndex
}

// IntegrationWorkflowMeta ties an integration run back to a workflow action
type IntegrationWorkflowMeta struct {
	// InstanceID identifies the workflow instance that triggered the run
	InstanceID string
	// ActionKey identifies the workflow action key
	ActionKey string
	// ActionIndex captures the workflow action index
	ActionIndex int
	// ObjectID identifies the workflow object
	ObjectID string
	// ObjectType identifies the workflow object type
	ObjectType enums.WorkflowObjectType
}

// IntegrationQueueRequest describes a queued integration operation
type IntegrationQueueRequest struct {
	// OrgID identifies the organization requesting the operation
	OrgID string
	// Provider identifies the integration provider
	Provider types.ProviderType
	// IntegrationID identifies the integration record
	IntegrationID string
	// Operation identifies the provider operation
	Operation types.OperationName
	// OperationKind identifies the provider operation kind when operation name is omitted
	OperationKind types.OperationKind
	// Config carries the operation configuration payload as a JSON object document
	Config json.RawMessage
	// ScopeExpression is an optional CEL expression gate for command execution
	ScopeExpression string
	// ScopePayload is optional data exposed to scope expression evaluation as a JSON object document
	ScopePayload json.RawMessage
	// ScopeResource is optional resource identity exposed to scope expression evaluation
	ScopeResource string
	// Force requests credential refresh
	Force bool
	// ClientForce requests client refresh
	ClientForce bool
	// RunType identifies the integration run type
	RunType enums.IntegrationRunType
	// WorkflowMeta links the operation to a workflow action
	WorkflowMeta *IntegrationWorkflowMeta
}

// IntegrationQueueResult captures queue results
type IntegrationQueueResult struct {
	// RunID identifies the integration run record
	RunID string
	// EventID identifies the emitted event
	EventID string
	// Status captures the run status at queue time
	Status enums.IntegrationRunStatus
}

// integrationActionRuntimeParams defines the integration action params used at runtime.
// Embeds IntegrationActionParams so the JSON schema type and the execution type stay in sync.
type integrationActionRuntimeParams struct {
	workflows.IntegrationActionParams
}

// integrationOpContext captures common integration operation log fields.
type integrationOpContext struct {
	provider      string
	operation     string
	operationKind string
	integrationID string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler for integrationOpContext.
func (c integrationOpContext) MarshalZerologObject(e *zerolog.Event) {
	e.Str("provider", c.provider).Str("operation", c.operation).Str("operation_kind", c.operationKind)
	if c.integrationID != "" {
		e.Str("integration_id", c.integrationID)
	}
}

// logIntegrationScopeSkipped logs a debug event when an integration action is skipped by scope evaluation.
func logIntegrationScopeSkipped(ctx context.Context, provider, operation, operationKind, integrationID, scopeExpression string) {
	logx.FromContext(ctx).Debug().EmbedObject(integrationOpContext{provider: provider, operation: operation, operationKind: operationKind, integrationID: integrationID}).Str("scope_expression", scopeExpression).Msg("integration action skipped by scope condition")
}

// SetIntegrationDeps attaches integration dependencies and registers listeners when possible
func (e *WorkflowEngine) SetIntegrationDeps(deps IntegrationDeps) error {
	if deps.Registry != nil {
		e.integrationRegistry = deps.Registry

		source, err := targetresolver.NewEntSource(e.client)
		if err != nil {
			return err
		}

		resolver, err := targetresolver.NewResolver(source)
		if err != nil {
			return err
		}

		e.integrationResolver = resolver

		evaluator, err := integrationscope.NewEvaluator(integrationscope.DefaultEvaluatorConfig())
		if err != nil {
			return err
		}

		e.scopeEvaluator = evaluator
	}
	if deps.MappingIndex != nil {
		e.integrationMappingIndex = deps.MappingIndex
	}
	if deps.Store != nil {
		e.integrationStore = deps.Store
	}
	if deps.Operations != nil {
		e.integrationOperations = deps.Operations
	}

	if e.integrationOperations == nil {
		return nil
	}

	if e.gala != nil {
		err := e.gala.ContextManager().Register(gala.NewKeyCodec("integration_execution", types.IntegrationExecutionContextKey()))
		if err != nil && !errors.Is(err, gala.ErrContextCodecAlreadyRegistered) {
			return err
		}
	}

	if e.gala == nil || e.integrationListenersRegistered {
		return nil
	}

	if _, err := gala.RegisterListeners(e.gala.Registry(),
		gala.Definition[integrations.IntegrationOperationEnvelope]{
			Topic: integrations.IntegrationOperationRequestedTopic,
			Name:  "integrations.operation.execute",
			Handle: func(ctx gala.HandlerContext, envelope integrations.IntegrationOperationEnvelope) error {
				return e.handleIntegrationOperationRequested(ctx, envelope)
			},
		},
	); err != nil {
		return err
	}

	e.integrationListenersRegistered = true
	return nil
}

// QueueIntegrationOperation queues an integration operation for async execution
func (e *WorkflowEngine) QueueIntegrationOperation(ctx context.Context, req IntegrationQueueRequest) (IntegrationQueueResult, error) {
	return e.queueIntegrationOperation(ctx, req)
}

// queueIntegrationOperation records the integration run and emits the request event
func (e *WorkflowEngine) queueIntegrationOperation(ctx context.Context, req IntegrationQueueRequest) (IntegrationQueueResult, error) {
	if e.integrationOperations == nil {
		return IntegrationQueueResult{}, ErrIntegrationOperationsRequired
	}

	orgID := req.OrgID
	if orgID == "" {
		return IntegrationQueueResult{}, ErrIntegrationOwnerRequired
	}

	operationName := req.Operation
	operationKind := req.OperationKind
	if operationName == "" && operationKind == "" {
		return IntegrationQueueResult{}, ErrIntegrationOperationCriteriaRequired
	}

	provider := req.Provider
	integrationID := req.IntegrationID

	allowCtx := workflows.AllowContext(ctx)
	if e.integrationRegistry == nil {
		return IntegrationQueueResult{}, ErrIntegrationRegistryRequired
	}

	criteria := targetresolver.ResolveCriteria{OwnerID: orgID}

	var integrationRecord *ent.Integration
	var operationDescriptor types.OperationDescriptor

	if integrationID == "" {
		if provider == types.ProviderUnknown {
			return IntegrationQueueResult{}, ErrIntegrationProviderRequired
		}
		resolvedOperation, err := e.integrationRegistry.ResolveOperation(provider, operationName, operationKind)
		if err != nil {
			return IntegrationQueueResult{}, err
		}
		if e.integrationStore == nil {
			return IntegrationQueueResult{}, ErrIntegrationStoreRequired
		}

		ensuredRecord, err := e.integrationStore.EnsureIntegration(allowCtx, orgID, provider)
		if err != nil {
			return IntegrationQueueResult{}, err
		}

		integrationRecord = ensuredRecord
		operationDescriptor = resolvedOperation
		operationName = operationDescriptor.Name
		operationKind = operationDescriptor.Kind
	} else {
		criteria.IntegrationID = mo.Some(integrationID)
		if provider != types.ProviderUnknown {
			criteria.Provider = mo.Some(provider)
		}

		resolution, err := e.integrationResolver.Resolve(allowCtx, criteria)
		if err != nil {
			return IntegrationQueueResult{}, err
		}

		integrationRecord = resolution.Integration
		provider = resolution.Provider
		operationDescriptor, err = e.integrationRegistry.ResolveOperation(provider, operationName, operationKind)
		if err != nil {
			return IntegrationQueueResult{}, err
		}
		operationName = operationDescriptor.Name
		operationKind = operationDescriptor.Kind
	}

	operationConfig, err := operations.ResolveOperationConfig(&integrationRecord.Config, string(operationName), req.Config)
	if err != nil {
		return IntegrationQueueResult{}, err
	}
	if operationConfig == nil {
		operationConfig = req.Config
	}

	scopeAllowed, err := evaluateIntegrationScope(allowCtx, e.scopeEvaluator, req, integrationRecord, provider, operationName, operationConfig, req.ScopePayload)
	if err != nil {
		return IntegrationQueueResult{}, err
	}
	if !scopeAllowed {
		return IntegrationQueueResult{}, ErrIntegrationScopeConditionFalse
	}

	if shouldEnsurePayloads(operationDescriptor.Ingest) {
		operationConfig = operations.EnsureIncludePayloads(operationConfig)
	}

	runType := req.RunType
	if runType == "" {
		runType = enums.IntegrationRunTypeEvent
	}

	runBuilder := e.client.IntegrationRun.Create().
		SetOwnerID(orgID).
		SetIntegrationID(integrationRecord.ID).
		SetOperationName(string(operationName)).
		SetOperationKind(integrationRunOperationKind(runType, operationKind)).
		SetRunType(runType).
		SetStatus(enums.IntegrationRunStatusPending)
	if len(operationConfig) > 0 {
		var configMap map[string]interface{}
		if err := json.Unmarshal(operationConfig, &configMap); err != nil {
			return IntegrationQueueResult{}, err
		}
		if configMap != nil {
			runBuilder.SetOperationConfig(configMap)
		}
	}

	runRecord, err := runBuilder.Save(allowCtx)
	if err != nil {
		return IntegrationQueueResult{}, err
	}

	payload := integrations.IntegrationOperationRequestedPayload{
		RunID:         runRecord.ID,
		OrgID:         orgID,
		Provider:      string(provider),
		Operation:     string(operationName),
		OperationKind: operationKind,
		RunType:       runType,
		Force:         req.Force,
		ClientForce:   req.ClientForce,
	}
	if req.WorkflowMeta != nil {
		payload.WorkflowInstanceID = req.WorkflowMeta.InstanceID
		payload.WorkflowActionKey = req.WorkflowMeta.ActionKey
		payload.WorkflowActionIndex = req.WorkflowMeta.ActionIndex
		payload.WorkflowObjectID = req.WorkflowMeta.ObjectID
		payload.WorkflowObjectType = req.WorkflowMeta.ObjectType
	}

	envelope := integrations.NewIntegrationOperationEnvelope(payload)
	emitCtx := types.WithIntegrationExecutionContext(ctx, types.IntegrationExecutionContext{
		OrgID:         orgID,
		IntegrationID: integrationRecord.ID,
		Provider:      provider,
		RunID:         runRecord.ID,
		Operation:     operationName,
	})

	receipt := workflows.EmitWorkflowEventWithHeaders(emitCtx, e.gala, integrations.IntegrationOperationRequestedTopic.Name, envelope, envelope.Headers())
	if receipt.Err != nil {
		if err := e.client.IntegrationRun.UpdateOneID(runRecord.ID).SetStatus(enums.IntegrationRunStatusFailed).SetError(receipt.Err.Error()).Exec(allowCtx); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("run_id", runRecord.ID).Msg("failed to mark integration run as failed after event emission error")
		}

		return IntegrationQueueResult{}, receipt.Err
	}

	if receipt.EventID != "" {
		if err := e.client.IntegrationRun.UpdateOneID(runRecord.ID).SetEventID(receipt.EventID).Exec(allowCtx); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("run_id", runRecord.ID).Msg("failed to update integration run event ID")
		}
	}

	return IntegrationQueueResult{
		RunID:   runRecord.ID,
		EventID: receipt.EventID,
		Status:  enums.IntegrationRunStatusPending,
	}, nil
}

// executeIntegrationAction queues an integration operation from a workflow action
func (e *WorkflowEngine) executeIntegrationAction(ctx context.Context, action models.WorkflowAction, instance *ent.WorkflowInstance, obj *workflows.Object) error {
	if e.integrationOperations == nil {
		return ErrIntegrationOperationsRequired
	}

	var params integrationActionRuntimeParams
	if err := jsonx.RoundTrip(action.Params, &params); err != nil {
		return errors.Join(ErrUnmarshalParams, err)
	}

	operationName := types.OperationName(params.OperationName)
	operationKind := types.OperationKind(params.OperationKind)
	if operationName == "" && operationKind == "" {
		return ErrIntegrationOperationCriteriaRequired
	}

	orgID := instance.OwnerID
	if orgID == "" {
		integCaller, integOk := auth.CallerFromContext(ctx)
		if !integOk || integCaller == nil || integCaller.OrganizationID == "" {
			return ErrIntegrationOwnerRequired
		}
		orgID = integCaller.OrganizationID
	}

	meta := &IntegrationWorkflowMeta{
		InstanceID:  instance.ID,
		ActionKey:   action.Key,
		ActionIndex: actionIndexForKey(instance.DefinitionSnapshot.Actions, action.Key),
	}
	if obj != nil {
		meta.ObjectID = obj.ID
		meta.ObjectType = obj.Type
	}

	_, err := e.queueIntegrationOperation(ctx, IntegrationQueueRequest{
		OrgID:           orgID,
		Provider:        types.ProviderTypeFromString(params.Provider),
		IntegrationID:   params.IntegrationID,
		Operation:       operationName,
		OperationKind:   operationKind,
		Config:          jsonx.CloneRawMessage(params.Config),
		ScopeExpression: params.ScopeExpression,
		ScopePayload:    jsonx.CloneRawMessage(params.ScopePayload),
		ScopeResource:   params.ScopeResource,
		Force:           params.Force,
		ClientForce:     params.ClientForce,
		RunType:         enums.IntegrationRunTypeEvent,
		WorkflowMeta:    meta,
	})
	if err != nil {
		if errors.Is(err, ErrIntegrationScopeConditionFalse) {
			logIntegrationScopeSkipped(ctx, params.Provider, string(operationName), string(operationKind), params.IntegrationID, params.ScopeExpression)
			return nil
		}
		return err
	}

	return ErrIntegrationActionQueued
}

// handleIntegrationOperationRequested executes integration operations triggered by events
func (e *WorkflowEngine) handleIntegrationOperationRequested(ctx gala.HandlerContext, envelope integrations.IntegrationOperationEnvelope) error {
	payload := envelope.Request

	if e.integrationOperations == nil {
		return ErrIntegrationOperationsRequired
	}

	runID := payload.RunID
	if runID == "" {
		return ErrIntegrationRunIDRequired
	}

	systemCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)
	run, err := e.client.IntegrationRun.Query().
		Where(integrationrun.IDEQ(runID)).
		WithIntegration().
		Only(systemCtx)
	if err != nil {
		return err
	}

	if run.Status != enums.IntegrationRunStatusPending {
		return nil
	}

	integrationRecord := run.Edges.Integration
	if integrationRecord == nil {
		return ErrIntegrationRecordMissing
	}

	provider := types.ProviderTypeFromString(integrationRecord.Kind)
	if provider == types.ProviderUnknown {
		return ErrIntegrationProviderUnknown
	}

	operationName := types.OperationName(run.OperationName)
	if operationName == "" {
		return ErrIntegrationOperationNameRequired
	}

	if e.integrationRegistry == nil {
		return ErrIntegrationRegistryRequired
	}

	operationDescriptor, err := e.integrationRegistry.ResolveOperation(provider, operationName, "")
	if err != nil {
		return err
	}

	startedAt := time.Now()
	update := e.client.IntegrationRun.UpdateOneID(run.ID).
		SetStatus(enums.IntegrationRunStatusRunning).
		SetStartedAt(startedAt)
	if eventID := string(ctx.Envelope.ID); eventID != "" {
		update.SetEventID(eventID)
	}
	if err := update.Exec(systemCtx); err != nil {
		return err
	}

	ingestContracts := operationDescriptor.Ingest

	var operationConfig json.RawMessage
	if len(run.OperationConfig) > 0 {
		operationConfig, err = json.Marshal(run.OperationConfig)
		if err != nil {
			return err
		}
	}
	if shouldEnsurePayloads(ingestContracts) {
		operationConfig = operations.EnsureIncludePayloads(operationConfig)
	}

	baseOperationCtx := types.WithIntegrationExecutionContext(systemCtx, types.IntegrationExecutionContext{
		OrgID:         run.OwnerID,
		IntegrationID: run.IntegrationID,
		Provider:      provider,
		RunID:         run.ID,
		Operation:     operationName,
	})

	operationCtx, cancel := integrationOperationContext(baseOperationCtx, envelope.TimeoutSeconds)
	defer cancel()

	result, opErr := e.integrationOperations.Run(operationCtx, types.OperationRequest{
		OrgID:         run.OwnerID,
		IntegrationID: run.IntegrationID,
		Provider:      provider,
		Name:          operationName,
		Config:        operationConfig,
		Force:         payload.Force,
		ClientForce:   payload.ClientForce,
	})

	runStatus := enums.IntegrationRunStatusSuccess
	summary := result.Summary
	errorText := ""
	if opErr != nil || result.Status != types.OperationStatusOK {
		runStatus = enums.IntegrationRunStatusFailed
		if opErr != nil {
			errorText = opErr.Error()
		} else {
			errorText = result.Summary
		}
	}

	metricsDoc := buildOperationMetrics(result)

	if runStatus == enums.IntegrationRunStatusSuccess && len(ingestContracts) > 0 {
		ingestBatches, err := extractIngestBatches(result.Details, ingestContracts)
		if err != nil {
			runStatus = enums.IntegrationRunStatusFailed
			errorText = err.Error()
		} else {
			for _, batch := range ingestBatches {
				ingestFn, ok := ingest.HandlerForSchema(batch.Schema)
				if !ok || ingestFn == nil {
					runStatus = enums.IntegrationRunStatusFailed
					errorText = fmt.Errorf("%w: schema=%s", ErrIntegrationAlertPayloadsMissing, batch.Schema).Error()
					break
				}

				ingestResult, ingestErr := ingestFn(operationCtx, ingest.IngestRequest{
					OrgID:             run.OwnerID,
					IntegrationID:     integrationRecord.ID,
					Provider:          provider,
					Operation:         operationName,
					IntegrationConfig: integrationRecord.Config,
					ProviderState:     integrationRecord.ProviderState,
					OperationConfig:   operationConfig,
					MappingIndex:      e.integrationMappingIndex,
					Envelopes:         batch.Envelopes,
					DB:                e.client,
				})

				metricsDoc = appendIngestMetrics(metricsDoc, batch.Schema, ingestResult)
				if ingestErr != nil {
					runStatus = enums.IntegrationRunStatusFailed
					errorText = ingestErr.Error()
					break
				}
			}
		}
	}

	finishedAt := time.Now()
	durationMs := int(finishedAt.Sub(startedAt).Milliseconds())
	metrics := encodeIntegrationRunMetrics(metricsDoc)
	finalize := e.client.IntegrationRun.UpdateOneID(run.ID).
		SetStatus(runStatus).
		SetFinishedAt(finishedAt).
		SetDurationMs(durationMs).
		SetSummary(summary).
		SetMetrics(metrics)
	if errorText != "" {
		finalize.SetError(errorText)
	}

	if err := finalize.Exec(systemCtx); err != nil {
		return err
	}

	if payload.WorkflowInstanceID != "" && payload.WorkflowActionIndex >= 0 {
		e.emitWorkflowActionCompleted(baseOperationCtx, envelope, runStatus, errorText)
	}

	if runStatus != enums.IntegrationRunStatusSuccess {
		return ErrIntegrationOperationFailed
	}

	return nil
}

// evaluateIntegrationScope evaluates optional scope expressions before queueing integration runs
func evaluateIntegrationScope(ctx context.Context, evaluator integrationscope.ConditionEvaluator, req IntegrationQueueRequest, integrationRecord *ent.Integration, provider types.ProviderType, operationName types.OperationName, operationConfig json.RawMessage, scopePayload json.RawMessage) (bool, error) {
	if evaluator == nil || req.ScopeExpression == "" {
		return true, nil
	}

	integrationConfigRaw, err := jsonx.ToRawMessage(integrationRecord.Config)
	if err != nil {
		return false, err
	}

	providerStateRaw, err := jsonx.ToRawMessage(integrationRecord.ProviderState)
	if err != nil {
		return false, err
	}

	return evaluator.EvaluateConditionWithVars(ctx, req.ScopeExpression, integrationscope.ScopeVars{
		Payload:           scopePayload,
		Resource:          req.ScopeResource,
		Provider:          provider,
		Operation:         operationName,
		Config:            operationConfig,
		IntegrationConfig: integrationConfigRaw,
		ProviderState:     providerStateRaw,
		OrgID:             req.OrgID,
		IntegrationID:     integrationRecord.ID,
	})
}

// integrationRunOperationKind maps operation descriptors into integration run operation kinds
func integrationRunOperationKind(runType enums.IntegrationRunType, operationKind types.OperationKind) enums.IntegrationOperationKind {
	switch runType {
	case enums.IntegrationRunTypeWebhook:
		return enums.IntegrationOperationKindWebhook
	case enums.IntegrationRunTypeScheduled:
		return enums.IntegrationOperationKindScheduled
	}

	switch operationKind {
	case types.OperationKindNotify:
		return enums.IntegrationOperationKindPush
	case types.OperationKindCollectFindings, types.OperationKindScanSettings:
		return enums.IntegrationOperationKindPull
	default:
		return enums.IntegrationOperationKindSync
	}
}

// integrationOperationContext applies optional timeout policy to operation execution
func integrationOperationContext(parent context.Context, timeoutSeconds int) (context.Context, context.CancelFunc) {
	if timeoutSeconds <= 0 {
		return parent, func() {}
	}

	return context.WithTimeout(parent, time.Duration(timeoutSeconds)*time.Second)
}

// emitWorkflowActionCompleted emits a completion event after an integration run finishes
func (e *WorkflowEngine) emitWorkflowActionCompleted(ctx context.Context, envelope integrations.IntegrationOperationEnvelope, status enums.IntegrationRunStatus, errorText string) {
	payload := envelope.Request

	if e.gala == nil {
		return
	}

	actionPayload := gala.WorkflowActionCompletedPayload{
		InstanceID:  payload.WorkflowInstanceID,
		ActionIndex: payload.WorkflowActionIndex,
		ActionType:  enums.WorkflowActionTypeIntegration,
		ObjectID:    payload.WorkflowObjectID,
		ObjectType:  payload.WorkflowObjectType,
		Success:     status == enums.IntegrationRunStatusSuccess,
		Skipped:     false,
	}
	if errorText != "" {
		actionPayload.ErrorMessage = errorText
	}

	receipt := workflows.EmitWorkflowEvent(ctx, e.gala, gala.TopicWorkflowActionCompleted, actionPayload)
	if receipt.Err != nil {
		logx.FromContext(ctx).Warn().Err(receipt.Err).Msg("failed to emit workflow action completed for integration run")
	}
}

type ingestBatch struct {
	Schema    types.MappingSchema   `json:"schema"`
	Envelopes []types.AlertEnvelope `json:"envelopes"`
}

// shouldEnsurePayloads reports whether any ingest contract requires include_payloads=true.
func shouldEnsurePayloads(contracts []types.IngestContract) bool {
	return lo.SomeBy(contracts, func(contract types.IngestContract) bool { return contract.EnsurePayloads })
}

// extractIngestBatches pulls ingest envelope batches from operation details.
func extractIngestBatches(details json.RawMessage, contracts []types.IngestContract) ([]ingestBatch, error) {
	if len(details) == 0 {
		return nil, ErrIntegrationAlertPayloadsMissing
	}

	var payload struct {
		IngestBatches []ingestBatch   `json:"ingest_batches"`
		Alerts        json.RawMessage `json:"alerts"`
	}
	if err := jsonx.RoundTrip(details, &payload); err != nil {
		return nil, err
	}

	contractIndex := lo.SliceToMap(contracts, func(contract types.IngestContract) (types.MappingSchema, struct{}) {
		return types.NormalizeMappingSchema(contract.Schema), struct{}{}
	})

	if len(payload.IngestBatches) > 0 {
		out := make([]ingestBatch, 0, len(payload.IngestBatches))
		for _, batch := range payload.IngestBatches {
			schema := types.NormalizeMappingSchema(batch.Schema)
			if schema == "" {
				return nil, fmt.Errorf("%w: empty ingest schema", ErrIntegrationAlertPayloadsMissing)
			}
			if _, ok := contractIndex[schema]; !ok {
				return nil, fmt.Errorf("%w: undeclared ingest schema=%s", ErrIntegrationAlertPayloadsMissing, schema)
			}

			out = append(out, ingestBatch{
				Schema:    schema,
				Envelopes: batch.Envelopes,
			})
		}

		return out, nil
	}

	if len(payload.Alerts) == 0 || len(contracts) == 0 {
		return nil, ErrIntegrationAlertPayloadsMissing
	}

	if len(contracts) != 1 {
		return nil, fmt.Errorf("%w: %d contracts require ingest_batches output", ErrIntegrationAlertPayloadsMissing, len(contracts))
	}

	var envelopes []types.AlertEnvelope
	if err := jsonx.RoundTrip(payload.Alerts, &envelopes); err == nil {
		return []ingestBatch{{
			Schema:    types.NormalizeMappingSchema(contracts[0].Schema),
			Envelopes: envelopes,
		}}, nil
	}

	var envelope types.AlertEnvelope
	if err := jsonx.RoundTrip(payload.Alerts, &envelope); err != nil {
		return nil, err
	}

	return []ingestBatch{{
		Schema:    types.NormalizeMappingSchema(contracts[0].Schema),
		Envelopes: []types.AlertEnvelope{envelope},
	}}, nil
}

// integrationRunOperationMetrics captures operation execution metrics
type integrationRunOperationMetrics struct {
	// Status is the operation status value
	Status string `json:"status"`
	// Summary is the operation summary value
	Summary string `json:"summary"`
	// Details captures optional operation details
	Details json.RawMessage `json:"details,omitempty"`
}

// integrationRunMetrics captures integration run metrics persisted on integration runs
type integrationRunMetrics struct {
	// Operation captures operation execution metrics
	Operation integrationRunOperationMetrics `json:"operation"`
	// IngestSummaries captures ingest summary metrics keyed by schema.
	IngestSummaries map[string]json.RawMessage `json:"ingest_summaries,omitempty"`
	// IngestErrors captures ingest error metrics keyed by schema.
	IngestErrors map[string][]string `json:"ingest_errors,omitempty"`
}

// encodeIntegrationRunOperationDetails normalizes operation details for metrics payloads
func encodeIntegrationRunOperationDetails(details json.RawMessage) json.RawMessage {
	return jsonx.CloneRawMessage(details)
}

// buildOperationMetrics builds a metrics document for operation output
func buildOperationMetrics(result types.OperationResult) integrationRunMetrics {
	return integrationRunMetrics{
		Operation: integrationRunOperationMetrics{
			Status:  string(result.Status),
			Summary: result.Summary,
			Details: encodeIntegrationRunOperationDetails(result.Details),
		},
	}
}

// encodeIntegrationRunMetrics converts integration run metrics into persisted map form
func encodeIntegrationRunMetrics(metrics integrationRunMetrics) map[string]any {
	metricsMap, err := jsonx.ToMap(metrics)
	if err != nil {
		return nil
	}

	return metricsMap
}

// appendIngestMetrics attaches ingest results for one schema to integration run metrics.
func appendIngestMetrics(metrics integrationRunMetrics, schema types.MappingSchema, result ingest.IngestResult) integrationRunMetrics {
	normalizedSchema := types.NormalizeMappingSchema(schema)
	if normalizedSchema == "" {
		return metrics
	}

	raw, err := jsonx.ToRawMessage(result.Summary)
	if err == nil {
		if metrics.IngestSummaries == nil {
			metrics.IngestSummaries = map[string]json.RawMessage{}
		}
		metrics.IngestSummaries[string(normalizedSchema)] = raw
	}

	if len(result.Errors) > 0 {
		if metrics.IngestErrors == nil {
			metrics.IngestErrors = map[string][]string{}
		}
		metrics.IngestErrors[string(normalizedSchema)] = append([]string(nil), result.Errors...)
	}

	return metrics
}
