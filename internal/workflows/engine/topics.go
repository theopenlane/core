package engine

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/gala"
)

// WorkflowCommandTopics is the namespace for workflow command topics
var WorkflowCommandTopics = gala.Workflow.Child("command")

// WorkflowTriggeredPayload contains data for a workflow instance creation event
type WorkflowTriggeredPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string `json:"instance_id"`
	// DefinitionID is the identifier for the workflow definition
	DefinitionID string `json:"definition_id"`
	// ObjectID is the identifier for the object associated with the workflow
	ObjectID string `json:"object_id"`
	// ObjectType is the type of the object associated with the workflow
	ObjectType enums.WorkflowObjectType `json:"object_type"`
	// TriggerEventType is the event type that triggered the workflow
	TriggerEventType string `json:"trigger_event_type"`
	// TriggerChangedFields are the fields that changed and triggered the workflow
	TriggerChangedFields []string `json:"trigger_changed_fields,omitempty"`
}

// WorkflowActionStartedPayload contains data for when a workflow action begins
type WorkflowActionStartedPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string `json:"instance_id"`
	// ActionIndex is the index of the action in the workflow
	ActionIndex int `json:"action_index"`
	// ActionType is the type of action being started
	ActionType enums.WorkflowActionType `json:"action_type"`
	// ObjectID is the identifier for the object associated with the workflow
	ObjectID string `json:"object_id"`
	// ObjectType is the type of the object associated with the workflow
	ObjectType enums.WorkflowObjectType `json:"object_type"`
}

// WorkflowActionCompletedPayload contains data for when a workflow action finishes
type WorkflowActionCompletedPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string `json:"instance_id"`
	// ActionIndex is the index of the action in the workflow
	ActionIndex int `json:"action_index"`
	// ActionType is the type of action being completed
	ActionType enums.WorkflowActionType `json:"action_type"`
	// ObjectID is the identifier for the object associated with the workflow
	ObjectID string `json:"object_id"`
	// ObjectType is the type of the object associated with the workflow
	ObjectType enums.WorkflowObjectType `json:"object_type"`
	// Success indicates if the action completed successfully
	Success bool `json:"success"`
	// Skipped indicates if the action was skipped
	Skipped bool `json:"skipped,omitempty"`
	// ErrorMessage contains the error message if the action failed
	ErrorMessage string `json:"error_message,omitempty"`
}

// WorkflowAssignmentCompletedPayload contains data for completed assignments
type WorkflowAssignmentCompletedPayload struct {
	// AssignmentID is the unique identifier for the assignment
	AssignmentID string `json:"assignment_id"`
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string `json:"instance_id"`
	// Status is the status of the assignment
	Status enums.WorkflowAssignmentStatus `json:"status"`
	// CompletedBy is the identifier of the user who completed the assignment
	CompletedBy string `json:"completed_by,omitempty"`
	// ObjectID is the identifier for the object associated with the workflow
	ObjectID string `json:"object_id"`
	// ObjectType is the type of the object associated with the workflow
	ObjectType enums.WorkflowObjectType `json:"object_type"`
}

// WorkflowInstanceCompletedPayload contains data for completed instances
type WorkflowInstanceCompletedPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string `json:"instance_id"`
	// State is the terminal state of the workflow instance
	State enums.WorkflowInstanceState `json:"state"`
	// ObjectID is the identifier for the object associated with the workflow
	ObjectID string `json:"object_id"`
	// ObjectType is the type of the object associated with the workflow
	ObjectType enums.WorkflowObjectType `json:"object_type"`
}

var (
	// WorkflowTriggeredEventTopic is the typed topic for workflow triggered events
	WorkflowTriggeredEventTopic = gala.NamespacedTopic[WorkflowTriggeredPayload](WorkflowCommandTopics, "trigger")
	// WorkflowActionStartedEventTopic is the typed topic for action started events
	WorkflowActionStartedEventTopic = gala.NamespacedTopic[WorkflowActionStartedPayload](WorkflowCommandTopics, "advance")
	// WorkflowActionCompletedEventTopic is the typed topic for action completed events
	WorkflowActionCompletedEventTopic = gala.NamespacedTopic[WorkflowActionCompletedPayload](WorkflowCommandTopics, "action_completed")
	// WorkflowAssignmentCompletedEventTopic is the typed topic for assignment completed events
	WorkflowAssignmentCompletedEventTopic = gala.NamespacedTopic[WorkflowAssignmentCompletedPayload](WorkflowCommandTopics, "assignment_completed")
	// WorkflowInstanceCompletedEventTopic is the typed topic for instance completed events
	WorkflowInstanceCompletedEventTopic = gala.NamespacedTopic[WorkflowInstanceCompletedPayload](WorkflowCommandTopics, "instance_completed")
)
