package workflows

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// init wires the workflow-owned context builders. Entity identity, fields, and edges are resolved
// directly from entityops rather than copied into mutable registries.
func init() {
	RegisterCELContextBuilder(buildCELContext)
	RegisterAssignmentContextBuilder(buildAssignmentContext)
	RegisterObservabilityFieldsBuilder(buildObservabilityFields)
}

// buildCELContext builds the CEL activation variables for any workflow object: the ent entity is
// JSON round-tripped so field names match JSON tags and enums become strings, which is
// type-agnostic by construction
func buildCELContext(obj *Object, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string, eventType, userID string, proposedChanges map[string]any) map[string]any {
	if obj == nil || obj.Node == nil {
		return nil
	}

	objectMap, err := jsonx.ToMap(obj.Node)
	if err != nil {
		return nil
	}

	return map[string]any{
		"object":           objectMap,
		"changed_fields":   changedFields,
		"changed_edges":    changedEdges,
		"added_ids":        addedIDs,
		"removed_ids":      removedIDs,
		"event_type":       eventType,
		"user_id":          userID,
		"proposed_changes": proposedChanges,
	}
}

// buildObservabilityFields returns standard log fields for a workflow object, resolving the
// object-ref foreign-key column for the object's type from the entityops edge catalog
func buildObservabilityFields(obj *Object) map[string]any {
	if obj == nil {
		return nil
	}

	fields := map[string]any{
		"object_type": obj.Type.String(),
	}

	for _, edge := range entityops.SchemaWorkflowObjectRef.Edges {
		if edge.Unique && edge.Field != "" && edge.TargetType == obj.Type.String() {
			fields[edge.Field] = obj.ID
			break
		}
	}

	return fields
}

// buildAssignmentContext builds workflow runtime context (assignments, instance, initiator) for
// CEL evaluation; the summary is JSON round-tripped because CEL traverses maps, not Go structs
func buildAssignmentContext(ctx context.Context, client *generated.Client, instanceID string) (map[string]any, error) {
	if client == nil || instanceID == "" {
		return nil, nil
	}

	summary, err := client.BuildAssignmentSummary(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	summaryMap, err := jsonx.ToMap(summary)
	if err != nil {
		return nil, err
	}

	instance, err := client.WorkflowInstance.Get(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	instanceContext := map[string]any{
		"id":                   instance.ID,
		"state":                instance.State.String(),
		"current_action_index": instance.CurrentActionIndex,
	}

	initiator := ""
	if instance.Context.TriggerUserID != "" {
		initiator = instance.Context.TriggerUserID
	}

	return map[string]any{
		"assignments": summaryMap,
		"instance":    instanceContext,
		"initiator":   initiator,
	}, nil
}
