package enums

import (
	"fmt"
	"io"
	"strings"
)

type ProgramStatus string

var (
	ProgramStatusNotStarted      ProgramStatus = "NOT_STARTED"
	ProgramStatusInProgress      ProgramStatus = "IN_PROGRESS"
	ProgramStatusActionRequired  ProgramStatus = "ACTION_REQUIRED"
	ProgramStatusReadyForAuditor ProgramStatus = "READY_FOR_AUDITOR"
	ProgramStatusCompleted       ProgramStatus = "COMPLETED"
	ProgramStatusInvalid         ProgramStatus = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the ProgramStatus enum.
// Possible default values are "OPEN", "IN_PROGRESS", "IN_REVIEW", "COMPLETED", and "ACTION_REQUIRED".
func (ProgramStatus) Values() (kinds []string) {
	for _, s := range []ProgramStatus{ProgramStatusNotStarted, ProgramStatusInProgress, ProgramStatusReadyForAuditor, ProgramStatusCompleted, ProgramStatusActionRequired} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the ProgramStatus as a string
func (r ProgramStatus) String() string {
	return string(r)
}

// ToProgramStatus returns the program status enum based on string input
func ToProgramStatus(r string) *ProgramStatus {
	switch r := strings.ToUpper(r); r {
	case ProgramStatusNotStarted.String():
		return &ProgramStatusNotStarted
	case ProgramStatusInProgress.String():
		return &ProgramStatusInProgress
	case ProgramStatusReadyForAuditor.String():
		return &ProgramStatusReadyForAuditor
	case ProgramStatusCompleted.String():
		return &ProgramStatusCompleted
	case ProgramStatusActionRequired.String():
		return &ProgramStatusActionRequired
	default:
		return &ProgramStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ProgramStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ProgramStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ProgramStatus, got: %T", v) //nolint:err113
	}

	*r = ProgramStatus(str)

	return nil
}
