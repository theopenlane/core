package engine

import (
	"context"
	"fmt"
	"maps"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/iam/auth"
)

// stringArrayContains creates a predicate that checks if a JSON string array field contains a value
// Returns false for NULL fields or non-array types
func stringArrayContains(field string, value string) predicate.WorkflowDefinition {
	return func(s *sql.Selector) {
		// Only check containment if field is not null and is an array
		s.Where(sql.And(
			sql.NotNull(field),
			sql.P(func(b *sql.Builder) {
				b.WriteString("jsonb_typeof(").Ident(field).WriteString(") = 'array'")
			}),
			sqljson.ValueContains(field, []string{value}),
		))
	}
}

// stringArrayEmpty creates a predicate that checks if a JSON string array field is null or empty
func stringArrayEmpty(field string) predicate.WorkflowDefinition {
	return func(s *sql.Selector) {
		// Use CASE to safely check array length only when field is actually an array
		// This handles NULL, non-array types, and empty arrays
		s.Where(sql.P(func(b *sql.Builder) {
			b.WriteString("CASE WHEN ").Ident(field).WriteString(" IS NULL THEN true ")
			b.WriteString("WHEN jsonb_typeof(").Ident(field).WriteString(") != 'array' THEN true ")
			b.WriteString("ELSE jsonb_array_length(").Ident(field).WriteString(") = 0 END")
		}))
	}
}

// EvaluateConditions checks if all conditions pass for a workflow.
func (e *WorkflowEngine) EvaluateConditions(ctx context.Context, def *generated.WorkflowDefinition, obj *workflows.Object, eventType string, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string, proposedChanges map[string]any) (bool, error) {
	conditions := def.DefinitionJSON.Conditions
	if len(conditions) == 0 {
		return true, nil
	}

	userID, _ := auth.GetSubjectIDFromContext(ctx)

	vars := workflows.BuildCELVars(obj, changedFields, changedEdges, addedIDs, removedIDs, eventType, userID, proposedChanges)

	for i, cond := range conditions {
		result, err := e.celEvaluator.Evaluate(ctx, cond.Expression, vars)
		if err != nil {
			return false, fmt.Errorf("%w: condition %d: %v", ErrConditionFailed, i, err)
		}

		if !result {
			return false, nil
		}
	}

	return true, nil
}

// EvaluateActionWhen evaluates an action's When expression with assignment context.
// This is used for re-evaluating NOTIFY actions when assignment status changes.
func (e *WorkflowEngine) EvaluateActionWhen(ctx context.Context, expression string, instance *generated.WorkflowInstance, obj *workflows.Object) (bool, error) {
	vars, err := e.buildActionCELVars(ctx, instance, obj)
	if err != nil {
		return false, err
	}

	return e.celEvaluator.Evaluate(ctx, expression, vars)
}

// buildActionCELVars assembles CEL variables for action evaluation (including assignment context).
func (e *WorkflowEngine) buildActionCELVars(ctx context.Context, instance *generated.WorkflowInstance, obj *workflows.Object) (map[string]any, error) {
	if instance == nil {
		return nil, ErrInstanceNotFound
	}

	// Use proposed changes from instance context (set when workflow was triggered)
	proposedChanges := instance.Context.TriggerProposedChanges

	// Ensure the object node is loaded so CEL has access to concrete fields.
	if obj != nil && obj.Node == nil {
		allowCtx := workflows.AllowContext(ctx)
		if _, err := e.loadObjectNode(allowCtx, obj); err != nil {
			return nil, err
		}
	}

	// Fallback to loading from proposal if instance context doesn't have proposed changes
	if len(proposedChanges) == 0 && instance.WorkflowProposalID != "" {
		allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
		if err != nil {
			return nil, err
		}

		proposal, err := e.client.WorkflowProposal.Query().
			Where(
				workflowproposal.IDEQ(instance.WorkflowProposalID),
				workflowproposal.OwnerIDEQ(orgID),
			).
			Only(allowCtx)
		if err == nil && proposal != nil {
			proposedChanges = proposal.Changes
		}
	}

	vars := workflows.BuildCELVars(
		obj,
		instance.Context.TriggerChangedFields,
		instance.Context.TriggerChangedEdges,
		instance.Context.TriggerAddedIDs,
		instance.Context.TriggerRemovedIDs,
		instance.Context.TriggerEventType,
		instance.Context.TriggerUserID,
		proposedChanges,
	)

	// Merge assignment context (assignments, instance, initiator)
	// Use privacy bypass for internal workflow operations that query assignment state
	allowCtx := workflows.AllowContext(ctx)
	assignmentCtx, err := workflows.BuildAssignmentContext(allowCtx, e.client, instance.ID)
	if err != nil {
		return nil, err
	}

	if assignmentCtx != nil {
		maps.Copy(vars, assignmentCtx)
	} else {
		// Provide empty defaults so CEL expressions don't fail on missing variables
		vars["assignments"] = map[string]any{}
		vars["instance"] = map[string]any{}
		vars["initiator"] = ""
	}

	return vars, nil
}

