package soiree

import (
	"github.com/theopenlane/core/common/enums"
)

// WorkflowTriggeredPayload contains data for a workflow instance creation event
type WorkflowTriggeredPayload struct {
	InstanceID           string
	DefinitionID         string
	ObjectID             string
	ObjectType           enums.WorkflowObjectType
	TriggerEventType     string
	TriggerChangedFields []string
}

// WorkflowActionStartedPayload contains data for when a workflow action begins
type WorkflowActionStartedPayload struct {
	InstanceID  string
	ActionIndex int
	ActionType  enums.WorkflowActionType
	ObjectID    string
	ObjectType  enums.WorkflowObjectType
}

// WorkflowActionCompletedPayload contains data for when a workflow action finishes
type WorkflowActionCompletedPayload struct {
	InstanceID   string
	ActionIndex  int
	ActionType   enums.WorkflowActionType
	ObjectID     string
	ObjectType   enums.WorkflowObjectType
	Success      bool
	Skipped      bool
	ErrorMessage string
}

// WorkflowAssignmentCreatedPayload contains data for when an approval is assigned
type WorkflowAssignmentCreatedPayload struct {
	AssignmentID string
	InstanceID   string
	TargetType   enums.WorkflowTargetType
	TargetIDs    []string
	ObjectID     string
	ObjectType   enums.WorkflowObjectType
}

// WorkflowAssignmentCompletedPayload contains data for when an approval decision is made
type WorkflowAssignmentCompletedPayload struct {
	AssignmentID string
	InstanceID   string
	Status       enums.WorkflowAssignmentStatus
	CompletedBy  string
	ObjectID     string
	ObjectType   enums.WorkflowObjectType
}

// WorkflowInstanceCompletedPayload contains data for when a workflow finishes
type WorkflowInstanceCompletedPayload struct {
	InstanceID string
	State      enums.WorkflowInstanceState
	ObjectID   string
	ObjectType enums.WorkflowObjectType
}

// WorkflowTimeoutExpiredPayload contains data for when a timeout occurs
type WorkflowTimeoutExpiredPayload struct {
	InstanceID   string
	AssignmentID string
	ObjectID     string
	ObjectType   enums.WorkflowObjectType
}

// MutationDetectedPayload contains data for mutations that might trigger workflows
type MutationDetectedPayload struct {
	SchemaType    string
	ObjectID      string
	Operation     string
	ChangedFields []string
	UserID        string
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
