package enums

import "io"

// WorkflowKind enumerates workflow kinds.
type WorkflowKind string

var (
	WorkflowKindApproval     WorkflowKind = "APPROVAL"
	WorkflowKindLifecycle    WorkflowKind = "LIFECYCLE"
	WorkflowKindNotification WorkflowKind = "NOTIFICATION"
)

var workflowKindValues = []WorkflowKind{WorkflowKindApproval, WorkflowKindLifecycle, WorkflowKindNotification}

// WorkflowKinds lists all valid workflow kinds as strings.
var WorkflowKinds = stringValues(workflowKindValues)

func (WorkflowKind) Values() []string    { return WorkflowKinds }
func (r WorkflowKind) String() string    { return string(r) }
func ToWorkflowKind(v string) *WorkflowKind { return parse(v, workflowKindValues, nil) }
func (r WorkflowKind) MarshalGQL(w io.Writer)   { marshalGQL(r, w) }
func (r *WorkflowKind) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// WorkflowInstanceState enumerates instance states.
type WorkflowInstanceState string

var (
	WorkflowInstanceStateRunning   WorkflowInstanceState = "RUNNING"
	WorkflowInstanceStateCompleted WorkflowInstanceState = "COMPLETED"
	WorkflowInstanceStateFailed    WorkflowInstanceState = "FAILED"
	WorkflowInstanceStatePaused    WorkflowInstanceState = "PAUSED"
)

var workflowInstanceStateValues = []WorkflowInstanceState{
	WorkflowInstanceStateRunning, WorkflowInstanceStateCompleted, WorkflowInstanceStateFailed, WorkflowInstanceStatePaused,
}

// WorkflowInstanceStates lists all valid workflow instance states as strings.
var WorkflowInstanceStates = stringValues(workflowInstanceStateValues)

func (WorkflowInstanceState) Values() []string       { return WorkflowInstanceStates }
func (r WorkflowInstanceState) String() string        { return string(r) }
func ToWorkflowInstanceState(v string) *WorkflowInstanceState { return parse(v, workflowInstanceStateValues, nil) }
func (r WorkflowInstanceState) MarshalGQL(w io.Writer)        { marshalGQL(r, w) }
func (r *WorkflowInstanceState) UnmarshalGQL(v any) error     { return unmarshalGQL(r, v) }

// WorkflowAssignmentStatus enumerates assignment statuses.
type WorkflowAssignmentStatus string

var (
	WorkflowAssignmentStatusPending          WorkflowAssignmentStatus = "PENDING"
	WorkflowAssignmentStatusApproved         WorkflowAssignmentStatus = "APPROVED"
	WorkflowAssignmentStatusRejected         WorkflowAssignmentStatus = "REJECTED"
	WorkflowAssignmentStatusChangesRequested WorkflowAssignmentStatus = "CHANGES_REQUESTED"
)

var workflowAssignmentStatusValues = []WorkflowAssignmentStatus{
	WorkflowAssignmentStatusPending, WorkflowAssignmentStatusApproved,
	WorkflowAssignmentStatusRejected, WorkflowAssignmentStatusChangesRequested,
}

// WorkflowAssignmentStatuses lists all valid workflow assignment statuses as strings.
var WorkflowAssignmentStatuses = stringValues(workflowAssignmentStatusValues)

func (WorkflowAssignmentStatus) Values() []string       { return WorkflowAssignmentStatuses }
func (r WorkflowAssignmentStatus) String() string        { return string(r) }
func ToWorkflowAssignmentStatus(v string) *WorkflowAssignmentStatus { return parse(v, workflowAssignmentStatusValues, nil) }
func (r WorkflowAssignmentStatus) MarshalGQL(w io.Writer)           { marshalGQL(r, w) }
func (r *WorkflowAssignmentStatus) UnmarshalGQL(v any) error        { return unmarshalGQL(r, v) }

// WorkflowProposalState enumerates proposal lifecycle states.
type WorkflowProposalState string

