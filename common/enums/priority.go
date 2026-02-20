package enums

import "io"

// Priority represents the priority of a object (e.g. task)
type Priority string

var (
	// PriorityLow represents the low priority
	PriorityLow Priority = "LOW"
	// PriorityMedium represents the medium priority
	PriorityMedium Priority = "MEDIUM"
	// PriorityHigh represents the high priority
	PriorityHigh Priority = "HIGH"
	// PriorityCritical represents the critical priority
	PriorityCritical Priority = "CRITICAL"
	// PriorityInvalid represents an invalid priority
	PriorityInvalid Priority = "INVALID"
)

var priorityValues = []Priority{PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical}

// Values returns a slice of strings that represents all the possible values of the Priority enum.
// Possible default values are "LOW", "MEDIUM", "HIGH", and "CRITICAL".
func (Priority) Values() []string { return stringValues(priorityValues) }

// String returns the Priority as a string
func (r Priority) String() string { return string(r) }

// ToPriority returns the user status enum based on string input
func ToPriority(r string) *Priority { return parse(r, priorityValues, &PriorityInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Priority) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Priority) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
