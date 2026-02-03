package enums

import (
	"fmt"
	"io"
	"strings"
)

// WorkflowKind enumerates workflow kinds.
type WorkflowKind string

var (
	WorkflowKindApproval     WorkflowKind = "APPROVAL"
	WorkflowKindLifecycle    WorkflowKind = "LIFECYCLE"
	WorkflowKindNotification WorkflowKind = "NOTIFICATION"
)

var WorkflowKinds = []string{
	string(WorkflowKindApproval),
	string(WorkflowKindLifecycle),
	string(WorkflowKindNotification),
}

func (WorkflowKind) Values() (kinds []string) {
	return WorkflowKinds
}

func (r WorkflowKind) String() string { return string(r) }

func ToWorkflowKind(v string) *WorkflowKind {
	switch strings.ToUpper(v) {
	case WorkflowKindApproval.String():
		return &WorkflowKindApproval
	case WorkflowKindLifecycle.String():
		return &WorkflowKindLifecycle
	case WorkflowKindNotification.String():
		return &WorkflowKindNotification
	default:
		return nil
	}
}

func (r WorkflowKind) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowKind) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowKind, v)
	}
	*r = WorkflowKind(str)
	return nil
}

// WorkflowInstanceState enumerates instance states.
type WorkflowInstanceState string

var (
	WorkflowInstanceStateRunning   WorkflowInstanceState = "RUNNING"
	WorkflowInstanceStateCompleted WorkflowInstanceState = "COMPLETED"
	WorkflowInstanceStateFailed    WorkflowInstanceState = "FAILED"
	WorkflowInstanceStatePaused    WorkflowInstanceState = "PAUSED"
)

var WorkflowInstanceStates = []string{
	string(WorkflowInstanceStateRunning),
	string(WorkflowInstanceStateCompleted),
	string(WorkflowInstanceStateFailed),
	string(WorkflowInstanceStatePaused),
}

func (WorkflowInstanceState) Values() (states []string) {
	return WorkflowInstanceStates
}

func (r WorkflowInstanceState) String() string { return string(r) }

func ToWorkflowInstanceState(v string) *WorkflowInstanceState {
	switch strings.ToUpper(v) {
	case WorkflowInstanceStateRunning.String():
		return &WorkflowInstanceStateRunning
	case WorkflowInstanceStateCompleted.String():
		return &WorkflowInstanceStateCompleted
	case WorkflowInstanceStateFailed.String():
		return &WorkflowInstanceStateFailed
	case WorkflowInstanceStatePaused.String():
		return &WorkflowInstanceStatePaused
	default:
		return nil
	}
}

func (r WorkflowInstanceState) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowInstanceState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowInstanceState, v)
	}
	*r = WorkflowInstanceState(str)
	return nil
}

// WorkflowAssignmentStatus enumerates assignment statuses.
type WorkflowAssignmentStatus string

var (
	WorkflowAssignmentStatusPending          WorkflowAssignmentStatus = "PENDING"
	WorkflowAssignmentStatusApproved         WorkflowAssignmentStatus = "APPROVED"
	WorkflowAssignmentStatusRejected         WorkflowAssignmentStatus = "REJECTED"
	WorkflowAssignmentStatusChangesRequested WorkflowAssignmentStatus = "CHANGES_REQUESTED"
)

var WorkflowAssignmentStatuses = []string{
	string(WorkflowAssignmentStatusPending),
	string(WorkflowAssignmentStatusApproved),
	string(WorkflowAssignmentStatusRejected),
	string(WorkflowAssignmentStatusChangesRequested),
}

func (WorkflowAssignmentStatus) Values() (vals []string) {
	return WorkflowAssignmentStatuses
}

func (r WorkflowAssignmentStatus) String() string { return string(r) }

func ToWorkflowAssignmentStatus(v string) *WorkflowAssignmentStatus {
	switch strings.ToUpper(v) {
	case WorkflowAssignmentStatusPending.String():
		return &WorkflowAssignmentStatusPending
	case WorkflowAssignmentStatusApproved.String():
		return &WorkflowAssignmentStatusApproved
	case WorkflowAssignmentStatusRejected.String():
		return &WorkflowAssignmentStatusRejected
	case WorkflowAssignmentStatusChangesRequested.String():
		return &WorkflowAssignmentStatusChangesRequested
	default:
		return nil
	}
}

func (r WorkflowAssignmentStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowAssignmentStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowAssignmentStatus, v)
	}
	*r = WorkflowAssignmentStatus(str)
	return nil
}

// WorkflowProposalState enumerates proposal lifecycle states.
type WorkflowProposalState string

