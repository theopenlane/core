package engine

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/ingest"
	integrationscope "github.com/theopenlane/core/internal/integrations/scope"
	"github.com/theopenlane/core/internal/integrations/targetresolver"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
	"github.com/theopenlane/iam/auth"
)

// IntegrationRegistry exposes provider operation descriptors to the engine
type IntegrationRegistry interface {
	OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor
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

		resolver, err := targetresolver.NewResolver(source, deps.Registry)
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
	if deps.Store != nil {
		e.integrationStore = deps.Store
	}
	if deps.Operations != nil {
		e.integrationOperations = deps.Operations
	}

	if e.integrationOperations == nil {
		return nil
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
	if operationName != "" {
		criteria.OperationName = mo.Some(operationName)
	}
	if operationKind != "" {
		criteria.OperationKind = mo.Some(operationKind)
	}

	var integrationRecord *ent.Integration

	if integrationID == "" {
		if provider == types.ProviderUnknown {
			return IntegrationQueueResult{}, ErrIntegrationProviderRequired
		}
		if e.integrationStore == nil {
			return IntegrationQueueResult{}, ErrIntegrationStoreRequired
		}

		ensuredRecord, err := e.integrationStore.EnsureIntegration(allowCtx, orgID, provider)
		if err != nil {
			return IntegrationQueueResult{}, err
		}

		operationDescriptor, err := e.integrationResolver.ResolveOperation(provider, criteria)
		if err != nil {
			return IntegrationQueueResult{}, err
		}

		integrationRecord = ensuredRecord
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
		operationName = resolution.Operation.Name
		operationKind = resolution.Operation.Kind
	}

	operationConfig, err := decodeJSONObjectDocument(req.Config)
	if err != nil {
		return IntegrationQueueResult{}, err
	}

	if merged, err := operations.ResolveOperationConfig(&integrationRecord.Config, string(operationName), operationConfig); err != nil {
		return IntegrationQueueResult{}, err
	} else if merged != nil {
		operationConfig = merged
	}

	scopePayload, err := decodeJSONObjectDocument(req.ScopePayload)
	if err != nil {
		return IntegrationQueueResult{}, err
	}

	scopeAllowed, err := evaluateIntegrationScope(allowCtx, e.scopeEvaluator, req, integrationRecord, provider, operationName, operationConfig, scopePayload)
	if err != nil {
		return IntegrationQueueResult{}, err
	}
	if !scopeAllowed {
		return IntegrationQueueResult{}, ErrIntegrationScopeConditionFalse
	}

	if _, ensurePayloads, ok := ingest.BindingForOperation(operationName); ok && ensurePayloads {
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
	if operationConfig != nil {
		runBuilder.SetOperationConfig(operationConfig)
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
	receipt := workflows.EmitWorkflowEventWithHeaders(ctx, e.gala, integrations.IntegrationOperationRequestedTopic.Name, envelope, envelope.Headers())
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
		Config:          append(json.RawMessage(nil), params.Config...),
		ScopeExpression: params.ScopeExpression,
		ScopePayload:    append(json.RawMessage(nil), params.ScopePayload...),
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

	ingestFn, ensurePayloads, hasIngest := ingest.BindingForOperation(operationName)

	operationConfig := mapx.DeepCloneMapAny(run.OperationConfig)
	if ensurePayloads {
		operationConfig = operations.EnsureIncludePayloads(operationConfig)
	}

	var operationConfigDoc json.RawMessage
	if len(operationConfig) > 0 {
		if err := jsonx.RoundTrip(operationConfig, &operationConfigDoc); err != nil {
			return err
		}
	}

	operationCtx, cancel := integrationOperationContext(systemCtx, envelope.TimeoutSeconds)
	defer cancel()

	result, opErr := e.integrationOperations.Run(operationCtx, types.OperationRequest{
		OrgID:         run.OwnerID,
		IntegrationID: run.IntegrationID,
		Provider:      provider,
		Name:          operationName,
		Config:        operationConfigDoc,
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

	if runStatus == enums.IntegrationRunStatusSuccess && hasIngest && ingestFn != nil {
		var ingestErr error
		var ingestResult ingest.IngestResult

		envelopes, err := extractAlertEnvelopes(result.Details)
		if err != nil {
			ingestErr = err
		} else {
			ingestResult, ingestErr = ingestFn(operationCtx, ingest.IngestRequest{
				OrgID:             run.OwnerID,
				IntegrationID:     integrationRecord.ID,
				Provider:          provider,
				Operation:         operationName,
				IntegrationConfig: integrationRecord.Config,
				ProviderState:     integrationRecord.ProviderState,
				OperationConfig:   operationConfig,
				Envelopes:         envelopes,
				DB:                e.client,
			})
		}

		metricsDoc = appendIngestMetrics(metricsDoc, ingestResult)
		if ingestErr != nil {
			runStatus = enums.IntegrationRunStatusFailed
			errorText = ingestErr.Error()
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
		e.emitWorkflowActionCompleted(ctx.Context, envelope, runStatus, errorText)
	}

	if runStatus != enums.IntegrationRunStatusSuccess {
		return ErrIntegrationOperationFailed
	}

	return nil
}

// evaluateIntegrationScope evaluates optional scope expressions before queueing integration runs
func evaluateIntegrationScope(ctx context.Context, evaluator integrationscope.ConditionEvaluator, req IntegrationQueueRequest, integrationRecord *ent.Integration, provider types.ProviderType, operationName types.OperationName, operationConfig map[string]any, scopePayload map[string]any) (bool, error) {
	if evaluator == nil || req.ScopeExpression == "" {
		return true, nil
	}

	integrationConfig, err := jsonx.ToMap(integrationRecord.Config)
	if err != nil {
		return false, err
	}

	providerState, err := jsonx.ToMap(integrationRecord.ProviderState)
	if err != nil {
		return false, err
	}

	return evaluator.EvaluateConditionWithVars(ctx, req.ScopeExpression, integrationscope.ScopeVars{
		Payload:           mapx.DeepCloneMapAny(scopePayload),
		Resource:          req.ScopeResource,
		Provider:          provider,
		Operation:         operationName,
		Config:            mapx.DeepCloneMapAny(operationConfig),
		IntegrationConfig: integrationConfig,
		ProviderState:     providerState,
		OrgID:             req.OrgID,
		IntegrationID:     integrationRecord.ID,
	})
}

// decodeJSONObjectDocument decodes a raw JSON document into a map payload when provided.
func decodeJSONObjectDocument(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	decoded, err := jsonx.ToMap(raw)
	if err != nil {
		return nil, err
	}

	if len(decoded) == 0 {
		return nil, nil
	}

	return decoded, nil
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

// extractAlertEnvelopes pulls alert envelopes from operation details
func extractAlertEnvelopes(details json.RawMessage) ([]types.AlertEnvelope, error) {
	if len(details) == 0 {
		return nil, ErrIntegrationAlertPayloadsMissing
	}

	var payload struct {
		Alerts json.RawMessage `json:"alerts"`
	}
	if err := jsonx.RoundTrip(details, &payload); err != nil {
		return nil, err
	}
	if len(payload.Alerts) == 0 {
		return nil, ErrIntegrationAlertPayloadsMissing
	}

	var envelopes []types.AlertEnvelope
	if err := jsonx.RoundTrip(payload.Alerts, &envelopes); err == nil {
		return envelopes, nil
	}

	var envelope types.AlertEnvelope
	if err := jsonx.RoundTrip(payload.Alerts, &envelope); err != nil {
		return nil, err
	}

	return []types.AlertEnvelope{envelope}, nil
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
	// IngestSummary captures ingest summary metrics
	IngestSummary json.RawMessage `json:"ingest_summary,omitempty"`
	// IngestErrors captures ingest error metrics
	IngestErrors []string `json:"ingest_errors,omitempty"`
}

// encodeIntegrationRunOperationDetails normalizes operation details for metrics payloads
func encodeIntegrationRunOperationDetails(details json.RawMessage) json.RawMessage {
	if len(details) == 0 {
		return nil
	}

	return append(json.RawMessage(nil), details...)
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

// appendIngestMetrics attaches a unified ingest result to integration run metrics
func appendIngestMetrics(metrics integrationRunMetrics, result ingest.IngestResult) integrationRunMetrics {
	raw, err := json.Marshal(result.Summary)
	if err == nil {
		metrics.IngestSummary = raw
	}

	if len(result.Errors) > 0 {
		metrics.IngestErrors = append([]string(nil), result.Errors...)
	}

	return metrics
}
