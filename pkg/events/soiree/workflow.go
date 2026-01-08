package soiree

import (
	"github.com/theopenlane/core/common/enums"
)

// WorkflowTriggeredPayload contains data for a workflow instance creation event
type WorkflowTriggeredPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string
	// DefinitionID is the identifier for the workflow definition
	DefinitionID string
	// ObjectID is the ID of the object the workflow is acting on
	ObjectID string
	// ObjectType is the type of the object the workflow is acting on
	ObjectType enums.WorkflowObjectType
	// TriggerEventType is the event type that triggered the workflow
	TriggerEventType string
	// TriggerChangedFields are the fields that changed and triggered the workflow
	TriggerChangedFields []string
}

// WorkflowActionStartedPayload contains data for when a workflow action begins
type WorkflowActionStartedPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string
	// ActionIndex is the index of the action in the workflow
	ActionIndex int
	// ActionType is the type of action being started
	ActionType enums.WorkflowActionType
	// ObjectID is the ID of the object the action is acting on
	ObjectID string
	// ObjectType is the type of the object the action is acting on
	ObjectType enums.WorkflowObjectType
}

// WorkflowActionCompletedPayload contains data for when a workflow action finishes
type WorkflowActionCompletedPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string
	// ActionIndex is the index of the action in the workflow
	ActionIndex int
	// ActionType is the type of action that was completed
	ActionType enums.WorkflowActionType
	// ObjectID is the ID of the object the action was acting on
	ObjectID string
	// ObjectType is the type of the object the action was acting on
	ObjectType enums.WorkflowObjectType
	// Success indicates if the action completed successfully
	Success bool
	// Skipped indicates if the action was skipped
	Skipped bool
	// ErrorMessage contains any error message if the action failed
	ErrorMessage string
}

// WorkflowAssignmentCreatedPayload contains data for when an approval is assigned
type WorkflowAssignmentCreatedPayload struct {
	// AssignmentID is the unique identifier for the assignment
	AssignmentID string
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string
	// TargetType is the type of the target for the assignment
	TargetType enums.WorkflowTargetType
	// TargetIDs are the IDs of the targets for the assignment
	TargetIDs []string
	// ObjectID is the ID of the object the assignment is related to
	ObjectID string
	// ObjectType is the type of the object the assignment is related to
	ObjectType enums.WorkflowObjectType
}

// WorkflowAssignmentCompletedPayload contains data for when an approval decision is made
type WorkflowAssignmentCompletedPayload struct {
	// AssignmentID is the unique identifier for the assignment
	AssignmentID string
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string
	// Status is the status of the assignment after completion
	Status enums.WorkflowAssignmentStatus
	// CompletedBy is the ID of the user who completed the assignment
	CompletedBy string
	// ObjectID is the ID of the object the assignment is related to
	ObjectID string
	// ObjectType is the type of the object the assignment is related to
	ObjectType enums.WorkflowObjectType
}

// WorkflowInstanceCompletedPayload contains data for when a workflow finishes
type WorkflowInstanceCompletedPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string
	// State is the final state of the workflow instance
	State enums.WorkflowInstanceState
	// ObjectID is the ID of the object the workflow was acting on
	ObjectID string
	// ObjectType is the type of the object the workflow was acting on
	ObjectType enums.WorkflowObjectType
}

// WorkflowTimeoutExpiredPayload contains data for when a timeout occurs
type WorkflowTimeoutExpiredPayload struct {
	// InstanceID is the unique identifier for the workflow instance
	InstanceID string
	// AssignmentID is the unique identifier for the assignment that timed out
	AssignmentID string
	// ObjectID is the ID of the object related to the timeout
	ObjectID string
	// ObjectType is the type of the object related to the timeout
	ObjectType enums.WorkflowObjectType
}

// MutationDetectedPayload contains data for mutations that might trigger workflows
type MutationDetectedPayload struct {
	// SchemaType is the type of schema where the mutation occurred
	SchemaType string
	// ObjectID is the ID of the object that was mutated
	ObjectID string
	// Operation is the type of operation performed (e.g., update, delete)
	Operation string
	// ChangedFields are the fields that were changed in the mutation
	ChangedFields []string
	// UserID is the ID of the user who performed the mutation
	UserID string
}

// Topic name constants
const (
	TopicWorkflowTriggered           = "workflow.triggered"
	TopicWorkflowActionStarted       = "workflow.action.started"
	TopicWorkflowActionCompleted     = "workflow.action.completed"
	TopicWorkflowAssignmentCreated   = "workflow.assignment.created"
	TopicWorkflowAssignmentCompleted = "workflow.assignment.completed"
	TopicWorkflowInstanceCompleted   = "workflow.instance.completed"
	TopicWorkflowTimeoutExpired      = "workflow.timeout.expired"
	TopicMutationDetected            = "mutation.detected"
)

// WorkflowTriggeredTopic is emitted when a workflow instance is created
var WorkflowTriggeredTopic = NewTypedTopic[WorkflowTriggeredPayload](TopicWorkflowTriggered)

// WorkflowActionStartedTopic is emitted when a workflow action begins execution
var WorkflowActionStartedTopic = NewTypedTopic[WorkflowActionStartedPayload](TopicWorkflowActionStarted)

// WorkflowActionCompletedTopic is emitted when a workflow action finishes
var WorkflowActionCompletedTopic = NewTypedTopic[WorkflowActionCompletedPayload](TopicWorkflowActionCompleted)

// WorkflowAssignmentCreatedTopic is emitted when an approval is assigned
var WorkflowAssignmentCreatedTopic = NewTypedTopic[WorkflowAssignmentCreatedPayload](TopicWorkflowAssignmentCreated)

// WorkflowAssignmentCompletedTopic is emitted when an approval decision is made
var WorkflowAssignmentCompletedTopic = NewTypedTopic[WorkflowAssignmentCompletedPayload](TopicWorkflowAssignmentCompleted)

// WorkflowInstanceCompletedTopic is emitted when a workflow finishes
var WorkflowInstanceCompletedTopic = NewTypedTopic[WorkflowInstanceCompletedPayload](TopicWorkflowInstanceCompleted)

// WorkflowTimeoutExpiredTopic is emitted when a timeout occurs
var WorkflowTimeoutExpiredTopic = NewTypedTopic[WorkflowTimeoutExpiredPayload](TopicWorkflowTimeoutExpired)

// MutationDetectedTopic is emitted when a mutation occurs that might trigger workflows
var MutationDetectedTopic = NewTypedTopic[MutationDetectedPayload](TopicMutationDetected)
