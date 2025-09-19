package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the TrustCenterDocumentVisibility enum.
// Possible default values are "PUBLISHED", "DRAFT", "NEEDS_APPROVAL", and "APPROVED"
func (TrustCenterDocumentVisibility) Values() (kinds []string) {
	for _, s := range []TrustCenterDocumentVisibility{
		TrustCenterDocumentVisibilityPubliclyVisible,
		TrustCenterDocumentVisibilityProtected,
		TrustCenterDocumentVisibilityNotVisible,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the document status as a string
func (r TrustCenterDocumentVisibility) String() string {
	return string(r)
}

// ToTrustCenterDocumentVisibility returns the document status enum based on string input
func ToTrustCenterDocumentVisibility(r string) *TrustCenterDocumentVisibility {
	switch r := strings.ToUpper(r); r {
	case TrustCenterDocumentVisibilityPubliclyVisible.String():
		return &TrustCenterDocumentVisibilityPubliclyVisible
	case TrustCenterDocumentVisibilityProtected.String():
		return &TrustCenterDocumentVisibilityProtected
	case TrustCenterDocumentVisibilityNotVisible.String():
		return &TrustCenterDocumentVisibilityNotVisible
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterDocumentVisibility) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterDocumentVisibility) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TrustCenterDocumentVisibility, got: %T", v) //nolint:err113
	}

	*r = TrustCenterDocumentVisibility(str)

	return nil
}
