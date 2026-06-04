package enums

import (
	"fmt"
	"io"
)

// DocumentManagementMode is a custom type representing how a document is managed:
// parsed and edited natively in Openlane, or kept as an external reference file
// (for example, a Word document managed through the user's native editor).
type DocumentManagementMode string

var (
	// DocumentManagementModeOpenlaneManaged indicates the document is parsed and edited in Openlane.
	DocumentManagementModeOpenlaneManaged DocumentManagementMode = "OPENLANE_MANAGED"
	// DocumentManagementModeExternalReference indicates the document is uploaded as a viewable
	// reference while being managed in an external editor (such as Microsoft Word).
	DocumentManagementModeExternalReference DocumentManagementMode = "EXTERNAL_REFERENCE"
	// DocumentManagementModeIntegration is used when the document is managed by an external system integrated with Openlane.
	DocumentManagementModeIntegration DocumentManagementMode = "INTEGRATION"
	// DocumentManagementModeInvalid is used when an unknown or unsupported value is provided.
	DocumentManagementModeInvalid DocumentManagementMode = "DOCUMENT_MANAGEMENT_MODE_INVALID"
)

var documentManagementModeValues = []DocumentManagementMode{
	DocumentManagementModeOpenlaneManaged,
	DocumentManagementModeExternalReference,
	DocumentManagementModeIntegration,
}

// Values returns a slice of strings representing all valid DocumentManagementMode values.
func (DocumentManagementMode) Values() []string { return stringValues(documentManagementModeValues) }

// String returns the string representation of the DocumentManagementMode value.
func (r DocumentManagementMode) String() string { return string(r) }

// IsValid reports whether r is a recognised DocumentManagementMode value.
func (r DocumentManagementMode) IsValid() bool {
	for _, v := range documentManagementModeValues {
		if r == v {
			return true
		}
	}
	return false
}

// ToDocumentManagementMode converts a string to its corresponding DocumentManagementMode enum value.
// Returns nil for unknown values. Prefer ToDocumentManagementModeOrDefault when a fallback is acceptable.
func ToDocumentManagementMode(r string) *DocumentManagementMode {
	return parse(r, documentManagementModeValues, nil)
}

// ToDocumentManagementModeOrDefault parses r and falls back to OPENLANE_MANAGED on unknown input.
// Safe to dereference at the call site.
func ToDocumentManagementModeOrDefault(r string) DocumentManagementMode {
	if v := parse(r, documentManagementModeValues, nil); v != nil {
		return *v
	}
	return DocumentManagementModeOpenlaneManaged
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r DocumentManagementMode) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface; rejects values outside the enum.
func (r *DocumentManagementMode) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return ErrInvalidType
	}
	candidate := DocumentManagementMode(str)
	if !candidate.IsValid() {
		return fmt.Errorf("%w: %q is not a valid DocumentManagementMode", ErrInvalidType, str)
	}
	*r = candidate
	return nil
}
