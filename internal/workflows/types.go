package workflows

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

// Object captures the workflow target along with its concrete ent entity when available.
type Object struct {
	// ID is the workflow object identifier.
	ID string
	// Type is the workflow object type.
	Type enums.WorkflowObjectType
	// Node is the concrete ent entity when available.
	Node any
}

// CELValue exposes the value used inside CEL expressions.
// Prefer the concrete ent entity so expressions can access real fields.
func (o *Object) CELValue() any {
	if o == nil {
		return nil
	}

	if o.Node != nil {
		return o.Node
	}

	return map[string]any{
		"id":   o.ID,
		"type": o.Type,
	}
}

// ObjectFromRef builds an Object from a WorkflowObjectRef record.
func ObjectFromRef(ref *generated.WorkflowObjectRef) (*Object, error) {
	for _, resolver := range objectFromRefRegistry {
		if obj, ok := resolver(ref); ok {
			return obj, nil
		}
	}

	return nil, ErrMissingObjectID
}

// TargetConfig defines who should receive workflow actions.
type TargetConfig struct {
	// Type selects how targets are resolved.
	Type enums.WorkflowTargetType `json:"type"`
	// ID identifies the target resource for static targets.
	ID string `json:"id,omitempty"`
	// ResolverKey names the resolver used for dynamic targets.
	ResolverKey string `json:"resolver_key,omitempty"`
}

// CELContextBuilder can override how CEL activation variables are built per object type.
// Codegen can register specialized builders (e.g., to expose typed fields) by calling
// RegisterCELContextBuilder in an init() function.
type CELContextBuilder func(obj *Object, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string, eventType, userID string) map[string]any

var celContextBuilders []CELContextBuilder

// RegisterCELContextBuilder adds a CEL context builder. Last registered wins.
func RegisterCELContextBuilder(builder CELContextBuilder) {
	celContextBuilders = append(celContextBuilders, builder)
}

// BuildCELVars constructs the activation map for CEL evaluation, allowing generated builders
// to provide typed contexts instead of ad-hoc maps
func BuildCELVars(obj *Object, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string, eventType, userID string) map[string]any {
	for i := len(celContextBuilders) - 1; i >= 0; i-- {
		if vars := celContextBuilders[i](obj, changedFields, changedEdges, addedIDs, removedIDs, eventType, userID); vars != nil {
			return vars
		}
	}

	objectValue := any(nil)
	if obj != nil {
		objectValue = obj.CELValue()
	}

	return map[string]any{
		"object":         objectValue,
		"changed_fields": changedFields,
		"changed_edges":  changedEdges,
		"added_ids":      addedIDs,
		"removed_ids":    removedIDs,
		"event_type":     eventType,
		"user_id":        userID,
	}
}

// ObjectFromRefRegistry allows generated code to register new mappings from WorkflowObjectRef to Object.
// This makes adding new workflowable schemas additive (codegen can call RegisterObjectRefResolver in init()).
var objectFromRefRegistry []func(*generated.WorkflowObjectRef) (*Object, bool)

// RegisterObjectRefResolver adds a resolver to the registry.
func RegisterObjectRefResolver(resolver func(*generated.WorkflowObjectRef) (*Object, bool)) {
	objectFromRefRegistry = append(objectFromRefRegistry, resolver)
}

// ObjectRefQueryBuilder allows generated code to register WorkflowObjectRef predicates per object type.
type ObjectRefQueryBuilder func(*generated.WorkflowObjectRefQuery, *Object) (*generated.WorkflowObjectRefQuery, bool)

var objectRefQueryBuilders []ObjectRefQueryBuilder

// RegisterObjectRefQueryBuilder adds a WorkflowObjectRef query builder to the registry.
func RegisterObjectRefQueryBuilder(builder ObjectRefQueryBuilder) {
	objectRefQueryBuilders = append(objectRefQueryBuilders, builder)
}

// AssignmentContextBuilder builds workflow runtime context (assignments, instance, initiator) for CEL evaluation.
// Generated code registers this to provide assignment state when evaluating NOTIFY action When expressions.
type AssignmentContextBuilder func(ctx context.Context, client *generated.Client, instanceID string) (map[string]any, error)

var assignmentContextBuilder AssignmentContextBuilder

// RegisterAssignmentContextBuilder sets the assignment context builder for CEL evaluation.
// Only one builder is needed since the generated code provides a comprehensive implementation.
func RegisterAssignmentContextBuilder(builder AssignmentContextBuilder) {
	assignmentContextBuilder = builder
}

// BuildAssignmentContext calls the registered assignment context builder to get workflow runtime state.
// Returns nil if no builder is registered or instanceID is empty.
func BuildAssignmentContext(ctx context.Context, client *generated.Client, instanceID string) (map[string]any, error) {
	if assignmentContextBuilder == nil {
		return nil, nil
	}

	return assignmentContextBuilder(ctx, client, instanceID)
}
