package enums

import "io"

// DocumentStatus is a custom type for document status
type DocumentStatus string

var (
	// DocumentPublished indicates that the document is published
	DocumentPublished DocumentStatus = "PUBLISHED"
	// DocumentDraft indicates that the document is in draft status
	DocumentDraft DocumentStatus = "DRAFT"
	// DocumentNeedsApproval indicates that the document needs approval
	DocumentNeedsApproval DocumentStatus = "NEEDS_APPROVAL"
	// DocumentApproved indicates that the document has been approved and is ready to be published
	DocumentApproved DocumentStatus = "APPROVED"
	// DocumentArchived indicates that the document has been archived and is no longer active
	DocumentArchived DocumentStatus = "ARCHIVED"
	// DocumentStatusInvalid indicates that the document status is invalid
	DocumentStatusInvalid DocumentStatus = "DOCUMENT_STATUS_INVALID"
)

var documentStatusValues = []DocumentStatus{
	DocumentPublished,
	DocumentDraft,
	DocumentNeedsApproval,
	DocumentApproved,
	DocumentArchived,
}

// Values returns a slice of strings that represents all the possible values of the DocumentStatus enum.
// Possible default values are "PUBLISHED", "DRAFT", "NEEDS_APPROVAL", and "APPROVED"
func (DocumentStatus) Values() []string { return stringValues(documentStatusValues) }

// String returns the document status as a string
func (r DocumentStatus) String() string { return string(r) }

// ToDocumentStatus returns the document status enum based on string input
func ToDocumentStatus(r string) *DocumentStatus {
	return parse(r, documentStatusValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r DocumentStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *DocumentStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
