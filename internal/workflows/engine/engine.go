package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/iam/auth"
)

// WorkflowEngine orchestrates workflow execution via event emission
type WorkflowEngine struct {
	// client is the ent database client
	client *generated.Client
	// emitter is the event emitter for workflow events
	emitter soiree.Emitter
	// observer is the observability observer for metrics and tracing
	observer *observability.Observer
	// config is the workflow configuration
	config *workflows.Config
	// env is the CEL environment for expression evaluation
	env *cel.Env
	// celEvaluator handles CEL expression compilation and evaluation
	celEvaluator *CELEvaluator
	// proposalManager handles workflow proposal operations
	proposalManager *ProposalManager
}

// NewWorkflowEngine creates a new workflow engine using the provided configuration options
func NewWorkflowEngine(client *generated.Client, emitter soiree.Emitter, opts ...workflows.ConfigOpts) (*WorkflowEngine, error) {
	config := workflows.NewDefaultConfig(opts...)

	return NewWorkflowEngineWithConfig(client, emitter, config)
}

// NewWorkflowEngineWithConfig creates a new workflow engine using the provided configuration
func NewWorkflowEngineWithConfig(client *generated.Client, emitter soiree.Emitter, config *workflows.Config) (*WorkflowEngine, error) {
	if client == nil {
		return nil, ErrNilClient
	}

	if config == nil {
		config = workflows.NewDefaultConfig()
	}

	env, err := workflows.NewCELEnv(config, workflows.CELScopeAction)
	if err != nil {
		return nil, err
	}

	celEvaluator := NewCELEvaluator(env, config)
	proposalManager := NewProposalManager(client)

	return &WorkflowEngine{
		client:          client,
		emitter:         emitter,
		observer:        observability.New(),
		config:          config,
		env:             env,
		celEvaluator:    celEvaluator,
		proposalManager: proposalManager,
	}, nil
}

// TriggerWorkflow starts a new workflow instance
func (e *WorkflowEngine) TriggerWorkflow(ctx context.Context, def *generated.WorkflowDefinition, obj *workflows.Object, input TriggerInput) (instance *generated.WorkflowInstance, err error) {
	scope := observability.BeginEngine(ctx, e.observer, observability.OpTriggerWorkflow, input.EventType, lo.Assign(observability.Fields(obj.ObservabilityFields()), observability.Fields{
		workflowinstance.FieldWorkflowDefinitionID: def.ID,
	}))
	ctx = scope.Context()
	defer scope.End(err, nil)

	shouldRun, err := e.EvaluateConditions(ctx, def, obj, input.EventType, input.ChangedFields, input.ChangedEdges, input.AddedIDs, input.RemovedIDs, input.ProposedChanges)
	if err != nil {
		return nil, scope.Fail(fmt.Errorf("failed to evaluate conditions: %w", err), nil)
	}

	if !shouldRun {
		scope.Skip("conditions_not_met", nil)
		return nil, nil
	}

	// Guard against multiple active instances per {object, definition}
	domain, err := approvalDomainForTrigger(def, input.ProposedChanges, input.ChangedFields)
	if err != nil {
		return nil, scope.Fail(err, nil)
	}
	if domain != nil && domain.DomainKey != "" {
		scope.WithFields(observability.Fields{
			workflowproposal.FieldDomainKey: domain.DomainKey,
		})
	}
	if err := e.guardTrigger(ctx, def, obj, domain); err != nil {
		return nil, scope.Fail(err, nil)
	}

	defSnapshot := e.serializeDefinition(def)

	userID, _ := auth.GetSubjectIDFromContext(ctx)
	contextData := buildTriggerContext(def.ID, obj, input, userID)

	// Wrap instance + object ref creation in transaction to prevent stranded instances
	// Use privacy bypass for internal workflow operations
	allowCtx, ownerID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		return nil, scope.Fail(err, nil)
	}
	instance, err = e.createInstanceTx(allowCtx, def, obj, domain, defSnapshot, contextData, ownerID, scope)
	if err != nil {
		return nil, scope.Fail(err, nil)
	}

	// Emit workflow triggered event AFTER transaction commits
	e.emitWorkflowTriggered(scope.Context(), observability.OpTriggerWorkflow, input.EventType, instance, def.ID, obj, input.ChangedFields)

	return instance, nil
}

