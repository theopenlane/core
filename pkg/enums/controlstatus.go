package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the ControlType enum.
// Possible default values are "NOT_IMPLEMENTED", "PREPARING", "NEEDS APPROVAL", "CHANGES REQUESTED",
// "APPROVED" and "ARCHIVED".
func (ControlStatus) Values() (kinds []string) {
	for _, s := range []ControlStatus{
		ControlStatusPreparing,
		ControlStatusNeedsApproval,
		ControlStatusChangesRequested,
		ControlStatusApproved,
		ControlStatusArchived,
		ControlStatusNotImplemented,
		ControlStatusNotApplicable,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the ControlStatus as a string
func (r ControlStatus) String() string {
	return string(r)
}

// ToControlStatus returns the control type enum based on string input
func ToControlStatus(r string) *ControlStatus {
	switch r := strings.ToUpper(r); r {
	case "", ControlStatusNotImplemented.String():
		return &ControlStatusNotImplemented
	case ControlStatusPreparing.String():
		return &ControlStatusPreparing
	case ControlStatusNeedsApproval.String():
		return &ControlStatusNeedsApproval
	case ControlStatusChangesRequested.String():
		return &ControlStatusChangesRequested
	case ControlStatusApproved.String():
		return &ControlStatusApproved
	case ControlStatusNotApplicable.String():
		return &ControlStatusNotApplicable
	case ControlStatusArchived.String():
		return &ControlStatusArchived
	default:
		return &ControlStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ControlStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ControlStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ControlStatus, got: %T", v) //nolint:err113
	}

	*r = ControlStatus(str)

	return nil
}
