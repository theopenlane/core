package engine

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
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
