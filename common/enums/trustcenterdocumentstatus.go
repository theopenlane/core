package enums

import "io"

// TrustCenterDocumentVisibility is a custom type for document status
type TrustCenterDocumentVisibility string

var (
	// TrustCenterDocumentVisibilityPubliclyVisible indicates that the document is publicly visible
	TrustCenterDocumentVisibilityPubliclyVisible TrustCenterDocumentVisibility = "PUBLICLY_VISIBLE"
	// TrustCenterDocumentVisibilityProtected indicates that the document is
	// protected and only visible to certain users that have signed ANDAs
	TrustCenterDocumentVisibilityProtected TrustCenterDocumentVisibility = "PROTECTED"
	// TrustCenterDocumentVisibilityNotVisible indicates that the document is not visible to anyone
	TrustCenterDocumentVisibilityNotVisible TrustCenterDocumentVisibility = "NOT_VISIBLE"
	// TrustCenterDocumentVisibilityInvalid indicates that the document status is invalid
	TrustCenterDocumentVisibilityInvalid TrustCenterDocumentVisibility = "DOCUMENT_STATUS_INVALID"
)

var trustCenterDocumentVisibilityValues = []TrustCenterDocumentVisibility{
	TrustCenterDocumentVisibilityPubliclyVisible,
	TrustCenterDocumentVisibilityProtected,
	TrustCenterDocumentVisibilityNotVisible,
}

// Values returns a slice of strings that represents all the possible values of the TrustCenterDocumentVisibility enum.
// Possible default values are "PUBLISHED", "DRAFT", "NEEDS_APPROVAL", and "APPROVED"
func (TrustCenterDocumentVisibility) Values() []string {
	return stringValues(trustCenterDocumentVisibilityValues)
}

// String returns the document status as a string
func (r TrustCenterDocumentVisibility) String() string { return string(r) }

// ToTrustCenterDocumentVisibility returns the document status enum based on string input
func ToTrustCenterDocumentVisibility(r string) *TrustCenterDocumentVisibility {
	return parse(r, trustCenterDocumentVisibilityValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterDocumentVisibility) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterDocumentVisibility) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
