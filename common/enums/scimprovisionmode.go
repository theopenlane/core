package enums

import "io"

// SCIMProvisionMode controls how SCIM push events are persisted.
type SCIMProvisionMode string

var (
	// SCIMProvisionModeUsers creates User and Group entities from SCIM push events.
	SCIMProvisionModeUsers SCIMProvisionMode = "USERS"
	// SCIMProvisionModeDirectory creates DirectoryAccount and DirectoryGroup snapshot records from SCIM push events.
	SCIMProvisionModeDirectory SCIMProvisionMode = "DIRECTORY"
	// SCIMProvisionModeBoth creates both User/Group entities and DirectoryAccount/DirectoryGroup records from SCIM push events.
	SCIMProvisionModeBoth SCIMProvisionMode = "BOTH"
)

var scimProvisionModeValues = []SCIMProvisionMode{
	SCIMProvisionModeUsers,
	SCIMProvisionModeDirectory,
	SCIMProvisionModeBoth,
}

// SCIMProvisionModes is a list of all valid SCIMProvisionMode values.
var SCIMProvisionModes = stringValues(scimProvisionModeValues)

// Values returns a slice of strings that represents all the possible values of the SCIMProvisionMode enum.
func (SCIMProvisionMode) Values() []string { return SCIMProvisionModes }

// String returns the SCIMProvisionMode as a string.
func (r SCIMProvisionMode) String() string { return string(r) }

// ToSCIMProvisionMode returns the SCIMProvisionMode based on string input.
func ToSCIMProvisionMode(r string) *SCIMProvisionMode {
	return parse(r, scimProvisionModeValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r SCIMProvisionMode) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *SCIMProvisionMode) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
