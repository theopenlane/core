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
	WorkflowAssignmentStatusPending  WorkflowAssignmentStatus = "PENDING"
	WorkflowAssignmentStatusApproved WorkflowAssignmentStatus = "APPROVED"
	WorkflowAssignmentStatusRejected WorkflowAssignmentStatus = "REJECTED"
)

var WorkflowAssignmentStatuses = []string{
	string(WorkflowAssignmentStatusPending),
	string(WorkflowAssignmentStatusApproved),
	string(WorkflowAssignmentStatusRejected),
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

// WorkflowObjectType enumerates supported object types.
type WorkflowObjectType string

var (
	WorkflowObjectTypeControl     WorkflowObjectType = "CONTROL"
	WorkflowObjectTypeTask        WorkflowObjectType = "TASK"
	WorkflowObjectTypePolicy      WorkflowObjectType = "POLICY"
	WorkflowObjectTypeProcedure   WorkflowObjectType = "PROCEDURE"
	WorkflowObjectTypeEvidence    WorkflowObjectType = "EVIDENCE"
	WorkflowObjectTypeAsset       WorkflowObjectType = "ASSET"
	WorkflowObjectTypeRisk        WorkflowObjectType = "RISK"
	WorkflowObjectTypeProgram     WorkflowObjectType = "PROGRAM"
	WorkflowObjectTypeFinding     WorkflowObjectType = "FINDING"
	WorkflowObjectTypeNarrative   WorkflowObjectType = "NARRATIVE"
	WorkflowObjectTypeIntegration WorkflowObjectType = "INTEGRATION"
)

var WorkflowObjectTypes = []string{
	string(WorkflowObjectTypeControl),
	string(WorkflowObjectTypeTask),
	string(WorkflowObjectTypePolicy),
	string(WorkflowObjectTypeProcedure),
	string(WorkflowObjectTypeEvidence),
	string(WorkflowObjectTypeAsset),
	string(WorkflowObjectTypeRisk),
	string(WorkflowObjectTypeProgram),
	string(WorkflowObjectTypeFinding),
	string(WorkflowObjectTypeNarrative),
	string(WorkflowObjectTypeIntegration),
}

func (WorkflowObjectType) Values() (vals []string) {
	return WorkflowObjectTypes
}

func (r WorkflowObjectType) String() string { return string(r) }

func ToWorkflowObjectType(v string) *WorkflowObjectType {
	switch strings.ToUpper(v) {
	case WorkflowObjectTypeControl.String():
		return &WorkflowObjectTypeControl
	case WorkflowObjectTypeTask.String():
		return &WorkflowObjectTypeTask
	case WorkflowObjectTypePolicy.String():
		return &WorkflowObjectTypePolicy
	case WorkflowObjectTypeProcedure.String():
		return &WorkflowObjectTypeProcedure
	case WorkflowObjectTypeEvidence.String():
		return &WorkflowObjectTypeEvidence
	case WorkflowObjectTypeAsset.String():
		return &WorkflowObjectTypeAsset
	case WorkflowObjectTypeRisk.String():
		return &WorkflowObjectTypeRisk
	case WorkflowObjectTypeProgram.String():
		return &WorkflowObjectTypeProgram
	case WorkflowObjectTypeFinding.String():
		return &WorkflowObjectTypeFinding
	case WorkflowObjectTypeNarrative.String():
		return &WorkflowObjectTypeNarrative
	case WorkflowObjectTypeIntegration.String():
		return &WorkflowObjectTypeIntegration
	default:
		return nil
	}
}

func (r WorkflowObjectType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *WorkflowObjectType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowObjectType, v)
	}
	*r = WorkflowObjectType(str)
	return nil
}

// WorkflowEventType enumerates event types.
type WorkflowEventType string

var (
	WorkflowEventTypeAction   WorkflowEventType = "ACTION"
	WorkflowEventTypeTrigger  WorkflowEventType = "TRIGGER"
	WorkflowEventTypeDecision WorkflowEventType = "DECISION"
)

var WorkflowEventTypes = []string{
	string(WorkflowEventTypeAction),
	string(WorkflowEventTypeTrigger),
	string(WorkflowEventTypeDecision),
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