// FindMatchingDefinitions returns all active workflow definitions that match the criteria
func (e *WorkflowEngine) FindMatchingDefinitions(ctx context.Context, schemaType string, eventType string, changedFields []string, changedEdges []string, addedIDs map[string][]string, removedIDs map[string][]string, proposedChanges map[string]any, obj *workflows.Object) (defs []*generated.WorkflowDefinition, err error) {
	scope := observability.BeginEngine(ctx, e.observer, observability.OpFindMatchingDefinitions, eventType, lo.Assign(observability.Fields(obj.ObservabilityFields()), observability.Fields{
		workflowevent.FieldEventType: eventType,
	}))
	ctx = scope.Context()
	defer scope.End(err, nil)

	// Use privacy bypass for internal workflow operations
	allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		return nil, scope.Fail(err, nil)
	}

	query := e.client.WorkflowDefinition.
		Query().
		Where(
			workflowdefinition.SchemaTypeEQ(schemaType),
			workflowdefinition.ActiveEQ(true),
			workflowdefinition.DraftEQ(false),
			workflowdefinition.OwnerIDEQ(orgID),
		)

	if eventType != "" {
		query = query.Where(stringArrayContains(workflowdefinition.FieldTriggerOperations, eventType))
	}

	allChanges := make([]string, 0, len(changedFields)+len(changedEdges))
	allChanges = append(allChanges, changedFields...)
	allChanges = append(allChanges, changedEdges...)
	if len(allChanges) > 0 {
		fieldPredicates := lo.Map(allChanges, func(field string, _ int) predicate.WorkflowDefinition {
			return stringArrayContains(workflowdefinition.FieldTriggerFields, field)
		})
		query = query.Where(workflowdefinition.Or(
			stringArrayEmpty(workflowdefinition.FieldTriggerFields),
			workflowdefinition.Or(fieldPredicates...),
		))
	}

	defs, err = query.All(allowCtx)
	if err != nil {
		return nil, scope.Fail(fmt.Errorf("%w: %w", ErrFailedToQueryDefinitions, err), nil)
	}

	var matching []*generated.WorkflowDefinition

	for _, def := range defs {
		if e.matchesTriggers(ctx, scope, def, eventType, changedFields, changedEdges, addedIDs, removedIDs, proposedChanges, obj) {
			matching = append(matching, def)
		}
	}

	return matching, nil
}

// matchesTriggers checks if the triggers match the event
func (e *WorkflowEngine) matchesTriggers(ctx context.Context, scope *observability.Scope, def *generated.WorkflowDefinition, eventType string, changedFields []string, changedEdges []string, addedIDs map[string][]string, removedIDs map[string][]string, proposedChanges map[string]any, obj *workflows.Object) bool {
	triggers := def.DefinitionJSON.Triggers
	if len(triggers) == 0 {
		return false
	}

	return lo.SomeBy(triggers, func(trigger models.WorkflowTrigger) bool {
		if trigger.Operation != eventType {
			return false
		}

		if !e.matchesSelector(ctx, scope, trigger.Selector, obj) {
			return false
		}

		// Evaluate trigger expression if present
		if trigger.Expression != "" {
			userID, _ := auth.GetSubjectIDFromContext(ctx)
			vars := workflows.BuildCELVars(obj, changedFields, changedEdges, addedIDs, removedIDs, eventType, userID, proposedChanges)

			result, err := e.celEvaluator.Evaluate(ctx, trigger.Expression, vars)
			if err != nil {
				scope.Warn(err, observability.Fields{
					workflowinstance.FieldWorkflowDefinitionID: def.ID,
					observability.FieldExpression:              trigger.Expression,
				})
				return false
			}

			if !result {
				return false
			}
		}

		// If no specific fields or edges are specified, any change to this operation type triggers
		if len(trigger.Fields) == 0 && len(trigger.Edges) == 0 {
			return true
		}

		// Check if any trigger field or edge is in the changed fields or edges
		allTriggerFields := make([]string, 0, len(trigger.Fields)+len(trigger.Edges))
		allTriggerFields = append(allTriggerFields, trigger.Fields...)
		allTriggerFields = append(allTriggerFields, trigger.Edges...)
		allChangedFields := make([]string, 0, len(changedFields)+len(changedEdges))
		allChangedFields = append(allChangedFields, changedFields...)
		allChangedFields = append(allChangedFields, changedEdges...)
		return len(lo.Intersect(allTriggerFields, allChangedFields)) > 0
	})
}

