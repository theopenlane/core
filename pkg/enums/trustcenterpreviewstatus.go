package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings representing all valid TrustCenterPreviewStatus values.
func (TrustCenterPreviewStatus) Values() []string {
	return []string{
		string(TrustCenterPreviewStatusProvisioning),
		string(TrustCenterPreviewStatusReady),
		string(TrustCenterPreviewStatusFailed),
		string(TrustCenterPreviewStatusDeprovisioning),
		string(TrustCenterPreviewStatusNone),
	}
}

// String returns the string representation of the TrustCenterPreviewStatus value.
func (r TrustCenterPreviewStatus) String() string {
	return string(r)
}

// ToTrustCenterPreviewStatus converts a string to its corresponding TrustCenterPreviewStatus enum value.
func ToTrustCenterPreviewStatus(r string) *TrustCenterPreviewStatus {
	switch strings.ToUpper(r) {
	case TrustCenterPreviewStatusProvisioning.String():
		return &TrustCenterPreviewStatusProvisioning
	case TrustCenterPreviewStatusReady.String():
		return &TrustCenterPreviewStatusReady
	case TrustCenterPreviewStatusFailed.String():
		return &TrustCenterPreviewStatusFailed
	case TrustCenterPreviewStatusDeprovisioning.String():
		return &TrustCenterPreviewStatusDeprovisioning
	case TrustCenterPreviewStatusNone.String():
		return &TrustCenterPreviewStatusNone
	default:
		return &TrustCenterPreviewStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TrustCenterPreviewStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TrustCenterPreviewStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TrustCenterPreviewStatus, got: %T", v) //nolint:err113
	}

	*r = TrustCenterPreviewStatus(str)

	return nil
}