// TriggerExistingInstance resumes a pre-created workflow instance and emits a trigger event
func (e *WorkflowEngine) TriggerExistingInstance(ctx context.Context, instance *generated.WorkflowInstance, def *generated.WorkflowDefinition, obj *workflows.Object, input TriggerInput) (err error) {
	scope := observability.BeginEngine(ctx, e.observer, observability.OpTriggerExistingInstance, input.EventType, lo.Assign(observability.Fields(obj.ObservabilityFields()), observability.Fields{
		workflowassignment.FieldWorkflowInstanceID: instance.ID,
		workflowinstance.FieldWorkflowDefinitionID: def.ID,
	}))
	ctx = scope.Context()
	defer scope.End(err, nil)

	if instance.State == enums.WorkflowInstanceStateCompleted || instance.State == enums.WorkflowInstanceStateFailed {
		return scope.Fail(ErrInvalidState, observability.Fields{
			workflowinstance.FieldState: instance.State.String(),
		})
	}

	userID, _ := auth.GetSubjectIDFromContext(ctx)
	contextData := applyTriggerContext(instance.Context, def.ID, obj, input, userID)

	allowCtx := workflows.AllowContext(ctx)
	if err := e.client.WorkflowInstance.UpdateOneID(instance.ID).
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStateRunning).
		SetDefinitionSnapshot(e.serializeDefinition(def)).
		SetContext(contextData).
		SetCurrentActionIndex(0).
		Exec(allowCtx); err != nil {
		return scope.Fail(fmt.Errorf("failed to resume workflow instance: %w", err), nil)
	}

	e.emitWorkflowTriggered(scope.Context(), observability.OpTriggerExistingInstance, input.EventType, instance, def.ID, obj, input.ChangedFields)

	return nil
}

// guardTrigger enforces active-instance checks for a trigger attempt
func (e *WorkflowEngine) guardTrigger(ctx context.Context, def *generated.WorkflowDefinition, obj *workflows.Object, domain *workflows.DomainChanges) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	// For PRE_COMMIT approval workflows, guard per (object, domain) to allow multiple concurrent instances
	// for different approval domains on the same object. POST_COMMIT approvals behave like reviews.
	if workflows.DefinitionUsesPreCommitApprovals(def.DefinitionJSON) && domain != nil && len(domain.Fields) > 0 {
		return e.guardTriggerPerDomain(ctx, def, obj, domain.Fields)
	}

	// For non-approval workflows, guard per (object, definition)
	instanceQuery := e.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.StateIn(enums.WorkflowInstanceStateRunning, enums.WorkflowInstanceStatePaused),
			workflowinstance.OwnerIDEQ(orgID),
		)

	instanceQuery = generated.ApplyWorkflowInstanceObjectPredicate(instanceQuery, obj.Type, obj.ID)

	if exists, err := instanceQuery.Exist(ctx); err != nil {
		return err
	} else if exists {
		return workflows.ErrWorkflowAlreadyActive
	}

	if def.CooldownSeconds <= 0 {
		return nil
	}

	cutoff := time.Now().Add(-time.Duration(def.CooldownSeconds) * time.Second)
	cooldownQuery := e.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.CreatedAtGTE(cutoff),
			workflowinstance.OwnerIDEQ(orgID),
		)

	cooldownQuery = generated.ApplyWorkflowInstanceObjectPredicate(cooldownQuery, obj.Type, obj.ID)

	if exists, err := cooldownQuery.Exist(ctx); err != nil {
		return err
	} else if exists {
		return workflows.ErrWorkflowAlreadyActive
	}

	return nil
}

