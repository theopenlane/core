package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the Priority enum.
// Possible default values are "LOW", "MEDIUM", "HIGH", and "CRITICAL".
func (Priority) Values() (kinds []string) {
	for _, s := range []Priority{PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the Priority as a string
func (r Priority) String() string {
	return string(r)
}

// ToPriority returns the user status enum based on string input
func ToPriority(r string) *Priority {
	switch r := strings.ToUpper(r); r {
	case PriorityLow.String():
		return &PriorityLow
	case PriorityMedium.String():
		return &PriorityMedium
	case PriorityHigh.String():
		return &PriorityHigh
	case PriorityCritical.String():
		return &PriorityCritical
	default:
		return &PriorityInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Priority) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Priority) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Priority, got: %T", v) //nolint:err113
	}

	*r = Priority(str)

	return nil
}