var (
	WorkflowProposalStateDraft      WorkflowProposalState = "DRAFT"
	WorkflowProposalStateSubmitted  WorkflowProposalState = "SUBMITTED"
	WorkflowProposalStateApplied    WorkflowProposalState = "APPLIED"
	WorkflowProposalStateRejected   WorkflowProposalState = "REJECTED"
	WorkflowProposalStateSuperseded WorkflowProposalState = "SUPERSEDED"
)

var WorkflowProposalStates = []string{
	string(WorkflowProposalStateDraft),
	string(WorkflowProposalStateSubmitted),
	string(WorkflowProposalStateApplied),
	string(WorkflowProposalStateRejected),
	string(WorkflowProposalStateSuperseded),
}

func (WorkflowProposalState) Values() (vals []string) {
	return WorkflowProposalStates
}

func (r WorkflowProposalState) String() string { return string(r) }

func ToWorkflowProposalState(v string) *WorkflowProposalState {
	switch strings.ToUpper(v) {
	case WorkflowProposalStateDraft.String():
		return &WorkflowProposalStateDraft
	case WorkflowProposalStateSubmitted.String():
		return &WorkflowProposalStateSubmitted
	case WorkflowProposalStateApplied.String():
		return &WorkflowProposalStateApplied
	case WorkflowProposalStateRejected.String():
		return &WorkflowProposalStateRejected
	case WorkflowProposalStateSuperseded.String():
		return &WorkflowProposalStateSuperseded
	default:
		return nil
	}
}

func (r WorkflowProposalState) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowProposalState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowProposalState, v)
	}
	*r = WorkflowProposalState(str)
	return nil
}

// WorkflowApprovalSubmissionMode enumerates how approval workflows are triggered for a domain.
type WorkflowApprovalSubmissionMode string

var (
	WorkflowApprovalSubmissionModeManualSubmit WorkflowApprovalSubmissionMode = "MANUAL_SUBMIT"
	WorkflowApprovalSubmissionModeAutoSubmit   WorkflowApprovalSubmissionMode = "AUTO_SUBMIT"
)

var WorkflowApprovalSubmissionModes = []string{
	string(WorkflowApprovalSubmissionModeManualSubmit),
	string(WorkflowApprovalSubmissionModeAutoSubmit),
}

func (WorkflowApprovalSubmissionMode) Values() (vals []string) {
	return WorkflowApprovalSubmissionModes
}

func (r WorkflowApprovalSubmissionMode) String() string { return string(r) }

func ToWorkflowApprovalSubmissionMode(v string) *WorkflowApprovalSubmissionMode {
	switch strings.ToUpper(v) {
	case WorkflowApprovalSubmissionModeManualSubmit.String():
		return &WorkflowApprovalSubmissionModeManualSubmit
	case WorkflowApprovalSubmissionModeAutoSubmit.String():
		return &WorkflowApprovalSubmissionModeAutoSubmit
	default:
		return nil
	}
}

func (r WorkflowApprovalSubmissionMode) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowApprovalSubmissionMode) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowApprovalSubmissionMode, v)
	}
	*r = WorkflowApprovalSubmissionMode(str)
	return nil
}

// WorkflowTargetType enumerates assignment target types.
type WorkflowTargetType string

var (
	WorkflowTargetTypeUser     WorkflowTargetType = "USER"
	WorkflowTargetTypeGroup    WorkflowTargetType = "GROUP"
	WorkflowTargetTypeRole     WorkflowTargetType = "ROLE"
	WorkflowTargetTypeResolver WorkflowTargetType = "RESOLVER"
)

var WorkflowTargetTypes = []string{
	string(WorkflowTargetTypeUser),
	string(WorkflowTargetTypeGroup),
	string(WorkflowTargetTypeRole),
	string(WorkflowTargetTypeResolver),
}

func (WorkflowTargetType) Values() (vals []string) {
	return WorkflowTargetTypes
}

func (r WorkflowTargetType) String() string { return string(r) }

func ToWorkflowTargetType(v string) *WorkflowTargetType {
	switch strings.ToUpper(v) {
	case WorkflowTargetTypeUser.String():
		return &WorkflowTargetTypeUser
	case WorkflowTargetTypeGroup.String():
		return &WorkflowTargetTypeGroup
	case WorkflowTargetTypeRole.String():
		return &WorkflowTargetTypeRole
	case WorkflowTargetTypeResolver.String():
		return &WorkflowTargetTypeResolver
	default:
		return nil
	}
}

func (r WorkflowTargetType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowTargetType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowTargetType, v)
	}
	*r = WorkflowTargetType(str)
	return nil
}

// WorkflowObjectType is auto-generated in workflow_object_type.go
// The enum values are dynamically generated based on entities with ApprovalRequiredMixin.
// See internal/ent/generate/templates/ent/workflow_object_type_enum.tmpl

