package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/common/enums"
)

// WorkflowDefinitionDocument represents the stored workflow definition with typed fields.
type WorkflowDefinitionDocument struct {
	// Name is the workflow definition name
	Name string `json:"name,omitempty"`
	// Description describes what the workflow does
	Description string `json:"description,omitempty"`
	// SchemaType is the primary schema/object type targeted by the workflow
	SchemaType string `json:"schemaType,omitempty"`
	// WorkflowKind selects the workflow execution behavior
	WorkflowKind enums.WorkflowKind `json:"workflowKind,omitempty"`
	// ApprovalSubmissionMode controls draft vs auto-submit behavior for approval domains
	ApprovalSubmissionMode enums.WorkflowApprovalSubmissionMode `json:"approvalSubmissionMode,omitempty"`
	// ApprovalTiming controls whether approvals block changes or happen after commit
	ApprovalTiming enums.WorkflowApprovalTiming `json:"approvalTiming,omitempty"`
	// Version tracks the definition document version
	Version string `json:"version,omitempty"`
	// Targets scopes which objects are eligible for this definition
	Targets WorkflowSelector `json:"targets,omitempty"`
	// Triggers defines which events start workflow evaluation
	Triggers []WorkflowTrigger `json:"triggers,omitempty"`
	// Conditions defines CEL predicates that must pass for execution
	Conditions []WorkflowCondition `json:"conditions,omitempty"`
	// Actions defines the ordered workflow steps to execute
	Actions []WorkflowAction `json:"actions,omitempty"`
	// Metadata stores optional extensible workflow metadata
	Metadata map[string]any `json:"metadata,omitempty"`
}

// WorkflowTrigger describes when to run a workflow.
type WorkflowTrigger struct {
	// Operation is the mutation operation that triggers evaluation such as CREATE UPDATE or DELETE
	Operation string `json:"operation,omitempty"`
	// Interval is the schedule interval for periodic triggers such as 1h
	Interval string `json:"interval,omitempty"`
	// ObjectType is the schema/object type targeted by this trigger
	ObjectType enums.WorkflowObjectType `json:"objectType,omitempty"`
	// Fields limits evaluation to changes on specific fields
	Fields []string `json:"fields,omitempty"`
	// Edges limits evaluation to changes on specific relationships
	Edges []string `json:"edges,omitempty"`
	// Selector further scopes trigger matching using tags groups or object types
	Selector WorkflowSelector `json:"selector,omitempty"`
	// Expression is an optional CEL expression gate for trigger matching
	Expression string `json:"expression,omitempty"`
	// Description is a human-readable trigger description
	Description string `json:"description,omitempty"`
}

// WorkflowCondition describes a CEL condition that must pass.
type WorkflowCondition struct {
	// Expression is the CEL condition that must evaluate to true
	Expression string `json:"expression,omitempty"`
	// Description is a human-readable condition description
	Description string `json:"description,omitempty"`
}

// WorkflowAction represents an action performed by the workflow.
type WorkflowAction struct {
	// Key is the unique action key within the workflow
	Key string `json:"key,omitempty"`
	// Type is the action type such as REQUEST_APPROVAL or NOTIFY
	Type string `json:"type,omitempty"`
	// Params contains action-specific configuration payload
	Params json.RawMessage `json:"params,omitempty"`
	// When is an optional CEL expression that conditionally executes the action
	When string `json:"when,omitempty"`
	// Description is a human-readable action description
	Description string `json:"description,omitempty"`
}

// WorkflowSelector scopes workflows to tags, groups, or object types.
type WorkflowSelector struct {
	// TagIDs scopes matching to objects carrying any of these tags
	TagIDs []string `json:"tagIds,omitempty"`
	// GroupIDs scopes matching to objects associated with any of these groups
	GroupIDs []string `json:"groupIds,omitempty"`
	// ObjectTypes scopes matching to specific workflow object types
	ObjectTypes []enums.WorkflowObjectType `json:"objectTypes,omitempty"`
}

// WorkflowDefinitionSchema represents a template schema for definitions.
type WorkflowDefinitionSchema struct {
	// Version tracks the schema document version
	Version string `json:"version,omitempty"`
	// Schema contains the optional JSONSchema used to validate definitions
	Schema json.RawMessage `json:"schema,omitempty"`
}

// WorkflowInstanceContext holds runtime context for a workflow instance.
type WorkflowInstanceContext struct {
	// WorkflowDefinitionID is the definition that produced this instance
	WorkflowDefinitionID string `json:"workflowDefinitionId,omitempty"`
	// ObjectType is the workflow object type under evaluation
	ObjectType enums.WorkflowObjectType `json:"objectType,omitempty"`
	// ObjectID is the concrete object identifier under evaluation
	ObjectID string `json:"objectId,omitempty"`
	// Version increments as instance context evolves
	Version int `json:"version,omitempty"`
	// Assignments snapshots assignment decisions for context-aware evaluation
	Assignments []WorkflowAssignmentContext `json:"assignments,omitempty"`
	// TriggerEventType is the source event type that triggered this instance
	TriggerEventType string `json:"triggerEventType,omitempty"`
	// TriggerChangedFields lists fields changed by the triggering mutation
	TriggerChangedFields []string `json:"triggerChangedFields,omitempty"`
	// TriggerChangedEdges lists relationships changed by the triggering mutation
	TriggerChangedEdges []string `json:"triggerChangedEdges,omitempty"`
	// TriggerAddedIDs maps relationship names to added identifiers from the triggering mutation
	TriggerAddedIDs map[string][]string `json:"triggerAddedIds,omitempty"`
	// TriggerRemovedIDs maps relationship names to removed identifiers from the triggering mutation
	TriggerRemovedIDs map[string][]string `json:"triggerRemovedIds,omitempty"`
	// TriggerUserID is the actor that initiated the triggering mutation
	TriggerUserID string `json:"triggerUserId,omitempty"`
	// TriggerProposedChanges contains normalized proposed field values from the triggering mutation
	TriggerProposedChanges map[string]any `json:"triggerProposedChanges,omitempty"`
	// ParallelApprovalKeys tracks approval action keys that can execute concurrently
	ParallelApprovalKeys []string `json:"parallelApprovalKeys,omitempty"`
	// Data carries optional runtime payload captured with the instance
	Data json.RawMessage `json:"data,omitempty"`
}

