package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/operations"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// integrationQueuedFlag is a mutable flag set when an integration operation is successfully
// queued, signaling callers without using an error return value.
type integrationQueuedFlag struct{ queued bool }

var integrationQueuedKey = contextx.NewKey[*integrationQueuedFlag]()

// withIntegrationQueuedFlag installs a mutable queued flag into the context.
func withIntegrationQueuedFlag(ctx context.Context) context.Context {
	return integrationQueuedKey.Set(ctx, &integrationQueuedFlag{})
}

// markIntegrationQueued signals that an integration operation was queued.
func markIntegrationQueued(ctx context.Context) {
	if flag, ok := integrationQueuedKey.Get(ctx); ok && flag != nil {
		flag.queued = true
	}
}

// isIntegrationQueued reports whether an integration operation was queued.
func isIntegrationQueued(ctx context.Context) bool {
	flag, ok := integrationQueuedKey.Get(ctx)
	return ok && flag != nil && flag.queued
}

// IntegrationDeps wires integration-specific dependencies into the workflow engine
type IntegrationDeps struct {
	// Runtime provides access to integration definition descriptors and execution.
	Runtime *integrationsruntime.Runtime
}

// IntegrationQueueRequest describes a queued integration operation
type IntegrationQueueRequest struct {
	// OrgID identifies the organization requesting the operation
	OrgID string
	// InstallationID is the explicit installation identifier for the operation
	InstallationID string
	// DefinitionID identifies the integration definition when no installation ID is set
	DefinitionID string
	// Operation identifies the operation to execute
	Operation string
	// Config carries the operation configuration payload as a JSON object document
	Config json.RawMessage
	// ScopeExpression is an optional CEL expression gate for command execution
	ScopeExpression string
	// ScopePayload is optional data exposed to scope expression evaluation as a JSON object document
	ScopePayload json.RawMessage
	// ScopeResource is optional resource identity exposed to scope expression evaluation
	ScopeResource string
	// ForceClientRebuild requests client cache bypass
	ForceClientRebuild bool
	// RunType identifies the integration run type
	RunType enums.IntegrationRunType
	// WorkflowMeta links the operation to a workflow action
	WorkflowMeta *operations.WorkflowMeta
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

// integrationOpContext captures common integration operation log fields
type integrationOpContext struct {
	definitionID   string
	operation      string
	installationID string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler for integrationOpContext
func (c integrationOpContext) MarshalZerologObject(e *zerolog.Event) {
	e.Str("definition_id", c.definitionID).Str("operation", c.operation)
	if c.installationID != "" {
		e.Str("installation_id", c.installationID)
	}
}

// logIntegrationScopeSkipped logs a debug event when an integration action is skipped by scope evaluation
func logIntegrationScopeSkipped(ctx context.Context, definitionID, operation, installationID, scopeExpression string) {
	logx.FromContext(ctx).Debug().
		EmbedObject(integrationOpContext{definitionID: definitionID, operation: operation, installationID: installationID}).
		Str("scope_expression", scopeExpression).
		Msg("integration action skipped by scope condition")
}

// SetIntegrationDeps attaches integration dependencies and registers per-operation listeners
func (e *WorkflowEngine) SetIntegrationDeps(deps IntegrationDeps) error {
	if deps.Runtime != nil {
		e.integrationRuntime = deps.Runtime

		evaluator, err := NewIntegrationScopeEvaluator()
		if err != nil {
			return err
		}

		e.scopeEvaluator = evaluator
	}

	if e.integrationRuntime == nil {
		return nil
	}

	if e.gala == nil {
		return nil
	}

	// Register one listener per operation topic; each listener executes the operation
	// via the executor and emits workflow action completed when workflow meta is present.
	// sync.Once ensures idempotent registration even if SetIntegrationDeps is called concurrently.
	e.integrationListenersOnce.Do(func() {
		for _, op := range e.integrationRuntime.Registry().Listeners() {
			operation := op // capture for closure
			if _, err := gala.RegisterListeners(e.gala.Registry(), gala.Definition[operations.Envelope]{
				Topic: gala.Topic[operations.Envelope]{Name: operation.Topic},
				Name:  fmt.Sprintf("engine.integrations.%s", operation.Topic),
				Handle: func(ctx gala.HandlerContext, envelope operations.Envelope) error {
					return e.handleIntegrationOperation(ctx, envelope)
				},
			}); err != nil {
				e.integrationListenersErr = err
				return
			}
		}
	})

	return e.integrationListenersErr
}

// QueueIntegrationOperation resolves the installation and dispatches the operation
func (e *WorkflowEngine) QueueIntegrationOperation(ctx context.Context, req IntegrationQueueRequest) (IntegrationQueueResult, error) {
	if e.integrationRuntime == nil {
		return IntegrationQueueResult{}, ErrIntegrationOperationsRequired
	}

	orgID := req.OrgID
	if orgID == "" {
		return IntegrationQueueResult{}, ErrIntegrationOwnerRequired
	}

	if req.Operation == "" {
		return IntegrationQueueResult{}, ErrIntegrationOperationCriteriaRequired
	}

	allowCtx := workflows.AllowContext(ctx)

	installationRecord, err := e.integrationRuntime.ResolveInstallation(allowCtx, orgID, req.InstallationID, req.DefinitionID)
	if err != nil {
		return IntegrationQueueResult{}, err
	}

	scopeAllowed, err := evaluateInstallationScope(allowCtx, e.scopeEvaluator, req, installationRecord, req.Operation, req.Config)
	if err != nil {
		return IntegrationQueueResult{}, err
	}
	if !scopeAllowed {
		return IntegrationQueueResult{}, ErrIntegrationScopeConditionFalse
	}

	runType := req.RunType
	if runType == "" {
		runType = enums.IntegrationRunTypeEvent
	}

	result, err := e.integrationRuntime.Dispatch(allowCtx, operations.DispatchRequest{
		InstallationID:     installationRecord.ID,
		Operation:          req.Operation,
		Config:             jsonx.CloneRawMessage(req.Config),
		ForceClientRebuild: req.ForceClientRebuild,
		RunType:            runType,
		WorkflowMeta:       req.WorkflowMeta,
	})
	if err != nil {
		return IntegrationQueueResult{}, err
	}

	return IntegrationQueueResult{
		RunID:   result.RunID,
		EventID: result.EventID,
		Status:  result.Status,
	}, nil
}

// executeIntegrationAction queues an integration operation from a workflow action
func (e *WorkflowEngine) executeIntegrationAction(ctx context.Context, action models.WorkflowAction, instance *ent.WorkflowInstance, obj *workflows.Object) error {
	if e.integrationRuntime == nil {
		return ErrIntegrationOperationsRequired
	}

	var params workflows.IntegrationActionParams
	if err := jsonx.RoundTrip(action.Params, &params); err != nil {
		return errors.Join(ErrUnmarshalParams, err)
	}

	operationName := params.OperationName
	if operationName == "" {
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

	meta := &operations.WorkflowMeta{
		InstanceID:  instance.ID,
		ActionKey:   action.Key,
		ActionIndex: actionIndexForKey(instance.DefinitionSnapshot.Actions, action.Key),
	}
	if obj != nil {
		meta.ObjectID = obj.ID
		meta.ObjectType = obj.Type
	}

	_, err := e.QueueIntegrationOperation(ctx, IntegrationQueueRequest{
		OrgID:              orgID,
		InstallationID:     params.InstallationID,
		DefinitionID:       params.DefinitionID,
		Operation:          operationName,
		Config:             jsonx.CloneRawMessage(params.Config),
		ScopeExpression:    params.ScopeExpression,
		ScopePayload:       jsonx.CloneRawMessage(params.ScopePayload),
		ScopeResource:      params.ScopeResource,
		ForceClientRebuild: params.ForceClientRebuild,
		RunType:            enums.IntegrationRunTypeEvent,
		WorkflowMeta:       meta,
	})
	if err != nil {
		if errors.Is(err, ErrIntegrationScopeConditionFalse) {
			logIntegrationScopeSkipped(ctx, params.DefinitionID, string(operationName), params.InstallationID, params.ScopeExpression)
			return nil
		}
		return err
	}

	markIntegrationQueued(ctx)
	return nil
}

// handleIntegrationOperation executes an integration operation and emits workflow action completed when applicable
func (e *WorkflowEngine) handleIntegrationOperation(ctx gala.HandlerContext, envelope operations.Envelope) error {
	if e.integrationRuntime == nil {
		return ErrIntegrationOperationsRequired
	}

	systemCtx := workflows.AllowContext(ctx.Context)

	execErr := e.integrationRuntime.HandleOperation(systemCtx, envelope)

	if envelope.WorkflowMeta != nil {
		e.emitWorkflowActionCompleted(systemCtx, envelope, execErr)
	}

	return execErr
}

// emitWorkflowActionCompleted emits a completion event after an integration run finishes
func (e *WorkflowEngine) emitWorkflowActionCompleted(ctx context.Context, envelope operations.Envelope, execErr error) {
	if e.gala == nil || envelope.WorkflowMeta == nil {
		return
	}

	meta := envelope.WorkflowMeta

	actionPayload := gala.WorkflowActionCompletedPayload{
		InstanceID:  meta.InstanceID,
		ActionIndex: meta.ActionIndex,
		ActionType:  enums.WorkflowActionTypeIntegration,
		ObjectID:    meta.ObjectID,
		ObjectType:  meta.ObjectType,
		Success:     execErr == nil,
		Skipped:     false,
	}
	if execErr != nil {
		actionPayload.ErrorMessage = execErr.Error()
	}

	receipt := workflows.EmitWorkflowEvent(ctx, e.gala, gala.TopicWorkflowActionCompleted, actionPayload)
	if receipt.Err != nil {
		logx.FromContext(ctx).Warn().Err(receipt.Err).Msg("failed to emit workflow action completed for integration run")
	}
}

// evaluateInstallationScope evaluates optional scope expressions before queueing integration runs
func evaluateInstallationScope(ctx context.Context, evaluator *IntegrationScopeEvaluator, req IntegrationQueueRequest, installationRecord *ent.Integration, operationName string, operationConfig json.RawMessage) (bool, error) {
	if evaluator == nil || req.ScopeExpression == "" {
		return true, nil
	}

	return evaluator.EvaluateConditionWithVars(ctx, req.ScopeExpression, types.ScopeVars{
		Payload:            req.ScopePayload,
		Resource:           req.ScopeResource,
		Definition:         installationRecord.DefinitionID,
		Operation:          operationName,
		Config:             operationConfig,
		InstallationConfig: jsonx.CloneRawMessage(installationRecord.Config.ClientConfig),
		OrgID:              req.OrgID,
		InstallationID:     installationRecord.ID,
	})
}
