package enums

import "io"

// DemoDataProfile is a custom type naming the seed time rules applied to a demo dataset's records
type DemoDataProfile string

var (
	// DemoDataProfileDemo fills the dataset out with statuses, program links and assignments
	DemoDataProfileDemo DemoDataProfile = "DEMO"
	// DemoDataProfileDraft seeds the dataset's controls, policies and procedures as drafts with nothing linked or assigned
	DemoDataProfileDraft DemoDataProfile = "DRAFT"
	// DemoDataProfileInvalid is used when an unknown or unsupported value is provided
	DemoDataProfileInvalid DemoDataProfile = "DEMO_DATA_PROFILE_INVALID"
)

var demoDataProfileValues = []DemoDataProfile{
	DemoDataProfileDemo,
	DemoDataProfileDraft,
}

// Values returns a slice of strings that represents all the possible values of the DemoDataProfile enum
// Possible default values are "DEMO" and "DRAFT"
func (DemoDataProfile) Values() []string { return stringValues(demoDataProfileValues) }

// String returns the DemoDataProfile as a string
func (r DemoDataProfile) String() string { return string(r) }

// ToDemoDataProfile returns the demo data profile enum based on string input
func ToDemoDataProfile(r string) *DemoDataProfile {
	return parse(r, demoDataProfileValues, &DemoDataProfileInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r DemoDataProfile) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *DemoDataProfile) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
