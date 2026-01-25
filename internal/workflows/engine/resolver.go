package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/core/internal/workflows/resolvers"
	"github.com/theopenlane/iam/auth"
)

// ResolveTargets converts a TargetConfig into concrete user IDs.
func (e *WorkflowEngine) ResolveTargets(ctx context.Context, target workflows.TargetConfig, obj *workflows.Object) ([]string, error) {
	switch target.Type {
	case enums.WorkflowTargetTypeUser:
		if target.ID == "" {
			return nil, fmt.Errorf("%w: user target requires ID", ErrMissingRequiredField)
		}
		return []string{target.ID}, nil

	case enums.WorkflowTargetTypeGroup:
		if target.ID == "" {
			return nil, fmt.Errorf("%w: group target requires ID", ErrMissingRequiredField)
		}
		return resolvers.ResolveGroupMembers(ctx, e.client, target.ID)

	case enums.WorkflowTargetTypeRole:
		if target.ID == "" {
			return nil, fmt.Errorf("%w: role target requires ID", ErrMissingRequiredField)
		}
		return e.resolveRoleMembers(ctx, target.ID, obj)

	case enums.WorkflowTargetTypeResolver:
		if target.ResolverKey == "" {
			return nil, fmt.Errorf("%w: resolver target requires resolver_key", ErrMissingRequiredField)
		}

		resolver, ok := resolvers.Get(target.ResolverKey)
		if !ok {
			observability.WarnEngine(ctx, observability.OpResolveTargets, target.Type.String(), observability.Fields{
				workflowassignmenttarget.FieldResolverKey: target.ResolverKey,
			}, nil)
			return []string{}, nil
		}

		return resolver(ctx, e.client, obj)

	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidTargetType, target.Type)
	}
}

// resolveRoleMembers returns all user IDs with a specific role.
func (e *WorkflowEngine) resolveRoleMembers(ctx context.Context, roleID string, obj *workflows.Object) ([]string, error) {
	if obj == nil {
		return nil, ErrObjectRefMissingID
	}

	role := enums.ToRole(roleID)
	if role == nil || *role == enums.RoleInvalid {
		return nil, fmt.Errorf("%w: invalid role %q", ErrMissingRequiredField, roleID)
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	memberships, err := e.client.OrgMembership.
		Query().
		Where(
			orgmembership.OrganizationIDEQ(orgID),
			orgmembership.RoleEQ(*role),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve role members: %w", err)
	}

	userIDs := lo.Map(memberships, func(m *generated.OrgMembership, _ int) string {
		return m.UserID
	})

	return lo.Uniq(userIDs), nil
}

// getObjectTags retrieves tag IDs for a workflow object
func (e *WorkflowEngine) getObjectTags(ctx context.Context, obj *workflows.Object) ([]string, error) {
	entity, err := e.loadObjectNode(ctx, obj)
	if err != nil {
		return nil, err
	}

	tags, err := workflows.StringSliceField(entity, "tags")
	if err != nil {
		return nil, err
	}
	tags = workflows.NormalizeStrings(tags)
	if len(tags) == 0 {
		return nil, nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tagIDs, err := e.client.TagDefinition.Query().
		Where(
			tagdefinition.OwnerIDEQ(orgID),
			tagdefinition.NameIn(tags...),
		).
		IDs(ctx)
	if err != nil {
		return nil, err
	}

	return tagIDs, nil
}

// getObjectGroups retrieves group IDs for a workflow object
func (e *WorkflowEngine) getObjectGroups(ctx context.Context, obj *workflows.Object) ([]string, error) {
	entity, err := e.loadObjectNode(ctx, obj)
	if err != nil {
		return nil, err
	}

	groupIDs := []string{}

	if q, ok := entity.(interface{ QueryEditors() *generated.GroupQuery }); ok {
		ids, err := q.QueryEditors().IDs(ctx)
		if err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, ids...)
	}

	if q, ok := entity.(interface{ QueryViewers() *generated.GroupQuery }); ok {
		ids, err := q.QueryViewers().IDs(ctx)
		if err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, ids...)
	}

	if q, ok := entity.(interface{ QueryBlockedGroups() *generated.GroupQuery }); ok {
		ids, err := q.QueryBlockedGroups().IDs(ctx)
		if err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, ids...)
	}

	if len(groupIDs) == 0 {
		return nil, nil
	}

	return lo.Uniq(groupIDs), nil
}

