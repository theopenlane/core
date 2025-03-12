package models

import (
	"fmt"
	"io"
	"strings"
)

// VersionBump is a custom type for version bumping
// It is used to represent the type of version bumping
type VersionBump string

var (
	// Major is the major version
	Major VersionBump = "MAJOR"
	// Minor is the minor version
	Minor VersionBump = "MINOR"
	// Patch is the patch version
	Patch VersionBump = "PATCH"
	// PreRelease is the pre-release version
	PreRelease VersionBump = "DRAFT"
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
func (s SemverVersion) String() *string {
	base := fmt.Sprintf("v%d.%d.%d", s.Major, s.Minor, s.Patch)

	if s.PreRelease != "" {
		ver := fmt.Sprintf("%s-%s", base, s.PreRelease)

		return &ver
	}

	return &base
}

// ToSemverVersion converts a string to a SemverVersion
// It parses the string and returns a SemverVersion object
// It supports the following formats:
// - v1.0.0
// - 1.0.0
// - v1.0.0-alpha
// - 1.0.0-alpha
// anything after the first "-" is considered a pre-release version
func ToSemverVersion(version *string) *SemverVersion {
	var semver SemverVersion

	if version == nil {
		semver.Patch = 1

		return &semver
	}

	ver := *version

	// Split the version into parts
	// strip the "v" prefix if it exists
	ver = strings.TrimPrefix(ver, "v")

	if strings.Contains(ver, "-") {
		versionParts := strings.Split(ver, "-")
		ver = versionParts[0]

		// set pre-release version if it exists
		semver.PreRelease = versionParts[1]
	}

	parts := strings.Split(ver, ".")

	if len(parts) < 3 {
		return &semver
	}

	// Parse the major, minor, and patch versions
	fmt.Sscanf(parts[0], "%d", &semver.Major)
	fmt.Sscanf(parts[1], "%d", &semver.Minor)
	fmt.Sscanf(parts[2], "%d", &semver.Patch)

	return &semver
}

// BumpMajor increments the major version by 1
// It resets the minor and patch versions to 0
// For example if the version is v1.7.1 the new version will be v2.0.0
// It resets the pre-release version to empty
func (s *SemverVersion) BumpMajor() {
	s.Major++
	s.Patch = 0
	s.Minor = 0
	if s.PreRelease != "" {
		s.PreRelease = ""
	}
}

// BumpMinor increments the minor version by 1
// It resets the patch version to 0
// For example if the version is v1.7.1 the new version will be v1.8.0
// It resets the pre-release version to empty
func (s *SemverVersion) BumpMinor() {
	s.Minor++
	s.Patch = 0
	if s.PreRelease != "" {
		s.PreRelease = ""
	}
}

// BumpPatch increments the patch version by 1
// For example if the version is v1.7.1 the new version will be v1.7.2
// It resets the pre-release version to empty
func (s *SemverVersion) BumpPatch() {
	s.Patch++
	if s.PreRelease != "" {
		s.PreRelease = ""
	}
}

// BumpMajorVersionString increments the major version by 1 of a string version
func BumpMajorVersionString(version *string) *string {
	semver := ToSemverVersion(version)
	semver.BumpMajor()
	return semver.String()
}

// BumpMinorVersionString increments the minor version by 1 of a string version
func BumpMinorVersionString(version *string) *string {
	semver := ToSemverVersion(version)
	semver.BumpMinor()
	return semver.String()
}

// BumpPatchVersionString increments the patch version by 1 of a string version
func BumpPatchVersionString(version *string) *string {
	semver := ToSemverVersion(version)
	semver.BumpPatch()
	return semver.String()
}

// String returns the role as a string
func (v VersionBump) String() string {
	return string(v)
}

// Values returns a slice of strings that represents all the possible values of the VersionBump enum.
// Possible default values are "MAJOR", "MINOR", "PATCH", "DRAFT"
func (VersionBump) Values() (kinds []string) {
	for _, s := range []VersionBump{Major, Minor, Patch, PreRelease} {
		kinds = append(kinds, string(s))
	}

	return
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r VersionBump) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *VersionBump) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for VersionBump, got: %T", v) //nolint:err113
	}

	*r = VersionBump(str)

	return nil
}