// WorkflowActionType enumerates workflow action types.
type WorkflowActionType string

var (
	WorkflowActionTypeApproval         WorkflowActionType = "REQUEST_APPROVAL"
	WorkflowActionTypeNotification     WorkflowActionType = "NOTIFY"
	WorkflowActionTypeWebhook          WorkflowActionType = "WEBHOOK"
	WorkflowActionTypeFieldUpdate      WorkflowActionType = "UPDATE_FIELD"
	WorkflowActionTypeIntegration      WorkflowActionType = "INTEGRATION"
	WorkflowActionTypeReassignApproval WorkflowActionType = "REASSIGN_APPROVAL"
	WorkflowActionTypeSendEmail        WorkflowActionType = "SEND_EMAIL"
	WorkflowActionTypeCreateObject     WorkflowActionType = "CREATE_OBJECT"
	WorkflowActionTypeReview           WorkflowActionType = "REQUEST_REVIEW"
)

var WorkflowActionTypes = []string{
	string(WorkflowActionTypeApproval),
	string(WorkflowActionTypeNotification),
	string(WorkflowActionTypeWebhook),
	string(WorkflowActionTypeFieldUpdate),
	string(WorkflowActionTypeIntegration),
	string(WorkflowActionTypeReassignApproval),
	string(WorkflowActionTypeSendEmail),
	string(WorkflowActionTypeReview),
	string(WorkflowActionTypeCreateObject),
}

func (WorkflowActionType) Values() (vals []string) {
	return WorkflowActionTypes
}

func (r WorkflowActionType) String() string { return string(r) }

func ToWorkflowActionType(v string) *WorkflowActionType {
	switch strings.ToUpper(v) {
	case WorkflowActionTypeApproval.String():
		return &WorkflowActionTypeApproval
	case WorkflowActionTypeNotification.String():
		return &WorkflowActionTypeNotification
	case WorkflowActionTypeReview.String():
		return &WorkflowActionTypeReview
	case WorkflowActionTypeCreateObject.String():
		return &WorkflowActionTypeCreateObject
	case WorkflowActionTypeWebhook.String():
		return &WorkflowActionTypeWebhook
	case WorkflowActionTypeFieldUpdate.String():
		return &WorkflowActionTypeFieldUpdate
	case WorkflowActionTypeIntegration.String():
		return &WorkflowActionTypeIntegration
	case WorkflowActionTypeReassignApproval.String():
		return &WorkflowActionTypeReassignApproval
	case WorkflowActionTypeSendEmail.String():
		return &WorkflowActionTypeSendEmail
	default:
		return nil
	}
}

func (r WorkflowActionType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowActionType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowActionType, v)
	}
	*r = WorkflowActionType(str)
	return nil
}

// WorkflowEventType enumerates event types.
type WorkflowEventType string

var (
	WorkflowEventTypeAction                WorkflowEventType = "ACTION"
	WorkflowEventTypeTrigger               WorkflowEventType = "TRIGGER"
	WorkflowEventTypeDecision              WorkflowEventType = "DECISION"
	WorkflowEventTypeInstanceTriggered     WorkflowEventType = "WORKFLOW_TRIGGERED"
	WorkflowEventTypeActionStarted         WorkflowEventType = "ACTION_STARTED"
	WorkflowEventTypeActionCompleted       WorkflowEventType = "ACTION_COMPLETED"
	WorkflowEventTypeActionFailed          WorkflowEventType = "ACTION_FAILED"
	WorkflowEventTypeActionSkipped         WorkflowEventType = "ACTION_SKIPPED"
	WorkflowEventTypeConditionEvaluated    WorkflowEventType = "CONDITION_EVALUATED"
	WorkflowEventTypeAssignmentCreated     WorkflowEventType = "ASSIGNMENT_CREATED"
	WorkflowEventTypeAssignmentResolved    WorkflowEventType = "ASSIGNMENT_COMPLETED"
	WorkflowEventTypeAssignmentInvalidated WorkflowEventType = "ASSIGNMENT_INVALIDATED"
	WorkflowEventTypeInstancePaused        WorkflowEventType = "INSTANCE_PAUSED"
	WorkflowEventTypeInstanceResumed       WorkflowEventType = "INSTANCE_RESUMED"
	WorkflowEventTypeInstanceCompleted     WorkflowEventType = "WORKFLOW_COMPLETED"
	WorkflowEventTypeEmitFailed            WorkflowEventType = "EMIT_FAILED"
	WorkflowEventTypeEmitRecovered         WorkflowEventType = "EMIT_RECOVERED"
	WorkflowEventTypeEmitFailedTerminal    WorkflowEventType = "EMIT_FAILED_TERMINAL"
)

