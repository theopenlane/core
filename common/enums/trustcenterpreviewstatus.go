package enums

import "io"

// TrustCenterPreviewStatus is a custom type representing the various states of TrustCenterPreviewStatus.
type TrustCenterPreviewStatus string

var (
	// TrustCenterPreviewStatusProvisioning indicates the preview is being provisioned.
	TrustCenterPreviewStatusProvisioning TrustCenterPreviewStatus = "PROVISIONING"
	// TrustCenterPreviewStatusReady indicates the preview is ready.
	TrustCenterPreviewStatusReady TrustCenterPreviewStatus = "READY"
	// TrustCenterPreviewStatusFailed indicates the preview has failed.
	TrustCenterPreviewStatusFailed TrustCenterPreviewStatus = "FAILED"
	// TrustCenterPreviewStatusDeprovisioning indicates the preview is being deprovisioned.
	TrustCenterPreviewStatusDeprovisioning TrustCenterPreviewStatus = "DEPROVISIONING"
	// TrustCenterPreviewStatusNone indicates there is no preview environment
	TrustCenterPreviewStatusNone TrustCenterPreviewStatus = "NONE"
	// TrustCenterPreviewStatusInvalid is used when an unknown or unsupported value is provided.
	TrustCenterPreviewStatusInvalid TrustCenterPreviewStatus = "TRUSTCENTERPREVIEWSTATUS_INVALID"
)

var trustCenterPreviewStatusValues = []TrustCenterPreviewStatus{
	TrustCenterPreviewStatusProvisioning,
	TrustCenterPreviewStatusReady,
	TrustCenterPreviewStatusFailed,
	TrustCenterPreviewStatusDeprovisioning,
	TrustCenterPreviewStatusNone,
}

// Values returns a slice of strings representing all valid TrustCenterPreviewStatus values.
func (TrustCenterPreviewStatus) Values() []string {
	return stringValues(trustCenterPreviewStatusValues)
}

// String returns the string representation of the TrustCenterPreviewStatus value.
func (r TrustCenterPreviewStatus) String() string { return string(r) }

// ToTrustCenterPreviewStatus converts a string to its corresponding TrustCenterPreviewStatus enum value.
func ToTrustCenterPreviewStatus(r string) *TrustCenterPreviewStatus {
	return parse(r, trustCenterPreviewStatusValues, &TrustCenterPreviewStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TrustCenterPreviewStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TrustCenterPreviewStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
