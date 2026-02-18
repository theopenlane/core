package enums

import "io"

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

var programStatusValues = []ProgramStatus{
	ProgramStatusNotStarted, ProgramStatusInProgress, ProgramStatusActionRequired,
	ProgramStatusReadyForAuditor, ProgramStatusCompleted, ProgramStatusArchived,
}

// Values returns a slice of strings representing all valid ProgramStatus values.
func (ProgramStatus) Values() []string { return stringValues(programStatusValues) }

// String returns the string representation of the ProgramStatus value.
func (r ProgramStatus) String() string { return string(r) }

// ToProgramStatus converts a string to its corresponding ProgramStatus enum value.
func ToProgramStatus(r string) *ProgramStatus {
	return parse(r, programStatusValues, &ProgramStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ProgramStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ProgramStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
