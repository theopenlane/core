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
	Name         string             `json:"name,omitempty"`
	Description  string             `json:"description,omitempty"`
	SchemaType   string             `json:"schemaType,omitempty"`
	WorkflowKind enums.WorkflowKind `json:"workflowKind,omitempty"`
	// ApprovalSubmissionMode controls draft vs auto-submit behavior for approval domains.
	ApprovalSubmissionMode enums.WorkflowApprovalSubmissionMode `json:"approvalSubmissionMode,omitempty"`
	Version                string                               `json:"version,omitempty"`
	Targets                WorkflowSelector                     `json:"targets,omitempty"`
	Triggers               []WorkflowTrigger                    `json:"triggers,omitempty"`
	Conditions             []WorkflowCondition                  `json:"conditions,omitempty"`
	Actions                []WorkflowAction                     `json:"actions,omitempty"`
	Metadata               map[string]any                       `json:"metadata,omitempty"`
}

// WorkflowTrigger describes when to run a workflow.
type WorkflowTrigger struct {
	Operation   string                   `json:"operation,omitempty"`   // e.g. CREATE, UPDATE, DELETE
	ObjectType  enums.WorkflowObjectType `json:"objectType,omitempty"`  // schema/object type the trigger targets
	Fields      []string                 `json:"fields,omitempty"`      // specific fields that should trigger evaluation
	Edges       []string                 `json:"edges,omitempty"`       // specific edges (relationships) that should trigger evaluation
	Selector    WorkflowSelector         `json:"selector,omitempty"`    // scoping for tags/groups/objects
	Expression  string                   `json:"expression,omitempty"`  // optional CEL expression
	Description string                   `json:"description,omitempty"` // human friendly description
}

// WorkflowCondition describes a CEL condition that must pass.
type WorkflowCondition struct {
	Expression  string `json:"expression,omitempty"`
	Description string `json:"description,omitempty"`
}

// WorkflowAction represents an action performed by the workflow.
type WorkflowAction struct {
	Key         string          `json:"key,omitempty"`    // unique key within the workflow
	Type        string          `json:"type,omitempty"`   // action type, e.g. REQUEST_APPROVAL, NOTIFY
	Params      json.RawMessage `json:"params,omitempty"` // opaque config for the action
	When        string          `json:"when,omitempty"`   // optional CEL expression to conditionally execute this action
	Description string          `json:"description,omitempty"`
}

// WorkflowSelector scopes workflows to tags, groups, or object types.
type WorkflowSelector struct {
	TagIDs      []string                   `json:"tagIds,omitempty"`
	GroupIDs    []string                   `json:"groupIds,omitempty"`
	ObjectTypes []enums.WorkflowObjectType `json:"objectTypes,omitempty"`
}

// WorkflowDefinitionSchema represents a template schema for definitions.
type WorkflowDefinitionSchema struct {
	Version string          `json:"version,omitempty"`
	Schema  json.RawMessage `json:"schema,omitempty"` // optional JSONSchema for validation
}

// WorkflowInstanceContext holds runtime context for a workflow instance.
type WorkflowInstanceContext struct {
	WorkflowDefinitionID string                      `json:"workflowDefinitionId,omitempty"`
	ObjectType           enums.WorkflowObjectType    `json:"objectType,omitempty"`
	ObjectID             string                      `json:"objectId,omitempty"`
	Version              int                         `json:"version,omitempty"`
	Assignments          []WorkflowAssignmentContext `json:"assignments,omitempty"`
	TriggerEventType     string                      `json:"triggerEventType,omitempty"`
	TriggerChangedFields []string                    `json:"triggerChangedFields,omitempty"`
	TriggerChangedEdges  []string                    `json:"triggerChangedEdges,omitempty"`
	TriggerAddedIDs      map[string][]string         `json:"triggerAddedIds,omitempty"`
	TriggerRemovedIDs    map[string][]string         `json:"triggerRemovedIds,omitempty"`
	TriggerUserID        string                      `json:"triggerUserId,omitempty"`
	Data                 json.RawMessage             `json:"data,omitempty"` // optional payload captured at runtime
}

// WorkflowAssignmentContext tracks an assignment decision within an instance.
type WorkflowAssignmentContext struct {
	AssignmentKey string                         `json:"assignmentKey,omitempty"`
	Status        enums.WorkflowAssignmentStatus `json:"status,omitempty"`
	ActorUserID   string                         `json:"actorUserId,omitempty"`
	ActorGroupID  string                         `json:"actorGroupId,omitempty"`
	DecidedAt     *time.Time                     `json:"decidedAt,omitempty"`
	Notes         string                         `json:"notes,omitempty"`
}

// WorkflowEventPayload stores workflow event payloads.
type WorkflowEventPayload struct {
	EventType enums.WorkflowEventType `json:"eventType,omitempty"`
	ActionKey string                  `json:"actionKey,omitempty"`
	Details   json.RawMessage         `json:"details,omitempty"`
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
	// PreviousStatus is the status before invalidation (e.g., "APPROVED")
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
	// RejectedAt is when the invalidation occurred
	RejectedAt string `json:"invalidated_at,omitempty"`
	// RejectedByUserID is the user who made the change that triggered rejection
	RejectedByUserID string `json:"invalidated_by_user_id,omitempty"`
	// RejectedHash is the hash that was rejected (encapsulating the changes that were not merged)
	RejectedHash string `json:"approved_hash,omitempty"`
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
