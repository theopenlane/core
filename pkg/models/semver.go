package models

import (
	"fmt"
	"io"
)

// SemverVersion is a custom type for semantic versioning
// It is used to represent the version of objects stored in the database
type SemverVersion struct {
	// Major is the major version
	Major int `json:"major,omitempty"`
	// Minor is the minor version
	Minor int `json:"minor,omitempty"`
	// Patch is the patch version
	Patch int `json:"patch,omitempty"`
	// PreRelease is the pre-release version (used for draft versions)
	PreRelease string `json:"preRelease,omitempty"`
}

// String returns a string representation of the version
func (s SemverVersion) String() string {
	base := fmt.Sprintf("v%d.%d.%d", s.Major, s.Minor, s.Patch)

	if s.PreRelease != "" {
		return fmt.Sprintf("%s-%s", base, s.PreRelease)
	}

	return base
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (s SemverVersion) MarshalGQL(w io.Writer) {
	marshalGQL(w, s)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (s *SemverVersion) UnmarshalGQL(v any) error {
	return unmarshalGQL(v, s)
}
