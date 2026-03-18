package engine

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/ingest"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
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
	// Config carries the operation configuration payload
	Config map[string]any
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

// SetIntegrationDeps attaches integration dependencies and registers listeners when possible
func (e *WorkflowEngine) SetIntegrationDeps(deps IntegrationDeps) error {
	if deps.Registry != nil {
		e.integrationRegistry = deps.Registry
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
	if operationName == "" {
		return IntegrationQueueResult{}, ErrIntegrationOperationNameRequired
	}

	provider := req.Provider
	integrationID := req.IntegrationID

	allowCtx := workflows.AllowContext(ctx)
	var integrationRecord *ent.Integration
	var err error

	switch {
	case integrationID != "":
		integrationRecord, err = e.client.Integration.Query().
			Where(
				integration.IDEQ(integrationID),
				integration.OwnerIDEQ(orgID),
			).
			Only(allowCtx)
		if err != nil {
			return IntegrationQueueResult{}, err
		}
		if provider == types.ProviderUnknown {
			provider = types.ProviderTypeFromString(integrationRecord.Kind)
			if provider == types.ProviderUnknown {
				return IntegrationQueueResult{}, ErrIntegrationProviderUnknown
			}
		}
	default:
		if provider == types.ProviderUnknown {
			return IntegrationQueueResult{}, ErrIntegrationProviderRequired
		}
		if e.integrationStore == nil {
			return IntegrationQueueResult{}, ErrIntegrationStoreRequired
		}
		integrationRecord, err = e.integrationStore.EnsureIntegration(allowCtx, orgID, provider)
		if err != nil {
			return IntegrationQueueResult{}, err
		}
	}

	if e.integrationRegistry != nil && !operationDescriptorRegistered(e.integrationRegistry, provider, operationName) {
		return IntegrationQueueResult{}, keystore.ErrOperationNotRegistered
	}

	operationConfig := maps.Clone(req.Config)
	if merged, err := operations.ResolveOperationConfig(&integrationRecord.Config, string(operationName), req.Config); err != nil {
		return IntegrationQueueResult{}, err
	} else if merged != nil {
		operationConfig = merged
	}

	if operationName == types.OperationVulnerabilitiesCollect {
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
		RunID:       runRecord.ID,
		OrgID:       orgID,
		Provider:    string(provider),
		Operation:   string(operationName),
		Force:       req.Force,
		ClientForce: req.ClientForce,
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

	var params workflows.IntegrationActionParams
	if err := json.Unmarshal(action.Params, &params); err != nil {
		return errors.Join(ErrUnmarshalParams, err)
	}

	operationName := types.OperationName(params.Operation)
	if operationName == "" {
		return ErrIntegrationOperationNameRequired
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
		OrgID:         orgID,
		Provider:      types.ProviderTypeFromString(params.Provider),
		IntegrationID: params.Integration,
		Operation:     operationName,
		Config:        maps.Clone(params.Config),
		Force:         params.Force,
		ClientForce:   params.ClientForce,
		RunType:       enums.IntegrationRunTypeEvent,
		WorkflowMeta:  meta,
	})
	if err != nil {
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

	operationConfig := maps.Clone(run.OperationConfig)
	if operationName == types.OperationVulnerabilitiesCollect {
		operationConfig = operations.EnsureIncludePayloads(operationConfig)
	}

	operationCtx, cancel := integrationOperationContext(systemCtx, envelope.TimeoutSeconds)
	defer cancel()

	result, opErr := e.integrationOperations.Run(operationCtx, types.OperationRequest{
		OrgID:       run.OwnerID,
		Provider:    provider,
		Name:        operationName,
		Config:      operationConfig,
		Force:       payload.Force,
		ClientForce: payload.ClientForce,
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

	metrics := buildOperationMetrics(result)

	if runStatus == enums.IntegrationRunStatusSuccess && operationName == types.OperationVulnerabilitiesCollect {
		var ingestErr error
		var ingestResult ingest.VulnerabilityIngestResult

		if ingest.SupportsVulnerabilityIngest(provider, integrationRecord.Config) {
			envelopes, err := extractAlertEnvelopes(result.Details)
			if err != nil {
				ingestErr = err
			} else {
				ingestResult, ingestErr = ingest.VulnerabilityAlerts(operationCtx, ingest.VulnerabilityIngestRequest{
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
		} else {
			ingestErr = ingest.ErrMappingNotFound
		}

		metrics = appendIngestMetrics(metrics, ingestResult)
		if ingestErr != nil {
			runStatus = enums.IntegrationRunStatusFailed
			errorText = ingestErr.Error()
		}
	}

	finishedAt := time.Now()
	durationMs := int(finishedAt.Sub(startedAt).Milliseconds())
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

// operationDescriptorRegistered reports whether a provider operation is registered
func operationDescriptorRegistered(reg IntegrationRegistry, provider types.ProviderType, name types.OperationName) bool {
	if reg == nil {
		return true
	}

	return lo.ContainsBy(reg.OperationDescriptors(provider), func(descriptor types.OperationDescriptor) bool {
		return descriptor.Name == name
	})
}

// extractAlertEnvelopes pulls alert envelopes from operation details
func extractAlertEnvelopes(details map[string]any) ([]types.AlertEnvelope, error) {
	if details == nil {
		return nil, ErrIntegrationAlertPayloadsMissing
	}

	raw, ok := details["alerts"]
	if !ok || raw == nil {
		return nil, ErrIntegrationAlertPayloadsMissing
	}

	switch value := raw.(type) {
	case []types.AlertEnvelope:
		return value, nil
	case []any:
		envelopes := make([]types.AlertEnvelope, 0, len(value))
		for _, item := range value {
			if typed, ok := item.(types.AlertEnvelope); ok {
				envelopes = append(envelopes, typed)
				continue
			}

			envelope, err := decodeAlertEnvelope(item)
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}

		return envelopes, nil
	default:
		envelope, err := decodeAlertEnvelope(value)
		if err != nil {
			return nil, err
		}

		return []types.AlertEnvelope{envelope}, nil
	}
}

// decodeAlertEnvelope coerces an envelope from a dynamic payload
func decodeAlertEnvelope(value any) (types.AlertEnvelope, error) {
	var envelope types.AlertEnvelope
	if err := jsonx.RoundTrip(value, &envelope); err != nil {
		return envelope, err
	}

	return envelope, nil
}

// buildOperationMetrics builds a metrics map for operation output
func buildOperationMetrics(result types.OperationResult) map[string]any {
	operation := map[string]any{
		"status":  string(result.Status),
		"summary": result.Summary,
	}
	metrics := map[string]any{
		"operation": operation,
	}
	if result.Details != nil {
		operation["details"] = result.Details
	}

	return metrics
}

// appendIngestMetrics attaches ingest results to metrics
func appendIngestMetrics(metrics map[string]any, result ingest.VulnerabilityIngestResult) map[string]any {
	if metrics == nil {
		metrics = map[string]any{}
	}
	metrics["ingest_summary"] = result.Summary
	if len(result.Errors) > 0 {
		metrics["ingest_errors"] = result.Errors
	}

	return metrics
}
