package enums

import (
	"fmt"
	"io"
	"strings"
)

// ImpersonationType is a custom type representing the various states of ImpersonationType.
type ImpersonationType string

var (
	// ImpersonationTypeSupport indicates the support.
	ImpersonationTypeSupport ImpersonationType = "SUPPORT"
	// ImpersonationTypeAdmin indicates the admin.
	ImpersonationTypeAdmin ImpersonationType = "ADMIN"
	// ImpersonationTypeJob indicates the job.
	ImpersonationTypeJob ImpersonationType = "JOB"
	// ImpersonationTypeInvalid is used when an unknown or unsupported value is provided.
	ImpersonationTypeInvalid ImpersonationType = "IMPERSONATIONTYPE_INVALID"
)

// Values returns a slice of strings representing all valid ImpersonationType values.
func (ImpersonationType) Values() []string {
	return []string{
		string(ImpersonationTypeSupport),
		string(ImpersonationTypeAdmin),
		string(ImpersonationTypeJob),
	}
}

// String returns the string representation of the ImpersonationType value.
func (r ImpersonationType) String() string {
	return string(r)
}

// ToImpersonationType converts a string to its corresponding ImpersonationType enum value.
func ToImpersonationType(r string) *ImpersonationType {
	switch strings.ToUpper(r) {
	case ImpersonationTypeSupport.String():
		return &ImpersonationTypeSupport
	case ImpersonationTypeAdmin.String():
		return &ImpersonationTypeAdmin
	case ImpersonationTypeJob.String():
		return &ImpersonationTypeJob
	default:
		return &ImpersonationTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ImpersonationType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ImpersonationType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ImpersonationType, got: %T", v) //nolint:err113
	}

	*r = ImpersonationType(str)

	return nil
}
