package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the DocumentStatus enum.
// Possible default values are "PUBLISHED", "DRAFT", "NEEDS_APPROVAL", and "APPROVED"
func (DocumentStatus) Values() (kinds []string) {
	for _, s := range []DocumentStatus{DocumentPublished, DocumentDraft, DocumentNeedsApproval, DocumentApproved, DocumentArchived} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the document status as a string
func (r DocumentStatus) String() string {
	return string(r)
}

// ToDocumentStatus returns the document status enum based on string input
func ToDocumentStatus(r string) *DocumentStatus {
	switch r := strings.ToUpper(r); r {
	case DocumentPublished.String():
		return &DocumentPublished
	case DocumentDraft.String():
		return &DocumentDraft
	case DocumentNeedsApproval.String():
		return &DocumentNeedsApproval
	case DocumentApproved.String():
		return &DocumentApproved
	case DocumentArchived.String():
		return &DocumentArchived
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r DocumentStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *DocumentStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for DocumentStatus, got: %T", v) //nolint:err113
	}

	*r = DocumentStatus(str)

	return nil
}
