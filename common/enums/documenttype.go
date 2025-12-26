package enums

import (
	"fmt"
	"io"
	"strings"
)

type DocumentType string

var (
	// RootTemplate are templates provided by the system
	RootTemplate DocumentType = "ROOTTEMPLATE"
	// Document are templates from root templates, or scratch, owned by the organization
	Document DocumentType = "DOCUMENT"
	// DocumentTypeInvalid is the default value for the DocumentType enum
	DocumentTypeInvalid DocumentType = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the DocumentType enum.
// Possible default values are "ROOTTEMPLATE", "DOCUMENT"
func (DocumentType) Values() (kinds []string) {
	for _, s := range []DocumentType{RootTemplate, Document} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the DocumentType as a string
func (r DocumentType) String() string {
	return string(r)
}

// ToDocumentType returns the user status enum based on string input
func ToDocumentType(r string) *DocumentType {
	switch r := strings.ToUpper(r); r {
	case RootTemplate.String():
		return &RootTemplate
	case Document.String():
		return &Document
	default:
		return &DocumentTypeInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r DocumentType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *DocumentType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for DocumentType, got: %T", v) //nolint:err113
	}

	*r = DocumentType(str)

	return nil
}
