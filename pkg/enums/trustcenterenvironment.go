package enums

import (
	"fmt"
	"io"
	"strings"
)

// TrustCenterEnvironment is a custom type representing the various states of TrustCenterEnvironment.
type TrustCenterEnvironment string

var (
	// TrustCenterEnvironmentLive indicates the live.
	TrustCenterEnvironmentLive TrustCenterEnvironment = "LIVE"
	// TrustCenterEnvironmentPreview indicates the preview.
	TrustCenterEnvironmentPreview TrustCenterEnvironment = "PREVIEW"
	// TrustCenterEnvironmentInvalid is used when an unknown or unsupported value is provided.
	TrustCenterEnvironmentInvalid TrustCenterEnvironment = "TRUSTCENTERENVIRONMENT_INVALID"
)

// Values returns a slice of strings representing all valid TrustCenterEnvironment values.
func (TrustCenterEnvironment) Values() []string {
	return []string{
		string(TrustCenterEnvironmentLive),
		string(TrustCenterEnvironmentPreview),
	}
}

// String returns the string representation of the TrustCenterEnvironment value.
func (r TrustCenterEnvironment) String() string {
	return string(r)
}

// ToTrustCenterEnvironment converts a string to its corresponding TrustCenterEnvironment enum value.
func ToTrustCenterEnvironment(r string) *TrustCenterEnvironment {
	switch strings.ToUpper(r) {
	case TrustCenterEnvironmentLive.String():
		return &TrustCenterEnvironmentLive
	case TrustCenterEnvironmentPreview.String():
		return &TrustCenterEnvironmentPreview
	default:
		return &TrustCenterEnvironmentInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TrustCenterEnvironment) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TrustCenterEnvironment) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TrustCenterEnvironment, got: %T", v) //nolint:err113
	}

	*r = TrustCenterEnvironment(str)

	return nil
}
