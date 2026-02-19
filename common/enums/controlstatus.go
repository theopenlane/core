package enums

import (
	"io"
	"strings"
)

// ControlStatus is a custom type for control status.
type ControlStatus string

var (
	// ControlStatusNotImplemented indicates that this control has not yet been worked on, this is the default value
	ControlStatusNotImplemented ControlStatus = "NOT_IMPLEMENTED"
	// ControlStatusPreparing indicates the control is being prepared
	ControlStatusPreparing ControlStatus = "PREPARING"
	// ControlStatusNeedsApproval indicates the control needs to be approved before it is available
	ControlStatusNeedsApproval ControlStatus = "NEEDS_APPROVAL"
	// ControlStatusChangesRequested indicates the control was rejected and needs some changes to be approved
	ControlStatusChangesRequested ControlStatus = "CHANGES_REQUESTED"
	// ControlStatusApproved indicates the control is approved
	ControlStatusApproved ControlStatus = "APPROVED"
	// ControlStatusArchived indicates the control is now archived
	ControlStatusArchived ControlStatus = "ARCHIVED"
	// ControlStatusNotApplicable indicates that this control does not apply to the organization
	ControlStatusNotApplicable ControlStatus = "NOT_APPLICABLE"
	// ControlStatusInvalid indicates the control is invalid or unknown
	ControlStatusInvalid ControlStatus = "CONTROL_STATUS_INVALID"
)

var controlStatusValues = []ControlStatus{
	ControlStatusPreparing,
	ControlStatusNeedsApproval,
	ControlStatusChangesRequested,
	ControlStatusApproved,
	ControlStatusArchived,
	ControlStatusNotImplemented,
	ControlStatusNotApplicable,
}

// Values returns a slice of strings that represents all the possible values of the ControlStatus enum.
func (ControlStatus) Values() []string { return stringValues(controlStatusValues) }

// String returns the ControlStatus as a string
func (r ControlStatus) String() string { return string(r) }

// ToControlStatus returns the control status enum based on string input.
// An empty string defaults to ControlStatusNotImplemented.
func ToControlStatus(r string) *ControlStatus {
	if strings.TrimSpace(r) == "" {
		return &ControlStatusNotImplemented
	}

	return parse(r, controlStatusValues, &ControlStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ControlStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ControlStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
