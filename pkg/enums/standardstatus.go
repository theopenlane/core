package enums

import (
	"fmt"
	"io"
	"strings"
)

type StandardStatus string

var (
	StandardActive   StandardStatus = "ACTIVE"
	StandardDraft    StandardStatus = "DRAFT"
	StandardArchived StandardStatus = "ARCHIVED"
	StandardInvalid  StandardStatus = "STANDARD_STATUS_INVALID"
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