// guardTriggerPerDomain checks for active instances per (object, approval domain)
func (e *WorkflowEngine) guardTriggerPerDomain(ctx context.Context, def *generated.WorkflowDefinition, obj *workflows.Object, approvalFields []string) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Derive domain key from approval fields
	domainKey := workflows.DeriveDomainKey(obj.Type, approvalFields)

	objRefIDs, err := workflows.ObjectRefIDs(ctx, e.client, obj)
	if err != nil {
		return fmt.Errorf("failed to query object refs: %w", err)
	}
	if len(objRefIDs) == 0 {
		return nil
	}

	// Find proposals for this domain
	proposals, err := e.client.WorkflowProposal.Query().
		Where(
			workflowproposal.WorkflowObjectRefIDIn(objRefIDs...),
			workflowproposal.DomainKeyEQ(domainKey),
			workflowproposal.OwnerIDEQ(orgID),
		).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to query proposals: %w", err)
	}

	// If no proposals exist for this domain, no guard needed
	if len(proposals) == 0 {
		return nil
	}

	// Check for active instances linked to any proposal in this domain
	for _, proposalID := range proposals {
		exists, err := e.client.WorkflowInstance.Query().
			Where(
				workflowinstance.WorkflowProposalIDEQ(proposalID),
				workflowinstance.StateIn(enums.WorkflowInstanceStateRunning, enums.WorkflowInstanceStatePaused),
				workflowinstance.OwnerIDEQ(orgID),
			).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check for active instances: %w", err)
		}
		if exists {
			return workflows.ErrWorkflowAlreadyActive
		}
	}

	// Check cooldown if configured
	if def.CooldownSeconds > 0 {
		cutoff := time.Now().Add(-time.Duration(def.CooldownSeconds) * time.Second)
		for _, proposalID := range proposals {
			exists, err := e.client.WorkflowInstance.Query().
				Where(
					workflowinstance.WorkflowProposalIDEQ(proposalID),
					workflowinstance.CreatedAtGTE(cutoff),
					workflowinstance.OwnerIDEQ(orgID),
				).
				Exist(ctx)
			if err != nil {
				return fmt.Errorf("failed to check cooldown: %w", err)
			}
			if exists {
				return workflows.ErrWorkflowAlreadyActive
			}
		}
	}

	return nil
}

// ProcessAction executes a workflow action
func (e *WorkflowEngine) ProcessAction(ctx context.Context, instance *generated.WorkflowInstance, action models.WorkflowAction) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Use allow context for internal workflow operations
	allowCtx := workflows.AllowContext(ctx)

	objRef, err := e.client.WorkflowObjectRef.
		Query().
		Where(
			workflowobjectref.WorkflowInstanceIDEQ(instance.ID),
			workflowobjectref.OwnerIDEQ(orgID),
		).
		First(allowCtx)

	if err != nil {
		return fmt.Errorf("failed to load object reference: %w", err)
	}

	obj, err := workflows.ObjectFromRef(objRef)
	if err != nil {
		return err
	}

	actionType := enums.ToWorkflowActionType(action.Type)

	if err := e.Execute(ctx, action, instance, obj); err != nil {
		return err
	}

	// Approval actions pause the workflow until decisions arrive
	if actionType != nil && isGatedActionType(*actionType) {
		allowCtx := workflows.AllowContext(ctx)
		if err := e.client.WorkflowInstance.UpdateOneID(instance.ID).
			SetState(enums.WorkflowInstanceStatePaused).
			Exec(allowCtx); err != nil {
			return err
		}
	}

	return nil
}

