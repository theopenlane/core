package workflows

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/iam/auth"
)

// WorkflowCreationStage identifies which step failed when building instances/object refs.
type WorkflowCreationStage string

const (
	WorkflowCreationStageInstance  WorkflowCreationStage = "instance"
	WorkflowCreationStageObjectRef WorkflowCreationStage = "object_ref"
)

// WorkflowCreationError wraps the underlying error and indicates what stage failed.
type WorkflowCreationError struct {
	// Stage is the creation stage that failed
	Stage WorkflowCreationStage
	// Err is the underlying error
	Err error
}

// Error formats the workflow creation error
func (e *WorkflowCreationError) Error() string {
	return fmt.Sprintf("workflow %s creation failed: %v", e.Stage, e.Err)
}

// Unwrap returns the underlying error
func (e *WorkflowCreationError) Unwrap() error {
	return e.Err
}

// BuildProposedChanges materializes mutation values (including cleared fields) for workflow proposals.
func BuildProposedChanges(m utils.GenericMutation, changedFields []string) map[string]any {
	clearedSet := lo.SliceToMap(m.ClearedFields(), func(f string) (string, struct{}) {
		return f, struct{}{}
	})

	proposed := make(map[string]any, len(changedFields))
	lo.ForEach(changedFields, func(field string, _ int) {
		if val, ok := m.Field(field); ok {
			proposed[field] = val
			return
		}
		if _, ok := clearedSet[field]; ok {
			proposed[field] = nil
		}
	})

	return proposed
}

// DefinitionMatchesTrigger reports whether the workflow definition has a trigger that matches the event type and changed fields.
func DefinitionMatchesTrigger(doc models.WorkflowDefinitionDocument, eventType string, changedFields []string) bool {
	targetEvent := strings.ToUpper(eventType)

	for _, trigger := range doc.Triggers {
		if trigger.Operation != "" {
			if targetEvent == "" || strings.ToUpper(trigger.Operation) != targetEvent {
				continue
			}
		}

		if len(trigger.Fields) == 0 && len(trigger.Edges) == 0 {
			return true
		}

		allFields := lo.Flatten([][]string{trigger.Fields, trigger.Edges})
		if len(allFields) == 0 {
			return true
		}

		if len(lo.Intersect(allFields, changedFields)) > 0 {
			return true
		}
	}

	return false
}

