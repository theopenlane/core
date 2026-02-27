package enums

import (
	"io"
	"strings"
)

// ControlImplementationStatus is a custom type for OSCAL-aligned control implementation status.
type ControlImplementationStatus string

var (
	// ControlImplementationStatusPlanned indicates the implementation is planned.
	ControlImplementationStatusPlanned ControlImplementationStatus = "PLANNED"
	// ControlImplementationStatusImplemented indicates the implementation is complete.
	ControlImplementationStatusImplemented ControlImplementationStatus = "IMPLEMENTED"
	// ControlImplementationStatusPartiallyImplemented indicates the implementation is partially complete.
	ControlImplementationStatusPartiallyImplemented ControlImplementationStatus = "PARTIALLY_IMPLEMENTED"
	// ControlImplementationStatusInherited indicates implementation is inherited from another system/provider.
	ControlImplementationStatusInherited ControlImplementationStatus = "INHERITED"
	// ControlImplementationStatusNotApplicable indicates the control does not apply in context.
	ControlImplementationStatusNotApplicable ControlImplementationStatus = "NOT_APPLICABLE"
	// ControlImplementationStatusInvalid indicates the status is invalid or unknown.
	ControlImplementationStatusInvalid ControlImplementationStatus = "CONTROL_IMPLEMENTATION_STATUS_INVALID"
)

var controlImplementationStatusValues = []ControlImplementationStatus{
	ControlImplementationStatusPlanned,
	ControlImplementationStatusImplemented,
	ControlImplementationStatusPartiallyImplemented,
	ControlImplementationStatusInherited,
	ControlImplementationStatusNotApplicable,
}

// Values returns all valid ControlImplementationStatus values.
func (ControlImplementationStatus) Values() []string {
	return stringValues(controlImplementationStatusValues)
}

// String returns the string representation of the status.
func (r ControlImplementationStatus) String() string { return string(r) }

// ToControlImplementationStatus parses a string into a ControlImplementationStatus.
// An empty input defaults to planned.
func ToControlImplementationStatus(r string) *ControlImplementationStatus {
	if strings.TrimSpace(r) == "" {
		return &ControlImplementationStatusPlanned
	}

	return parse(r, controlImplementationStatusValues, &ControlImplementationStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ControlImplementationStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ControlImplementationStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