var WorkflowEventTypes = []string{
	string(WorkflowEventTypeAction),
	string(WorkflowEventTypeTrigger),
	string(WorkflowEventTypeDecision),
	string(WorkflowEventTypeInstanceTriggered),
	string(WorkflowEventTypeActionStarted),
	string(WorkflowEventTypeActionCompleted),
	string(WorkflowEventTypeActionFailed),
	string(WorkflowEventTypeActionSkipped),
	string(WorkflowEventTypeConditionEvaluated),
	string(WorkflowEventTypeAssignmentCreated),
	string(WorkflowEventTypeAssignmentResolved),
	string(WorkflowEventTypeAssignmentInvalidated),
	string(WorkflowEventTypeInstancePaused),
	string(WorkflowEventTypeInstanceResumed),
	string(WorkflowEventTypeInstanceCompleted),
	string(WorkflowEventTypeEmitFailed),
	string(WorkflowEventTypeEmitRecovered),
	string(WorkflowEventTypeEmitFailedTerminal),
}

func (WorkflowEventType) Values() (vals []string) {
	return WorkflowEventTypes
}

func (r WorkflowEventType) String() string { return string(r) }

func ToWorkflowEventType(v string) *WorkflowEventType {
	switch strings.ToUpper(v) {
	case WorkflowEventTypeAction.String():
		return &WorkflowEventTypeAction
	case WorkflowEventTypeTrigger.String():
		return &WorkflowEventTypeTrigger
	case WorkflowEventTypeDecision.String():
		return &WorkflowEventTypeDecision
	case WorkflowEventTypeInstanceTriggered.String():
		return &WorkflowEventTypeInstanceTriggered
	case WorkflowEventTypeActionStarted.String():
		return &WorkflowEventTypeActionStarted
	case WorkflowEventTypeActionCompleted.String():
		return &WorkflowEventTypeActionCompleted
	case WorkflowEventTypeActionFailed.String():
		return &WorkflowEventTypeActionFailed
	case WorkflowEventTypeActionSkipped.String():
		return &WorkflowEventTypeActionSkipped
	case WorkflowEventTypeConditionEvaluated.String():
		return &WorkflowEventTypeConditionEvaluated
	case WorkflowEventTypeAssignmentCreated.String():
		return &WorkflowEventTypeAssignmentCreated
	case WorkflowEventTypeAssignmentResolved.String():
		return &WorkflowEventTypeAssignmentResolved
	case WorkflowEventTypeAssignmentInvalidated.String():
		return &WorkflowEventTypeAssignmentInvalidated
	case WorkflowEventTypeInstancePaused.String():
		return &WorkflowEventTypeInstancePaused
	case WorkflowEventTypeInstanceResumed.String():
		return &WorkflowEventTypeInstanceResumed
	case WorkflowEventTypeInstanceCompleted.String():
		return &WorkflowEventTypeInstanceCompleted
	case WorkflowEventTypeEmitFailed.String():
		return &WorkflowEventTypeEmitFailed
	case WorkflowEventTypeEmitRecovered.String():
		return &WorkflowEventTypeEmitRecovered
	case WorkflowEventTypeEmitFailedTerminal.String():
		return &WorkflowEventTypeEmitFailedTerminal
	default:
		return nil
	}
}

func (r WorkflowEventType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowEventType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowEventType, v)
	}
	*r = WorkflowEventType(str)
	return nil
}

// WorkflowApprovalTiming enumerates when approvals should block changes.
type WorkflowApprovalTiming string

var (
	WorkflowApprovalTimingPreCommit  WorkflowApprovalTiming = "PRE_COMMIT"
	WorkflowApprovalTimingPostCommit WorkflowApprovalTiming = "POST_COMMIT"
)

var WorkflowApprovalTimings = []string{
	string(WorkflowApprovalTimingPreCommit),
	string(WorkflowApprovalTimingPostCommit),
}

func (WorkflowApprovalTiming) Values() (vals []string) {
	return WorkflowApprovalTimings
}

func (r WorkflowApprovalTiming) String() string { return string(r) }

func ToWorkflowApprovalTiming(v string) *WorkflowApprovalTiming {
	switch strings.ToUpper(v) {
	case WorkflowApprovalTimingPreCommit.String():
		return &WorkflowApprovalTimingPreCommit
	case WorkflowApprovalTimingPostCommit.String():
		return &WorkflowApprovalTimingPostCommit
	default:
		return nil
	}
}

func (r WorkflowApprovalTiming) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowApprovalTiming) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowApprovalTiming, v)
	}
	*r = WorkflowApprovalTiming(str)
	return nil
}
