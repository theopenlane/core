package enums

import (
	"fmt"
	"io"
	"strings"
)

// TrustCenterNDARequestStatus is a custom type for NDA request status
type TrustCenterNDARequestStatus string

var (
	// TrustCenterNDARequestStatusRequested indicates the NDA has been requested
	TrustCenterNDARequestStatusRequested TrustCenterNDARequestStatus = "REQUESTED"
	// TrustCenterNDARequestStatusNeedsApproval indicates the NDA request needs approval
	TrustCenterNDARequestStatusNeedsApproval TrustCenterNDARequestStatus = "NEEDS_APPROVAL"
	// TrustCenterNDARequestStatusApproved indicates the NDA request has been approved
	TrustCenterNDARequestStatusApproved TrustCenterNDARequestStatus = "APPROVED"
	// TrustCenterNDARequestStatusSigned indicates the NDA has been signed
	TrustCenterNDARequestStatusSigned TrustCenterNDARequestStatus = "SIGNED"
	// TrustCenterNDARequestStatusDeclined indicates the NDA request has been declined
	TrustCenterNDARequestStatusDeclined TrustCenterNDARequestStatus = "DECLINED"
	// TrustCenterNDARequestStatusInvalid indicates the NDA request status is invalid
	TrustCenterNDARequestStatusInvalid TrustCenterNDARequestStatus = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the TrustCenterNDARequestStatus enum
func (TrustCenterNDARequestStatus) Values() (kinds []string) {
	for _, s := range []TrustCenterNDARequestStatus{
		TrustCenterNDARequestStatusRequested,
		TrustCenterNDARequestStatusNeedsApproval,
		TrustCenterNDARequestStatusApproved,
		TrustCenterNDARequestStatusSigned,
		TrustCenterNDARequestStatusDeclined,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the NDA request status as a string
func (r TrustCenterNDARequestStatus) String() string {
	return string(r)
}

// ToTrustCenterNDARequestStatus returns the NDA request status enum based on string input
func ToTrustCenterNDARequestStatus(r string) *TrustCenterNDARequestStatus {
	switch r := strings.ToUpper(r); r {
	case TrustCenterNDARequestStatusRequested.String():
		return &TrustCenterNDARequestStatusRequested
	case TrustCenterNDARequestStatusNeedsApproval.String():
		return &TrustCenterNDARequestStatusNeedsApproval
	case TrustCenterNDARequestStatusApproved.String():
		return &TrustCenterNDARequestStatusApproved
	case TrustCenterNDARequestStatusSigned.String():
		return &TrustCenterNDARequestStatusSigned
	case TrustCenterNDARequestStatusDeclined.String():
		return &TrustCenterNDARequestStatusDeclined
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterNDARequestStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterNDARequestStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TrustCenterNDARequestStatus, got: %T", v) //nolint:err113
	}

	*r = TrustCenterNDARequestStatus(str)

	return nil
}
