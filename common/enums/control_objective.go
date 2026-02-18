package enums

import "io"

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

var objectiveStatusValues = []ObjectiveStatus{
	ObjectiveActiveStatus,
	ObjectiveArchivedStatus,
	ObjectiveDraftStatus,
}

// Values returns a slice of strings that represents all the possible values of the ObjectiveStatus enum.
// Possible default values are "DRAFT", "ARCHIVED", and "ACTIVE"
func (ObjectiveStatus) Values() []string { return stringValues(objectiveStatusValues) }

// String returns the objective status as a string
func (r ObjectiveStatus) String() string { return string(r) }

// ToObjectiveStatus returns the objective status enum based on string input
func ToObjectiveStatus(r string) *ObjectiveStatus {
	return parse(r, objectiveStatusValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ObjectiveStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ObjectiveStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