// loadObjectNode resolves and caches the workflow object node
func (e *WorkflowEngine) loadObjectNode(ctx context.Context, obj *workflows.Object) (any, error) {
	if obj.Node != nil {
		return obj.Node, nil
	}

	entity, err := workflows.LoadWorkflowObject(ctx, e.client, obj.Type.String(), obj.ID)
	if err != nil {
		return nil, err
	}

	obj.Node = entity
	return entity, nil
}

// ResolveAssignmentState checks assignments for an instance and updates instance state.
// This is intended for tests that do not use event emitters. It synchronously performs
// the same state resolution that HandleAssignmentCompleted does via events.
func (e *WorkflowEngine) ResolveAssignmentState(ctx context.Context, instanceID string) error {
	allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		return err
	}

	instance, err := e.client.WorkflowInstance.Query().Where(workflowinstance.IDEQ(instanceID), workflowinstance.OwnerIDEQ(orgID)).Only(allowCtx)
	if err != nil {
		return fmt.Errorf("failed to load workflow instance: %w", err)
	}

	def := instance.DefinitionSnapshot

	assignments, err := e.client.WorkflowAssignment.Query().Where(workflowassignment.WorkflowInstanceIDEQ(instanceID), workflowassignment.OwnerIDEQ(orgID)).All(allowCtx)
	if err != nil {
		return fmt.Errorf("failed to load assignments: %w", err)
	}

	if len(assignments) == 0 {
		return nil
	}

	// Check each approval action in sequence
	lastCompletedAction := -1
	for i, action := range def.Actions {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType == nil || *actionType != enums.WorkflowActionTypeApproval {
			lastCompletedAction = i
			continue
		}

		prefix := fmt.Sprintf("approval_%s_", action.Key)
		actionAssignments := lo.Filter(assignments, func(a *generated.WorkflowAssignment, _ int) bool {
			return strings.HasPrefix(a.AssignmentKey, prefix)
		})

		if len(actionAssignments) == 0 {
			lastCompletedAction = i
			continue
		}

		meta := actionAssignments[0].ApprovalMetadata
		required := actionAssignments[0].Required
		requiredCount := requiredApprovalCount(action, meta, required)
		statusCounts := CountAssignmentStatus(actionAssignments)
		switch resolveApproval(requiredCount, statusCounts) {
		case approvalFailed:
			return e.client.WorkflowInstance.Update().Where(workflowinstance.IDEQ(instanceID), workflowinstance.OwnerIDEQ(orgID)).SetState(enums.WorkflowInstanceStateFailed).Exec(allowCtx)
		case approvalPending:
			return nil
		}

		lastCompletedAction = i
	}

	if instance.WorkflowProposalID != "" {
		objRef, err := e.client.WorkflowObjectRef.Query().Where(workflowobjectref.WorkflowInstanceIDEQ(instanceID), workflowobjectref.OwnerIDEQ(orgID)).First(allowCtx)
		if err != nil {
			return fmt.Errorf("failed to load object ref: %w", err)
		}
		obj, err := workflows.ObjectFromRef(objRef)
		if err != nil {
			return err
		}
		scope := observability.BeginEngine(ctx, e.observer, observability.OpCompleteAssignment, "resolve", nil)
		if err := e.proposalManager.Apply(scope, instance.WorkflowProposalID, obj); err != nil {
			return fmt.Errorf("failed to apply proposal: %w", err)
		}
	}

	nextIndex := lastCompletedAction + 1
	if nextIndex >= len(def.Actions) {
		return e.client.WorkflowInstance.Update().Where(workflowinstance.IDEQ(instanceID), workflowinstance.OwnerIDEQ(orgID)).SetState(enums.WorkflowInstanceStateCompleted).SetCurrentActionIndex(nextIndex).Exec(allowCtx)
	}

	return e.client.WorkflowInstance.Update().Where(workflowinstance.IDEQ(instanceID), workflowinstance.OwnerIDEQ(orgID)).SetState(enums.WorkflowInstanceStateRunning).SetCurrentActionIndex(nextIndex).Exec(allowCtx)
}