// WorkflowAssignmentContext tracks an assignment decision within an instance.
type WorkflowAssignmentContext struct {
	// AssignmentKey is the workflow action key that produced this assignment
	AssignmentKey string `json:"assignmentKey,omitempty"`
	// Status is the current assignment status
	Status enums.WorkflowAssignmentStatus `json:"status,omitempty"`
	// ActorUserID is the user actor associated with the decision when available
	ActorUserID string `json:"actorUserId,omitempty"`
	// ActorGroupID is the group actor associated with the decision when available
	ActorGroupID string `json:"actorGroupId,omitempty"`
	// DecidedAt is when the assignment transitioned to a decided state
	DecidedAt *time.Time `json:"decidedAt,omitempty"`
	// Notes stores optional assignment decision notes
	Notes string `json:"notes,omitempty"`
}

// WorkflowEventPayload stores workflow event payloads.
type WorkflowEventPayload struct {
	// EventType identifies the workflow event kind
	EventType enums.WorkflowEventType `json:"eventType,omitempty"`
	// ActionKey identifies the related action when applicable
	ActionKey string `json:"actionKey,omitempty"`
	// Details stores event-specific payload data
	Details json.RawMessage `json:"details,omitempty"`
}

// AssignmentMetadata captures structured metadata for workflow assignments
type WorkflowAssignmentApproval struct {
	// ActionKey is the workflow action key this assignment belongs to
	ActionKey string `json:"action_key,omitempty"`
	// Required indicates if this assignment is required for workflow progression
	Required bool `json:"required,omitempty"`
	// RequiredCount is the quorum count needed if using count-based approval
	RequiredCount int `json:"required_count,omitempty"`
	// Label is an optional human-readable label for the assignment
	Label string `json:"label,omitempty"`
	// ProposedHash is the hash of the proposal changes when this assignment was created
	ProposedHash string `json:"proposed_hash,omitempty"`
	// ApprovedAt captures when the assignment was approved
	ApprovedAt string `json:"approved_at,omitempty"`
	// ApprovedByUserID is the user who approved the assignment
	ApprovedByUserID string `json:"approved_by_user_id,omitempty"`
}

// AssignmentInvalidation captures details when an approval is invalidated (approvals are invalidated when there is a subsequent change to the proposed changes)
type WorkflowAssignmentInvalidation struct {
	// Reason explains why the approval was invalidated
	Reason string `json:"reason,omitempty"`
	// PreviousStatus is the status before invalidation such as APPROVED
	PreviousStatus string `json:"previous_status,omitempty"`
	// InvalidatedAt is when the invalidation occurred
	InvalidatedAt string `json:"invalidated_at,omitempty"`
	// InvalidatedByUserID is the user who made the change that triggered invalidation
	InvalidatedByUserID string `json:"invalidated_by_user_id,omitempty"`
	// ApprovedHash is the hash that was approved before invalidation
	ApprovedHash string `json:"approved_hash,omitempty"`
	// NewProposedHash is the new hash after the changes that triggered invalidation
	NewProposedHash string `json:"new_proposed_hash,omitempty"`
}

// WorkflowAssignmentRejection captures details when an approval is rejected / denied
type WorkflowAssignmentRejection struct {
	// ActionKey is the workflow action key this assignment belongs to
	ActionKey string `json:"action_key,omitempty"`
	// RejectionReason stores an optional rejection reason
	RejectionReason string `json:"rejection_reason,omitempty"`
	// RejectedAt is when the rejection occurred
	RejectedAt string `json:"invalidated_at,omitempty"`
	// RejectedByUserID is the user who made the rejection decision
	RejectedByUserID string `json:"invalidated_by_user_id,omitempty"`
	// RejectedHash is the hash that was rejected encapsulating the changes that were not merged
	RejectedHash string `json:"approved_hash,omitempty"`
	// ChangeRequestInputs stores optional structured inputs for change requests
	ChangeRequestInputs map[string]any `json:"change_request_inputs,omitempty"`
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (d WorkflowAssignmentApproval) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(d)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling workflow definition document")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing workflow definition document")
	}
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (d *WorkflowAssignmentApproval) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, d)
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (d WorkflowAssignmentInvalidation) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(d)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling workflow definition document")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing workflow definition document")
	}
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (d *WorkflowAssignmentInvalidation) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, d)
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (d WorkflowAssignmentRejection) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(d)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling workflow definition document")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing workflow definition document")
	}
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (d *WorkflowAssignmentRejection) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, d)
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (d WorkflowDefinitionDocument) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(d)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling workflow definition document")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing workflow definition document")
	}
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (d *WorkflowDefinitionDocument) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, d)
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (d WorkflowDefinitionSchema) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(d)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling workflow definition schema")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing workflow definition schema")
	}
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (d *WorkflowDefinitionSchema) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, d)
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (c WorkflowInstanceContext) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(c)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling workflow instance context")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing workflow instance context")
	}
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (c *WorkflowInstanceContext) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, c)
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (p WorkflowEventPayload) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(p)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling workflow event payload")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing workflow event payload")
	}
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (p *WorkflowEventPayload) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, p)
}