// ResolveOwnerID returns the provided owner ID or derives it from the context when empty.
func ResolveOwnerID(ctx context.Context, ownerID string) (string, error) {
	if ownerID != "" {
		return ownerID, nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	return orgID, nil
}

// WorkflowInstanceBuilderParams defines the inputs for creating a workflow instance + object ref.
type WorkflowInstanceBuilderParams struct {
	// WorkflowDefinitionID is the definition id for the instance
	WorkflowDefinitionID string
	// DefinitionSnapshot is the snapshot of the workflow definition
	DefinitionSnapshot models.WorkflowDefinitionDocument
	// State is the initial workflow instance state
	State enums.WorkflowInstanceState
	// Context is the workflow instance context payload
	Context models.WorkflowInstanceContext
	// OwnerID is the organization owner id
	OwnerID string
	// ObjectType is the workflow object type
	ObjectType enums.WorkflowObjectType
	// ObjectID is the workflow object id
	ObjectID string
}

// CreateWorkflowInstanceWithObjectRef builds a workflow instance and its object ref in a single helper.
func CreateWorkflowInstanceWithObjectRef(ctx context.Context, tx *generated.Tx, params WorkflowInstanceBuilderParams) (*generated.WorkflowInstance, *generated.WorkflowObjectRef, error) {
	instanceCreate := tx.WorkflowInstance.Create().
		SetWorkflowDefinitionID(params.WorkflowDefinitionID).
		SetState(params.State).
		SetDefinitionSnapshot(params.DefinitionSnapshot).
		SetContext(params.Context).
		SetOwnerID(params.OwnerID)

	instanceCreate = generated.SetWorkflowInstanceObjectID(instanceCreate, params.ObjectType, params.ObjectID)

	instance, err := instanceCreate.Save(ctx)
	if err != nil {
		return nil, nil, &WorkflowCreationError{
			Stage: WorkflowCreationStageInstance,
			Err:   err,
		}
	}

	objectRefCreate := tx.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetOwnerID(params.OwnerID)
	objectRefCreate = generated.SetWorkflowObjectRefObjectID(objectRefCreate, params.ObjectType, params.ObjectID)

	objectRef, err := objectRefCreate.Save(ctx)
	if err != nil {
		return nil, nil, &WorkflowCreationError{
			Stage: WorkflowCreationStageObjectRef,
			Err:   err,
		}
	}

	return instance, objectRef, nil
}

// BuildWorkflowActionContext returns the replacement map and base data for notifications/webhooks.
func BuildWorkflowActionContext(instance *generated.WorkflowInstance, obj *Object, actionKey string) (map[string]string, map[string]any) {
	replacements := map[string]string{
		"instance_id":   "",
		"definition_id": "",
		"object_id":     "",
		"object_type":   "",
		"action_key":    actionKey,
	}

	if instance != nil {
		replacements["instance_id"] = instance.ID
		replacements["definition_id"] = instance.WorkflowDefinitionID
	}

	if obj != nil {
		replacements["object_id"] = obj.ID
		replacements["object_type"] = obj.Type.String()
	}

	data := lo.MapValues(replacements, func(value string, _ string) any { return value })

	return replacements, data
}

// FindProposalForObjectRefs locates a workflow proposal for the given object refs and domain key.
// The lookup first honors priorityStates (each queried individually) before falling back to states (queried together).
func FindProposalForObjectRefs(ctx context.Context, client *generated.Client, objRefIDs []string, domainKey string, priorityStates []enums.WorkflowProposalState, states []enums.WorkflowProposalState) (*generated.WorkflowProposal, error) {
	if len(objRefIDs) == 0 || len(states) == 0 {
		return nil, nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	queryBase := func() *generated.WorkflowProposalQuery {
		return client.WorkflowProposal.Query().
			Where(
				workflowproposal.WorkflowObjectRefIDIn(objRefIDs...),
				workflowproposal.DomainKeyEQ(domainKey),
				workflowproposal.OwnerIDEQ(orgID),
			)
	}

	search := func(states []enums.WorkflowProposalState) (*generated.WorkflowProposal, error) {
		q := queryBase()
		if len(states) == 0 {
			return nil, nil
		}
		if len(states) > 1 {
			q = q.Where(workflowproposal.StateIn(states...))
		} else {
			q = q.Where(workflowproposal.StateEQ(states[0]))
		}
		q = q.Order(generated.Desc(workflowproposal.FieldUpdatedAt))
		proposal, err := q.First(ctx)
		if err != nil && !generated.IsNotFound(err) {
			return nil, err
		}
		return proposal, nil
	}

	stateGroups := lo.Map(priorityStates, func(state enums.WorkflowProposalState, _ int) []enums.WorkflowProposalState {
		return []enums.WorkflowProposalState{state}
	})
	stateGroups = append(stateGroups, states)

	for _, group := range stateGroups {
		proposal, err := search(group)
		if err != nil {
			return nil, err
		}
		if proposal != nil {
			return proposal, nil
		}
	}

	return nil, nil
}

// ValidateCELExpression ensures the expression compiles against a CEL environment
// configured with the provided workflow config and scope.
func ValidateCELExpression(cfg *Config, scope CELExpressionScope, expression string) error {
	env, err := NewCELEnv(cfg, scope)
	if err != nil {
		return err
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("CEL compilation failed: %w", issues.Err())
	}

	if _, err := env.Program(ast); err != nil {
		return fmt.Errorf("CEL program creation failed: %w", err)
	}

	return nil
}

// ResolveUserDisplayName fetches a user by ID and returns their display name (FirstName LastName).
// If the user cannot be found or has no name, returns the original userID as fallback.
func ResolveUserDisplayName(ctx context.Context, client *generated.Client, userID string) string {
	user, err := client.User.Get(ctx, userID)
	if err != nil {
		return userID
	}

	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name == "" {
		return userID
	}

	return name
}

// GetObjectUpdatedBy extracts the UpdatedBy field from an Object's Node if available.
// Returns empty string if the object or node is nil, or if UpdatedBy is not accessible.
func GetObjectUpdatedBy(obj *Object) string {
	if obj == nil || obj.Node == nil {
		return ""
	}

	var fields struct {
		UpdatedBy string `json:"updated_by"`
	}

	if err := jsonx.RoundTrip(obj.Node, &fields); err != nil {
		return ""
	}

	return fields.UpdatedBy
}

// MutationPayload carries the raw ent mutation, the resolved operation, the entity ID and the ent
// client so listeners can act without additional lookups
type MutationPayload struct {
	// Mutation is the raw ent mutation that triggered the event
	Mutation ent.Mutation
	// MutationType is the ent schema type that emitted the mutation
	MutationType string
	// Operation is the string representation of the mutation operation
	Operation string
	// EntityID is the ID of the entity that was mutated
	EntityID string
	// ChangedFields captures updated/cleared fields for the mutation
	ChangedFields []string
	// ClearedFields captures fields explicitly cleared in the mutation
	ClearedFields []string
	// ChangedEdges captures changed edge names for workflow-eligible edges
	ChangedEdges []string
	// AddedIDs captures edge IDs added by edge name
	AddedIDs map[string][]string
	// RemovedIDs captures edge IDs removed by edge name
	RemovedIDs map[string][]string
	// ProposedChanges captures field-level proposed values (including nil for clears)
	ProposedChanges map[string]any
	// Client is the ent client that can be used to perform additional queries or mutations
	Client *generated.Client
}

// MutationEntityID derives the entity identifier from the payload or event properties.
func MutationEntityID(ctx *soiree.EventContext, payload *MutationPayload) (string, bool) {
	if payload != nil && payload.EntityID != "" {
		return payload.EntityID, true
	}

	if ctx == nil {
		return "", false
	}

	if id, ok := ctx.PropertyString("ID"); ok && id != "" {
		return id, true
	}

	if raw, ok := ctx.Property("ID"); ok && raw != nil {
		if str, ok := raw.(fmt.Stringer); ok {
			value := str.String()
			if value == "" {
				return "", false
			}

			return value, true
		}

		value := fmt.Sprint(raw)
		if value == "" || value == "<nil>" {
			return "", false
		}

		return value, true
	}

	return "", false
}

// MutationClient returns the ent client associated with the mutation.
func MutationClient(ctx *soiree.EventContext, payload *MutationPayload) *generated.Client {
	if payload != nil && payload.Client != nil {
		return payload.Client
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	return client
}
