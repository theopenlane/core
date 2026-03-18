package enums

import "io"

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

var trustCenterNDARequestStatusValues = []TrustCenterNDARequestStatus{
	TrustCenterNDARequestStatusRequested,
	TrustCenterNDARequestStatusNeedsApproval,
	TrustCenterNDARequestStatusApproved,
	TrustCenterNDARequestStatusSigned,
	TrustCenterNDARequestStatusDeclined,
}

// Values returns a slice of strings that represents all the possible values of the TrustCenterNDARequestStatus enum
func (TrustCenterNDARequestStatus) Values() []string {
	return stringValues(trustCenterNDARequestStatusValues)
}

// String returns the NDA request status as a string
func (r TrustCenterNDARequestStatus) String() string { return string(r) }

// ToTrustCenterNDARequestStatus returns the NDA request status enum based on string input
func ToTrustCenterNDARequestStatus(r string) *TrustCenterNDARequestStatus {
	return parse(r, trustCenterNDARequestStatusValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterNDARequestStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterNDARequestStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
