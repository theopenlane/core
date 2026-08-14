package gala

import "github.com/theopenlane/core/common/enums"

// WorkflowCommandTopics is the namespace for workflow command topics
var WorkflowCommandTopics = NewTopicNamespace(TopicPrefixWorkflowCommand, JobKindWorkflow)

const (
	// TopicWorkflowTriggered is emitted when a workflow instance is created
	TopicWorkflowTriggered TopicName = TopicPrefixWorkflowCommand + "trigger"
	// TopicWorkflowActionStarted is emitted when a workflow action begins
	TopicWorkflowActionStarted TopicName = TopicPrefixWorkflowCommand + "advance"
	// TopicWorkflowActionCompleted is emitted when a workflow action finishes
	TopicWorkflowActionCompleted TopicName = TopicPrefixWorkflowCommand + "action_completed"
	// TopicWorkflowAssignmentCompleted is emitted when an assignment resolves
	TopicWorkflowAssignmentCompleted TopicName = TopicPrefixWorkflowCommand + "assignment_completed"
	// TopicWorkflowInstanceCompleted is emitted when an instance reaches a terminal state
	TopicWorkflowInstanceCompleted TopicName = TopicPrefixWorkflowCommand + "instance_completed"
)

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
	// WorkflowTriggeredEventTopic is the typed topic for workflow triggered events.
	WorkflowTriggeredEventTopic = Topic[WorkflowTriggeredPayload]{Name: TopicWorkflowTriggered, Kind: WorkflowCommandTopics.Kind()}
	// WorkflowActionStartedEventTopic is the typed topic for action started events.
	WorkflowActionStartedEventTopic = Topic[WorkflowActionStartedPayload]{Name: TopicWorkflowActionStarted, Kind: WorkflowCommandTopics.Kind()}
	// WorkflowActionCompletedEventTopic is the typed topic for action completed events.
	WorkflowActionCompletedEventTopic = Topic[WorkflowActionCompletedPayload]{Name: TopicWorkflowActionCompleted, Kind: WorkflowCommandTopics.Kind()}
	// WorkflowAssignmentCompletedEventTopic is the typed topic for assignment completed events.
	WorkflowAssignmentCompletedEventTopic = Topic[WorkflowAssignmentCompletedPayload]{Name: TopicWorkflowAssignmentCompleted, Kind: WorkflowCommandTopics.Kind()}
	// WorkflowInstanceCompletedEventTopic is the typed topic for instance completed events.
	WorkflowInstanceCompletedEventTopic = Topic[WorkflowInstanceCompletedPayload]{Name: TopicWorkflowInstanceCompleted, Kind: WorkflowCommandTopics.Kind()}
)
