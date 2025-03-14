package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the StandardStatus enum.
// Possible default values are "ACTIVE", "DRAFT", and "ARCHIVED"
func (StandardStatus) Values() (kinds []string) {
	for _, s := range []StandardStatus{StandardActive, StandardDraft, StandardArchived} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the standard status as a string
func (r StandardStatus) String() string {
	return string(r)
}

// ToStandardStatus returns the standard status enum based on string input
func ToStandardStatus(r string) *StandardStatus {
	switch r := strings.ToUpper(r); r {
	case StandardActive.String():
		return &StandardActive
	case StandardDraft.String():
		return &StandardDraft
	case StandardArchived.String():
		return &StandardArchived
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r StandardStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *StandardStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for StandardStatus, got: %T", v) //nolint:err113
	}

	*r = StandardStatus(str)

	return nil
}
