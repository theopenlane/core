package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/enums"
)

// WorkflowDefinitionDocument represents the stored workflow definition with typed fields.
type WorkflowDefinitionDocument struct {
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	SchemaType   string                 `json:"schemaType,omitempty"`
	WorkflowKind enums.WorkflowKind     `json:"workflowKind,omitempty"`
	Version      string                 `json:"version,omitempty"`
	Targets      WorkflowSelector       `json:"targets,omitempty"`
	Triggers     []WorkflowTrigger      `json:"triggers,omitempty"`
	Conditions   []WorkflowCondition    `json:"conditions,omitempty"`
	Actions      []WorkflowAction       `json:"actions,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowTrigger describes when to run a workflow.
type WorkflowTrigger struct {
	Operation   string                   `json:"operation,omitempty"`   // e.g. CREATE, UPDATE, DELETE
	ObjectType  enums.WorkflowObjectType `json:"objectType,omitempty"`  // schema/object type the trigger targets
	Fields      []string                 `json:"fields,omitempty"`      // specific fields that should trigger evaluation
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
func (d *WorkflowDefinitionDocument) UnmarshalGQL(v interface{}) error {
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
func (d *WorkflowDefinitionSchema) UnmarshalGQL(v interface{}) error {
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
func (c *WorkflowInstanceContext) UnmarshalGQL(v interface{}) error {
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
func (p *WorkflowEventPayload) UnmarshalGQL(v interface{}) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, p)
}
