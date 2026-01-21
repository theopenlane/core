package workflows

import "errors"

var (
	// ErrWorkflowNotFound is returned when a workflow instance cannot be found
	ErrWorkflowNotFound = errors.New("workflow instance not found")
	// ErrNoProposedChanges is returned when attempting to apply changes but none exist
	ErrNoProposedChanges = errors.New("no proposed changes to apply")
	// ErrWorkflowAlreadyActive is returned when trying to start a workflow on an object that already has one
	ErrWorkflowAlreadyActive = errors.New("workflow already active for this object")
	// ErrInvalidWorkflowState is returned when a workflow operation is attempted on an instance in an invalid state
	ErrInvalidWorkflowState = errors.New("invalid workflow state for this operation")
	// ErrInsufficientApprovals is returned when trying to complete a workflow without required approvals
	ErrInsufficientApprovals = errors.New("insufficient approvals to complete workflow")
	// ErrUnauthorizedApproval is returned when a user attempts to approve a workflow they're not assigned to
	ErrUnauthorizedApproval = errors.New("user not authorized to approve this workflow")
	// ErrFieldNotWorkflowEligible is returned when proposed changes include non-eligible fields
	ErrFieldNotWorkflowEligible = errors.New("field is not eligible for workflow modification")
	// ErrUnsupportedObjectType is returned when a workflow is triggered for an unsupported object type
	ErrUnsupportedObjectType = errors.New("object type does not support workflows")
	// ErrMissingObjectID is returned when a workflow object is missing an ID
	ErrMissingObjectID = errors.New("workflow object is missing an ID")
	// ErrFailedToBuildCELEnv is returned when the CEL environment cannot be built
	ErrFailedToBuildCELEnv = errors.New("failed to build CEL environment")
)