// matchesSelector checks if the object matches the trigger selector criteria
func (e *WorkflowEngine) matchesSelector(ctx context.Context, scope *observability.Scope, selector models.WorkflowSelector, obj *workflows.Object) bool {
	// If selector has objectTypes constraint, check if object type matches
	if len(selector.ObjectTypes) > 0 {
		if !lo.Contains(selector.ObjectTypes, obj.Type) {
			return false
		}
	}

	// If selector has tagIds constraint, check if object has any of those tags
	if len(selector.TagIDs) > 0 {
		objectTags, err := e.getObjectTags(ctx, obj)
		if err != nil {
			scope.Warn(err, observability.Fields(obj.ObservabilityFields()))
			return false
		}

		if len(lo.Intersect(selector.TagIDs, objectTags)) == 0 {
			return false
		}
	}

	// If selector has groupIds constraint, check if object belongs to any of those groups
	if len(selector.GroupIDs) > 0 {
		objectGroups, err := e.getObjectGroups(ctx, obj)
		if err != nil {
			scope.Warn(err, observability.Fields(obj.ObservabilityFields()))
			return false
		}

		if len(lo.Intersect(selector.GroupIDs, objectGroups)) == 0 {
			return false
		}
	}

	return true
}

// reEvaluateNotifyActions re-evaluates NOTIFY actions with When expressions after assignment changes.
// This enables dynamic notifications based on assignment state (e.g., notify when quorum is reached).
func (l *WorkflowListeners) reEvaluateNotifyActions(scope *observability.Scope, instance *generated.WorkflowInstance, obj *workflows.Object) {
	executedNotifications := l.getExecutedNotifications(instance)

	var newlyExecuted []string

	for i, action := range instance.DefinitionSnapshot.Actions {
		notifyKey, executed := l.processNotifyAction(scope, instance, obj, action, i, executedNotifications)
		if executed {
			newlyExecuted = append(newlyExecuted, notifyKey)
		}
	}

	// Persist newly executed notification keys to instance context
	if len(newlyExecuted) > 0 {
		l.trackExecutedNotifications(scope, instance.ID, newlyExecuted)
	}
}

// processNotifyAction evaluates and executes a NOTIFY action during assignment reconciliation
func (l *WorkflowListeners) processNotifyAction(scope *observability.Scope, instance *generated.WorkflowInstance, obj *workflows.Object, action models.WorkflowAction, index int, executed map[string]bool) (string, bool) {
	actionType := enums.ToWorkflowActionType(action.Type)
	if actionType == nil || *actionType != enums.WorkflowActionTypeNotification {
		return "", false
	}

	// Skip NOTIFY actions without When expressions (they execute in normal sequence)
	if action.When == "" {
		return "", false
	}

	notifyKey := notifyActionKey(action, index)
	if executed[notifyKey] {
		return "", false
	}

	ctx := scope.Context()
	shouldExecute, err := l.engine.EvaluateActionWhen(ctx, action.When, instance, obj)
	if err != nil {
		observability.WarnListener(ctx, observability.OpExecuteAction, action.Type, observability.ActionFields(action.Key, observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}), err)
		return "", false
	}

	if !shouldExecute {
		return "", false
	}

	if err := l.engine.Execute(ctx, action, instance, obj); err != nil {
		observability.WarnListener(ctx, observability.OpExecuteAction, action.Type, observability.ActionFields(action.Key, observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}), err)
		return "", false
	}

	l.recordEvent(scope, instance, enums.WorkflowEventTypeActionCompleted, action.Key, map[string]any{
		"triggered_by": "assignment_state_change",
	})

	return notifyKey, true
}

// notifyActionKey builds a stable key for a notification action execution
func notifyActionKey(action models.WorkflowAction, index int) string {
	return fmt.Sprintf("notify_%s_%d", action.Key, index)
}
