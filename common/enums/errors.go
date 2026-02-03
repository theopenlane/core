package enums

import "errors"

var (
	// ErrWrongTypeDirectoryAccountType indicates the value type for DirectoryAccountType is incorrect.
	ErrWrongTypeDirectoryAccountType = errors.New("wrong type for DirectoryAccountType")
	// ErrWrongTypeDirectoryAccountStatus indicates the value type for DirectoryAccountStatus is incorrect.
	ErrWrongTypeDirectoryAccountStatus = errors.New("wrong type for DirectoryAccountStatus")
	// ErrWrongTypeDirectoryAccountMFAState indicates the value type for DirectoryAccountMFAState is incorrect.
	ErrWrongTypeDirectoryAccountMFAState = errors.New("wrong type for DirectoryAccountMFAState")
	// ErrWrongTypeDirectoryGroupClassification indicates the value type for DirectoryGroupClassification is incorrect.
	ErrWrongTypeDirectoryGroupClassification = errors.New("wrong type for DirectoryGroupClassification")
	// ErrWrongTypeDirectoryGroupStatus indicates the value type for DirectoryGroupStatus is incorrect.
	ErrWrongTypeDirectoryGroupStatus = errors.New("wrong type for DirectoryGroupStatus")
	// ErrWrongTypeDirectoryMembershipRole indicates the value type for DirectoryMembershipRole is incorrect.
	ErrWrongTypeDirectoryMembershipRole = errors.New("wrong type for DirectoryMembershipRole")
	// ErrWrongTypeDirectorySyncRunStatus indicates the value type for DirectorySyncRunStatus is incorrect.
	ErrWrongTypeDirectorySyncRunStatus = errors.New("wrong type for DirectorySyncRunStatus")
	// ErrWrongTypeWorkflowKind indicates the value type for WorkflowKind is incorrect.
	ErrWrongTypeWorkflowKind = errors.New("wrong type for WorkflowKind")
	// ErrWrongTypeWorkflowInstanceState indicates the value type for WorkflowInstanceState is incorrect.
	ErrWrongTypeWorkflowInstanceState = errors.New("wrong type for WorkflowInstanceState")
	// ErrWrongTypeWorkflowAssignmentStatus indicates the value type for WorkflowAssignmentStatus is incorrect.
	ErrWrongTypeWorkflowAssignmentStatus = errors.New("wrong type for WorkflowAssignmentStatus")
	// ErrWrongTypeWorkflowTargetType indicates the value type for WorkflowTargetType is incorrect.
	ErrWrongTypeWorkflowTargetType = errors.New("wrong type for WorkflowTargetType")
	// ErrWrongTypeWorkflowObjectType indicates the value type for WorkflowObjectType is incorrect.
	ErrWrongTypeWorkflowObjectType = errors.New("wrong type for WorkflowObjectType")
	// ErrWrongTypeWorkflowEventType indicates the value type for WorkflowEventType is incorrect.
	ErrWrongTypeWorkflowEventType = errors.New("wrong type for WorkflowEventType")
	// ErrWrongTypeWorkflowProposalState indicates the value type for WorkflowProposalState is incorrect.
	ErrWrongTypeWorkflowProposalState = errors.New("wrong type for WorkflowProposalState")
	// ErrWrongTypeWorkflowApprovalSubmissionMode indicates the value type for WorkflowApprovalSubmissionMode is incorrect.
	ErrWrongTypeWorkflowApprovalSubmissionMode = errors.New("wrong type for WorkflowApprovalSubmissionMode")
	// ErrWrongTypeWorkflowActionType indicates the value type for WorkflowActionType is incorrect.
	ErrWrongTypeWorkflowActionType = errors.New("wrong type for WorkflowActionType")
	// ErrWrongTypeWorkflowApprovalTiming indicates the value type for WorkflowApprovalTiming is incorrect.
	ErrWrongTypeWorkflowApprovalTiming = errors.New("wrong type for WorkflowApprovalTiming")
)
