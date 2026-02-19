package enums

import "io"

// StandardStatus is a custom type for standard status
type StandardStatus string

var (
	// StandardActive indicates that the standard is active and in use
	StandardActive StandardStatus = "ACTIVE"
	// StandardDraft indicates that the standard is in draft status and not yet finalized
	StandardDraft StandardStatus = "DRAFT"
	// StandardArchived indicates that the standard has been archived and is no longer active
	StandardArchived StandardStatus = "ARCHIVED"
	// StandardInvalid indicates that the standard status is invalid
	StandardInvalid StandardStatus = "STANDARD_STATUS_INVALID"
)

var standardStatusValues = []StandardStatus{StandardActive, StandardDraft, StandardArchived}

// Values returns a slice of strings that represents all the possible values of the StandardStatus enum.
// Possible default values are "ACTIVE", "DRAFT", and "ARCHIVED"
func (StandardStatus) Values() []string { return stringValues(standardStatusValues) }

// String returns the standard status as a string
func (r StandardStatus) String() string { return string(r) }

// ToStandardStatus returns the standard status enum based on string input
func ToStandardStatus(r string) *StandardStatus { return parse(r, standardStatusValues, nil) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r StandardStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *StandardStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
