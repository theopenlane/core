package enums

import "io"

// DocumentType is a custom type representing the type of a document.
type DocumentType string

var (
	// RootTemplate are templates provided by the system
	RootTemplate DocumentType = "ROOTTEMPLATE"
	// Document are templates from root templates, or scratch, owned by the organization
	Document DocumentType = "DOCUMENT"
	// DocumentTypeInvalid is the default value for the DocumentType enum
	DocumentTypeInvalid DocumentType = "INVALID"
)

var documentTypeValues = []DocumentType{RootTemplate, Document}

// Values returns a slice of strings that represents all the possible values of the DocumentType enum.
// Possible default values are "ROOTTEMPLATE", "DOCUMENT"
func (DocumentType) Values() []string { return stringValues(documentTypeValues) }

// String returns the DocumentType as a string
func (r DocumentType) String() string { return string(r) }

// ToDocumentType returns the user status enum based on string input
func ToDocumentType(r string) *DocumentType {
	return parse(r, documentTypeValues, &DocumentTypeInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r DocumentType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *DocumentType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