var (
	WorkflowProposalStateDraft      WorkflowProposalState = "DRAFT"
	WorkflowProposalStateSubmitted  WorkflowProposalState = "SUBMITTED"
	WorkflowProposalStateApplied    WorkflowProposalState = "APPLIED"
	WorkflowProposalStateRejected   WorkflowProposalState = "REJECTED"
	WorkflowProposalStateSuperseded WorkflowProposalState = "SUPERSEDED"
)

var workflowProposalStateValues = []WorkflowProposalState{
	WorkflowProposalStateDraft, WorkflowProposalStateSubmitted, WorkflowProposalStateApplied,
	WorkflowProposalStateRejected, WorkflowProposalStateSuperseded,
}

// WorkflowProposalStates lists all valid workflow proposal states as strings.
var WorkflowProposalStates = stringValues(workflowProposalStateValues)

func (WorkflowProposalState) Values() []string       { return WorkflowProposalStates }
func (r WorkflowProposalState) String() string        { return string(r) }
func ToWorkflowProposalState(v string) *WorkflowProposalState { return parse(v, workflowProposalStateValues, nil) }
func (r WorkflowProposalState) MarshalGQL(w io.Writer)        { marshalGQL(r, w) }
func (r *WorkflowProposalState) UnmarshalGQL(v any) error     { return unmarshalGQL(r, v) }

// WorkflowApprovalSubmissionMode enumerates how approval workflows are triggered for a domain.
type WorkflowApprovalSubmissionMode string

var (
	WorkflowApprovalSubmissionModeManualSubmit WorkflowApprovalSubmissionMode = "MANUAL_SUBMIT"
	WorkflowApprovalSubmissionModeAutoSubmit   WorkflowApprovalSubmissionMode = "AUTO_SUBMIT"
)

var workflowApprovalSubmissionModeValues = []WorkflowApprovalSubmissionMode{
	WorkflowApprovalSubmissionModeManualSubmit, WorkflowApprovalSubmissionModeAutoSubmit,
}

// WorkflowApprovalSubmissionModes lists all valid workflow approval submission modes as strings.
var WorkflowApprovalSubmissionModes = stringValues(workflowApprovalSubmissionModeValues)

func (WorkflowApprovalSubmissionMode) Values() []string       { return WorkflowApprovalSubmissionModes }
func (r WorkflowApprovalSubmissionMode) String() string        { return string(r) }
func ToWorkflowApprovalSubmissionMode(v string) *WorkflowApprovalSubmissionMode { return parse(v, workflowApprovalSubmissionModeValues, nil) }
func (r WorkflowApprovalSubmissionMode) MarshalGQL(w io.Writer)                { marshalGQL(r, w) }
func (r *WorkflowApprovalSubmissionMode) UnmarshalGQL(v any) error             { return unmarshalGQL(r, v) }

// WorkflowTargetType enumerates assignment target types.
type WorkflowTargetType string

var (
	WorkflowTargetTypeUser     WorkflowTargetType = "USER"
	WorkflowTargetTypeGroup    WorkflowTargetType = "GROUP"
	WorkflowTargetTypeRole     WorkflowTargetType = "ROLE"
	WorkflowTargetTypeResolver WorkflowTargetType = "RESOLVER"
)

var workflowTargetTypeValues = []WorkflowTargetType{
	WorkflowTargetTypeUser, WorkflowTargetTypeGroup, WorkflowTargetTypeRole, WorkflowTargetTypeResolver,
}

// WorkflowTargetTypes lists all valid workflow target types as strings.
var WorkflowTargetTypes = stringValues(workflowTargetTypeValues)

func (WorkflowTargetType) Values() []string       { return WorkflowTargetTypes }
func (r WorkflowTargetType) String() string        { return string(r) }
func ToWorkflowTargetType(v string) *WorkflowTargetType { return parse(v, workflowTargetTypeValues, nil) }
func (r WorkflowTargetType) MarshalGQL(w io.Writer)     { marshalGQL(r, w) }
func (r *WorkflowTargetType) UnmarshalGQL(v any) error   { return unmarshalGQL(r, v) }

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

