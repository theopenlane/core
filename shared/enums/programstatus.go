package enums

import (
	"fmt"
	"io"
	"strings"
)

// ProgramStatus is a custom type representing the various states of ProgramStatus.
type ProgramStatus string

var (
	// ProgramStatusNotStarted indicates the not started status.
	ProgramStatusNotStarted ProgramStatus = "NOT_STARTED"
	// ProgramStatusInProgress indicates the in progress status.
	ProgramStatusInProgress ProgramStatus = "IN_PROGRESS"
	// ProgramStatusActionRequired indicates an action required.
	ProgramStatusActionRequired ProgramStatus = "ACTION_REQUIRED"
	// ProgramStatusReadyForAuditor indicates the ready for auditor.
	ProgramStatusReadyForAuditor ProgramStatus = "READY_FOR_AUDITOR"
	// ProgramStatusCompleted indicates the completed status.
	ProgramStatusCompleted ProgramStatus = "COMPLETED"
	// ProgramStatusArchived indicates the archived status.
	ProgramStatusArchived ProgramStatus = "ARCHIVED"
	// ProgramStatusInvalid is used when an unknown or unsupported value is provided.
	ProgramStatusInvalid ProgramStatus = "INVALID"
)

// Values returns a slice of strings representing all valid ProgramStatus values.
func (ProgramStatus) Values() []string {
	return []string{
		string(ProgramStatusNotStarted),
		string(ProgramStatusInProgress),
		string(ProgramStatusActionRequired),
		string(ProgramStatusReadyForAuditor),
		string(ProgramStatusCompleted),
		string(ProgramStatusArchived),
	}
}

// String returns the string representation of the ProgramStatus value.
func (r ProgramStatus) String() string {
	return string(r)
}

// ToProgramStatus converts a string to its corresponding ProgramStatus enum value.
func ToProgramStatus(r string) *ProgramStatus {
	switch strings.ToUpper(r) {
	case ProgramStatusNotStarted.String():
		return &ProgramStatusNotStarted
	case ProgramStatusInProgress.String():
		return &ProgramStatusInProgress
	case ProgramStatusActionRequired.String():
		return &ProgramStatusActionRequired
	case ProgramStatusReadyForAuditor.String():
		return &ProgramStatusReadyForAuditor
	case ProgramStatusCompleted.String():
		return &ProgramStatusCompleted
	case ProgramStatusArchived.String():
		return &ProgramStatusArchived
	default:
		return &ProgramStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ProgramStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ProgramStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ProgramStatus, got: %T", v) //nolint:err113
	}

	*r = ProgramStatus(str)

	return nil
}
