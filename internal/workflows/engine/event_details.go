package engine

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const (
	approvalRejectedMessage   = "approval rejected"
	approvalQuorumNotMet      = "approval quorum not met"
	approvalFailedMessage     = "approval failed"
	actionExecutionFailedMsg  = "action execution failed"
	actionIndexOutOfBoundsMsg = "action index out of bounds"
)

// assignmentCreatedDetails captures a single assignment batch for UI history
type assignmentCreatedDetails struct {
	// ActionKey is the workflow action key
	ActionKey string `json:"action_key"`
	// ActionIndex is the index of the workflow action
	ActionIndex int `json:"action_index"`
	// ActionType is the action type
	ActionType enums.WorkflowActionType `json:"action_type"`
	// ObjectID is the target object ID
	ObjectID string `json:"object_id"`
	// ObjectType is the target object type
	ObjectType enums.WorkflowObjectType `json:"object_type"`
	// AssignmentIDs are the created assignment IDs
	AssignmentIDs []string `json:"assignment_ids"`
	// TargetUserIDs are the resolved user IDs for the assignment
	TargetUserIDs []string `json:"target_user_ids"`
	// Required indicates whether approvals are required
	Required bool `json:"required"`
	// RequiredCount is the approval quorum when set
	RequiredCount int `json:"required_count,omitempty"`
	// Label is the optional assignment label
	Label string `json:"label,omitempty"`
}

// actionCompletedDetails captures completion metadata for a workflow action
type actionCompletedDetails struct {
	// ActionKey is the workflow action key
	ActionKey string `json:"action_key"`
	// ActionIndex is the index of the workflow action
	ActionIndex int `json:"action_index"`
	// ActionType is the action type
	ActionType enums.WorkflowActionType `json:"action_type"`
	// ObjectID is the target object ID
	ObjectID string `json:"object_id"`
	// ObjectType is the target object type
	ObjectType enums.WorkflowObjectType `json:"object_type"`
	// Success indicates the action completed successfully
	Success bool `json:"success"`
	// Skipped indicates the action was skipped
	Skipped bool `json:"skipped,omitempty"`
	// ErrorMessage captures a failure reason
	ErrorMessage string `json:"error_message,omitempty"`
	// AssignmentIDs lists assignments when the action is an approval
	AssignmentIDs []string `json:"assignment_ids,omitempty"`
	// TargetUserIDs lists assignment target users
	TargetUserIDs []string `json:"target_user_ids,omitempty"`
	// ApprovedCount is the count of approved assignments
	ApprovedCount int `json:"approved_count,omitempty"`
	// RejectedCount is the count of rejected assignments
	RejectedCount int `json:"rejected_count,omitempty"`
	// PendingCount is the count of pending assignments
	PendingCount int `json:"pending_count,omitempty"`
	// RequiredCount is the approval quorum when set
	RequiredCount int `json:"required_count,omitempty"`
	// Required indicates whether approvals are required
	Required bool `json:"required,omitempty"`
	// Label is the optional assignment label
	Label string `json:"label,omitempty"`
}

// actionCompletedDetailsFromPayload builds details from a completion payload
func actionCompletedDetailsFromPayload(actionKey string, payload soiree.WorkflowActionCompletedPayload) actionCompletedDetails {
	return actionCompletedDetails{
		ActionKey:    actionKey,
		ActionIndex:  payload.ActionIndex,
		ActionType:   payload.ActionType,
		ObjectID:     payload.ObjectID,
		ObjectType:   payload.ObjectType,
		Success:      payload.Success,
		Skipped:      payload.Skipped,
		ErrorMessage: payload.ErrorMessage,
	}
}

// actionFailureDetails builds a failed completion detail for a specific action
func actionFailureDetails(actionKey string, actionIndex int, actionType enums.WorkflowActionType, obj *workflows.Object, err error) actionCompletedDetails {
	details := actionCompletedDetails{
		ActionKey:   actionKey,
		ActionIndex: actionIndex,
		ActionType:  actionType,
		Success:     false,
	}
	if obj != nil {
		details.ObjectID = obj.ID
		details.ObjectType = obj.Type
	}
	if err != nil {
		details.ErrorMessage = err.Error()
	}
	if details.ErrorMessage == "" {
		details.ErrorMessage = actionExecutionFailedMsg
	}
	return details
}

// actionIndexOutOfBoundsDetails builds a failure detail for out of bounds actions
func actionIndexOutOfBoundsDetails(payload soiree.WorkflowActionStartedPayload) actionCompletedDetails {
	return actionCompletedDetails{
		ActionKey:    "",
		ActionIndex:  payload.ActionIndex,
		ActionType:   payload.ActionType,
		ObjectID:     payload.ObjectID,
		ObjectType:   payload.ObjectType,
		Success:      false,
		ErrorMessage: actionIndexOutOfBoundsMsg,
	}
}

// assignmentCreatedDetailsForApproval builds assignment created details for approval actions
func assignmentCreatedDetailsForApproval(action models.WorkflowAction, actionIndex int, obj *workflows.Object, assignmentIDs, targetUserIDs []string, required bool, requiredCount int, label string) assignmentCreatedDetails {
	details := assignmentCreatedDetails{
		ActionKey:     action.Key,
		ActionIndex:   actionIndex,
		ActionType:    enums.WorkflowActionTypeApproval,
		AssignmentIDs: assignmentIDs,
		TargetUserIDs: targetUserIDs,
		Required:      required,
		RequiredCount: requiredCount,
		Label:         label,
	}
	if obj != nil {
		details.ObjectID = obj.ID
		details.ObjectType = obj.Type
	}
	return details
}

// approvalCompletedDetails builds action completion details for approval actions
func approvalCompletedDetails(action models.WorkflowAction, actionIndex int, obj *workflows.Object, counts AssignmentStatusCounts, requiredCount int, assignmentIDs, targetUserIDs []string, success bool, label string, required bool) actionCompletedDetails {
	details := actionCompletedDetails{
		ActionKey:     action.Key,
		ActionIndex:   actionIndex,
		ActionType:    enums.WorkflowActionTypeApproval,
		Success:       success,
		AssignmentIDs: assignmentIDs,
		TargetUserIDs: targetUserIDs,
		ApprovedCount: counts.Approved,
		RejectedCount: counts.Rejected,
		PendingCount:  counts.Pending,
		RequiredCount: requiredCount,
		Required:      required,
		Label:         label,
	}
	if obj != nil {
		details.ObjectID = obj.ID
		details.ObjectType = obj.Type
	}
	if !success {
		details.ErrorMessage = approvalFailureMessage(counts, requiredCount)
	}
	return details
}

// approvalFailureMessage returns a user-facing failure message for approvals
func approvalFailureMessage(counts AssignmentStatusCounts, requiredCount int) string {
	if counts.RejectedRequired {
		return approvalRejectedMessage
	}
	if requiredCount > 0 && counts.Approved < requiredCount && counts.Pending == 0 {
		return approvalQuorumNotMet
	}
	return approvalFailedMessage
}
