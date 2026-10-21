package enums

import "io"

// DomainPurpose is a custom type representing the various states of DomainPurpose.
type DomainPurpose string

var (
	// DomainPurposeTrustcenter indicates the trustcenter.
	DomainPurposeTrustcenter DomainPurpose = "TRUSTCENTER"
	// DomainPurposeAssessment indicates the assessment.
	DomainPurposeAssessment DomainPurpose = "ASSESSMENT"
	// DomainPurposeInvalid is used when an unknown or unsupported value is provided.
	DomainPurposeInvalid DomainPurpose = "DOMAINPURPOSE_INVALID"
)

var domainPurposeValues = []DomainPurpose{
	DomainPurposeTrustcenter,
	DomainPurposeAssessment,
}

// Values returns a slice of strings representing all valid DomainPurpose values.
func (DomainPurpose) Values() []string { return stringValues(domainPurposeValues) }

// String returns the string representation of the DomainPurpose value.
func (r DomainPurpose) String() string { return string(r) }

// ToDomainPurpose converts a string to its corresponding DomainPurpose enum value.
func ToDomainPurpose(r string) *DomainPurpose { return parse(r, domainPurposeValues, &DomainPurposeInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r DomainPurpose) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *DomainPurpose) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