// CompleteAssignment marks an assignment as approved/rejected
func (e *WorkflowEngine) CompleteAssignment(ctx context.Context, assignmentID string, status enums.WorkflowAssignmentStatus, approvalMetadata *models.WorkflowAssignmentApproval, rejectionMetadata *models.WorkflowAssignmentRejection) (err error) {
	scope := observability.BeginEngine(ctx, e.observer, observability.OpCompleteAssignment, status.String(), observability.Fields{
		workflowassignmenttarget.FieldWorkflowAssignmentID: assignmentID,
		workflowassignment.FieldStatus:                     status.String(),
	})
	ctx = scope.Context()
	defer scope.End(err, nil)

	// Use privacy bypass for internal workflow operations
	allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		return scope.Fail(err, nil)
	}

	assignment, err := e.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.IDEQ(assignmentID),
			workflowassignment.OwnerIDEQ(orgID),
		).
		Only(allowCtx)
	if err != nil {
		return scope.Fail(fmt.Errorf("%w: %v", ErrAssignmentNotFound, err), nil)
	}

	scope.WithFields(observability.Fields{
		workflowassignment.FieldWorkflowInstanceID: assignment.WorkflowInstanceID,
	})

	update := assignment.Update().SetStatus(status)
	switch status {
	case enums.WorkflowAssignmentStatusApproved:
		if approvalMetadata != nil {
			update.SetApprovalMetadata(*approvalMetadata)
		}
	case enums.WorkflowAssignmentStatusRejected, enums.WorkflowAssignmentStatusChangesRequested:
		if rejectionMetadata != nil {
			update.SetRejectionMetadata(*rejectionMetadata)
		}
	}

	if err = update.Exec(allowCtx); err != nil {
		return scope.Fail(fmt.Errorf("%w: %w", ErrAssignmentUpdateFailed, err), nil)
	}

	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return scope.Fail(fmt.Errorf("failed to get subject ID from context: %w", err), nil)
	}

	payload := soiree.WorkflowAssignmentCompletedPayload{
		AssignmentID: assignmentID,
		InstanceID:   assignment.WorkflowInstanceID,
		Status:       status,
		CompletedBy:  userID,
	}

	allowCtx = workflows.AllowContext(ctx)

	instance, instanceErr := loadWorkflowInstance(allowCtx, e.client, assignment.WorkflowInstanceID, orgID)
	if instanceErr != nil {
		observability.WarnEngine(scope.Context(), observability.OpCompleteAssignment, status.String(), observability.Fields{
			workflowassignment.FieldWorkflowInstanceID: assignment.WorkflowInstanceID,
		}, instanceErr)
		instance = &generated.WorkflowInstance{
			ID:      assignment.WorkflowInstanceID,
			OwnerID: orgID,
		}
	}

	meta := workflows.EmitFailureMeta{
		EventType:   enums.WorkflowEventTypeAssignmentResolved,
		ActionKey:   assignment.ApprovalMetadata.ActionKey,
		ActionIndex: -1,
	}
	if instance != nil {
		meta.ObjectID = instance.Context.ObjectID
		meta.ObjectType = instance.Context.ObjectType
	}

	emitEngineEvent(allowCtx, e, observability.OpCompleteAssignment, status.String(), instance, meta, soiree.WorkflowAssignmentCompletedTopic, payload, observability.Fields{
		workflowassignment.FieldWorkflowInstanceID: assignment.WorkflowInstanceID,
	})

	return nil
}

// serializeDefinition converts a workflow definition to the storage format
func (e *WorkflowEngine) serializeDefinition(def *generated.WorkflowDefinition) models.WorkflowDefinitionDocument {
	doc := def.DefinitionJSON
	if workflows.DefinitionUsesPostCommitApprovals(doc) {
		doc = workflows.ConvertApprovalActionsToReview(doc)
	}

	return doc
}

// approvalDomainForTrigger picks the first matching approval domain for a trigger event
func approvalDomainForTrigger(def *generated.WorkflowDefinition, proposedChanges map[string]any, changedFields []string) (*workflows.DomainChanges, error) {
	objectType := enums.ToWorkflowObjectType(def.SchemaType)
	if objectType == nil {
		return nil, workflows.ErrUnsupportedObjectType
	}

	if len(proposedChanges) > 0 {
		matches, err := workflows.DomainChangesForDefinition(def.DefinitionJSON, *objectType, proposedChanges)
		if err != nil {
			return nil, err
		}

		if len(matches) > 0 {
			return &matches[0], nil
		}
	}

	domains, err := workflows.ApprovalDomains(def.DefinitionJSON)
	if err != nil {
		return nil, err
	}

	if len(changedFields) > 0 {
		changes := lo.SliceToMap(changedFields, func(field string) (string, any) {
			return field, true
		})

		matches := workflows.SplitChangesByDomains(changes, *objectType, domains)
		if len(matches) > 0 {
			return &workflows.DomainChanges{
				DomainKey: matches[0].DomainKey,
				Fields:    matches[0].Fields,
			}, nil
		}
	}

	return nil, nil
}
