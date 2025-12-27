package enums

import (
	"fmt"
	"io"
	"strings"
)

// ObjectiveStatus is a custom type for control objective
type ObjectiveStatus string

var (
	// ObjectiveActiveStatus indicates it is currently in draft mode
	ObjectiveDraftStatus ObjectiveStatus = "DRAFT"
	// ObjectiveArchivedStatus indicates it is currently archived
	ObjectiveArchivedStatus ObjectiveStatus = "ARCHIVED"
	// ObjectiveActiveStatus indicates it is currently active
	ObjectiveActiveStatus ObjectiveStatus = "ACTIVE"
)

// Values returns a slice of strings that represents all the possible values of the ObjectiveStatus enum.
// Possible default values are "DRAFT", "ARCHIVED", and "ACTIVE"
func (ObjectiveStatus) Values() (kinds []string) {
	for _, s := range []ObjectiveStatus{
		ObjectiveActiveStatus,
		ObjectiveArchivedStatus,
		ObjectiveDraftStatus,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the objective status as a string
func (r ObjectiveStatus) String() string {
	return string(r)
}

// ToObjectiveStatus returns the objective status enum based on string input
func ToObjectiveStatus(r string) *ObjectiveStatus {
	switch r := strings.ToUpper(r); r {
	case ObjectiveActiveStatus.String():
		return &ObjectiveActiveStatus
	case ObjectiveArchivedStatus.String():
		return &ObjectiveArchivedStatus
	case ObjectiveDraftStatus.String():
		return &ObjectiveDraftStatus
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ObjectiveStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ObjectiveStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ObjectiveStatus, got: %T", v) //nolint:err113
	}

	*r = ObjectiveStatus(str)

	return nil
}