var workflowActionTypeValues = []WorkflowActionType{
	WorkflowActionTypeApproval, WorkflowActionTypeNotification, WorkflowActionTypeWebhook,
	WorkflowActionTypeFieldUpdate, WorkflowActionTypeIntegration, WorkflowActionTypeReassignApproval,
	WorkflowActionTypeSendEmail, WorkflowActionTypeCreateObject, WorkflowActionTypeReview,
}

// WorkflowActionTypes lists all valid workflow action types as strings.
var WorkflowActionTypes = stringValues(workflowActionTypeValues)

func (WorkflowActionType) Values() []string       { return WorkflowActionTypes }
func (r WorkflowActionType) String() string        { return string(r) }
func ToWorkflowActionType(v string) *WorkflowActionType { return parse(v, workflowActionTypeValues, nil) }
func (r WorkflowActionType) MarshalGQL(w io.Writer)     { marshalGQL(r, w) }
func (r *WorkflowActionType) UnmarshalGQL(v any) error   { return unmarshalGQL(r, v) }

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

var workflowEventTypeValues = []WorkflowEventType{
	WorkflowEventTypeAction, WorkflowEventTypeTrigger, WorkflowEventTypeDecision,
	WorkflowEventTypeInstanceTriggered, WorkflowEventTypeActionStarted, WorkflowEventTypeActionCompleted,
	WorkflowEventTypeActionFailed, WorkflowEventTypeActionSkipped, WorkflowEventTypeConditionEvaluated,
	WorkflowEventTypeAssignmentCreated, WorkflowEventTypeAssignmentResolved, WorkflowEventTypeAssignmentInvalidated,
	WorkflowEventTypeInstancePaused, WorkflowEventTypeInstanceResumed, WorkflowEventTypeInstanceCompleted,
	WorkflowEventTypeEmitFailed, WorkflowEventTypeEmitRecovered, WorkflowEventTypeEmitFailedTerminal,
}

// WorkflowEventTypes lists all valid workflow event types as strings.
var WorkflowEventTypes = stringValues(workflowEventTypeValues)

func (WorkflowEventType) Values() []string       { return WorkflowEventTypes }
func (r WorkflowEventType) String() string        { return string(r) }
func ToWorkflowEventType(v string) *WorkflowEventType { return parse(v, workflowEventTypeValues, nil) }
func (r WorkflowEventType) MarshalGQL(w io.Writer)    { marshalGQL(r, w) }
func (r *WorkflowEventType) UnmarshalGQL(v any) error  { return unmarshalGQL(r, v) }

// WorkflowApprovalTiming enumerates when approvals should block changes.
type WorkflowApprovalTiming string

var (
	WorkflowApprovalTimingPreCommit  WorkflowApprovalTiming = "PRE_COMMIT"
	WorkflowApprovalTimingPostCommit WorkflowApprovalTiming = "POST_COMMIT"
)

var workflowApprovalTimingValues = []WorkflowApprovalTiming{
	WorkflowApprovalTimingPreCommit, WorkflowApprovalTimingPostCommit,
}

// WorkflowApprovalTimings lists all valid workflow approval timings as strings.
var WorkflowApprovalTimings = stringValues(workflowApprovalTimingValues)

func (WorkflowApprovalTiming) Values() []string       { return WorkflowApprovalTimings }
func (r WorkflowApprovalTiming) String() string        { return string(r) }
func ToWorkflowApprovalTiming(v string) *WorkflowApprovalTiming { return parse(v, workflowApprovalTimingValues, nil) }
func (r WorkflowApprovalTiming) MarshalGQL(w io.Writer)         { marshalGQL(r, w) }
func (r *WorkflowApprovalTiming) UnmarshalGQL(v any) error       { return unmarshalGQL(r, v) }
