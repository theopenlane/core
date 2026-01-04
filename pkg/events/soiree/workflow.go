package soiree

import (
	"encoding/json"
	"fmt"

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

// UnwrapWorkflowPayload is a helper that unwraps workflow payloads with type checking
func UnwrapWorkflowPayload[T any](event Event) (T, error) {
	var zero T

	if event == nil {
		return zero, ErrNilPayload
	}

	payload := event.Payload()
	if payload == nil {
		return zero, ErrNilPayload
	}

	typed, ok := payload.(T)
	if ok {
		return typed, nil
	}

	var raw json.RawMessage
	switch v := payload.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	default:
		encoded, err := json.Marshal(payload)
		if err != nil {
			return zero, fmt.Errorf("%w: expected %T, got %T", ErrPayloadTypeMismatch, zero, payload)
		}
		raw = encoded
	}

	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return zero, fmt.Errorf("%w: expected %T, got %T", ErrPayloadTypeMismatch, zero, payload)
	}

	event.SetPayload(decoded)

	return decoded, nil
}

// WorkflowTriggeredTopic is emitted when a workflow instance is created
var WorkflowTriggeredTopic = NewTypedTopic(
	TopicWorkflowTriggered,
	func(p WorkflowTriggeredPayload) Event {
		return NewBaseEvent(TopicWorkflowTriggered, p)
	},
	UnwrapWorkflowPayload[WorkflowTriggeredPayload],
)

// WorkflowActionStartedTopic is emitted when a workflow action begins execution
var WorkflowActionStartedTopic = NewTypedTopic(
	TopicWorkflowActionStarted,
	func(p WorkflowActionStartedPayload) Event {
		return NewBaseEvent(TopicWorkflowActionStarted, p)
	},
	UnwrapWorkflowPayload[WorkflowActionStartedPayload],
)

// WorkflowActionCompletedTopic is emitted when a workflow action finishes
var WorkflowActionCompletedTopic = NewTypedTopic(
	TopicWorkflowActionCompleted,
	func(p WorkflowActionCompletedPayload) Event {
		return NewBaseEvent(TopicWorkflowActionCompleted, p)
	},
	UnwrapWorkflowPayload[WorkflowActionCompletedPayload],
)

// WorkflowAssignmentCreatedTopic is emitted when an approval is assigned
var WorkflowAssignmentCreatedTopic = NewTypedTopic(
	TopicWorkflowAssignmentCreated,
	func(p WorkflowAssignmentCreatedPayload) Event {
		return NewBaseEvent(TopicWorkflowAssignmentCreated, p)
	},
	UnwrapWorkflowPayload[WorkflowAssignmentCreatedPayload],
)

// WorkflowAssignmentCompletedTopic is emitted when an approval decision is made
var WorkflowAssignmentCompletedTopic = NewTypedTopic(
	TopicWorkflowAssignmentCompleted,
	func(p WorkflowAssignmentCompletedPayload) Event {
		return NewBaseEvent(TopicWorkflowAssignmentCompleted, p)
	},
	UnwrapWorkflowPayload[WorkflowAssignmentCompletedPayload],
)

// WorkflowInstanceCompletedTopic is emitted when a workflow finishes
var WorkflowInstanceCompletedTopic = NewTypedTopic(
	TopicWorkflowInstanceCompleted,
	func(p WorkflowInstanceCompletedPayload) Event {
		return NewBaseEvent(TopicWorkflowInstanceCompleted, p)
	},
	UnwrapWorkflowPayload[WorkflowInstanceCompletedPayload],
)

// WorkflowTimeoutExpiredTopic is emitted when a timeout occurs
var WorkflowTimeoutExpiredTopic = NewTypedTopic(
	TopicWorkflowTimeoutExpired,
	func(p WorkflowTimeoutExpiredPayload) Event {
		return NewBaseEvent(TopicWorkflowTimeoutExpired, p)
	},
	UnwrapWorkflowPayload[WorkflowTimeoutExpiredPayload],
)

// MutationDetectedTopic is emitted when a mutation occurs that might trigger workflows
var MutationDetectedTopic = NewTypedTopic(
	TopicMutationDetected,
	func(p MutationDetectedPayload) Event {
		return NewBaseEvent(TopicMutationDetected, p)
	},
	UnwrapWorkflowPayload[MutationDetectedPayload],
)
