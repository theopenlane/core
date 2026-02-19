package enums

import "io"

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

var impersonationTypeValues = []ImpersonationType{ImpersonationTypeSupport, ImpersonationTypeAdmin, ImpersonationTypeJob}

// Values returns a slice of strings representing all valid ImpersonationType values.
func (ImpersonationType) Values() []string { return stringValues(impersonationTypeValues) }

// String returns the string representation of the ImpersonationType value.
func (r ImpersonationType) String() string { return string(r) }

// ToImpersonationType converts a string to its corresponding ImpersonationType enum value.
func ToImpersonationType(r string) *ImpersonationType {
	return parse(r, impersonationTypeValues, &ImpersonationTypeInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ImpersonationType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ImpersonationType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
