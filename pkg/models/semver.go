package models

import (
	"fmt"
	"strings"
)

const (
	// DefaultRevision is the default revision to be used for new records
	DefaultRevision = "v0.0.1"
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
		ver := fmt.Sprintf("%s-%s", base, s.PreRelease)

		return ver
	}

	return base
}

// ToSemverVersion converts a string to a SemverVersion
// It parses the string and returns a SemverVersion object
// It supports the following formats:
// - v1.0.0
// - 1.0.0
// - v1.0.0-alpha
// - 1.0.0-alpha
// anything after the first "-" is considered a pre-release version
func ToSemverVersion(version *string) (*SemverVersion, error) {
	var semver SemverVersion

	if version == nil {
		semver.Patch = 1

		return &semver, nil
	}

	ver := *version

	// Split the version into parts
	// strip the "v" prefix if it exists
	ver = strings.TrimPrefix(ver, "v")

	if strings.Contains(ver, "-") {
		versionParts := strings.SplitN(ver, "-", 2) //nolint:mnd

		// set pre-release version if it exists
		semver.PreRelease = versionParts[1]
	}

	parts := strings.Split(ver, ".")

	// formatted as 1.0.0 after the `v` prefix is removed
	semverParts := 3
	if len(parts) < semverParts {
		return &semver, nil
	}

	// Parse the major, minor, and patch versions
	if _, err := fmt.Sscanf(parts[0], "%d", &semver.Major); err != nil {
		return nil, err
	}

	if _, err := fmt.Sscanf(parts[1], "%d", &semver.Minor); err != nil {
		return nil, err
	}

	if _, err := fmt.Sscanf(parts[2], "%d", &semver.Patch); err != nil {
		return nil, err
	}

	return &semver, nil
}

// BumpMajor increments the major version by 1
// It resets the minor and patch versions to 0
// For example if the version is v1.7.1 the new version will be v2.0.0
// It resets the pre-release version to empty
func BumpMajor(v string) (string, error) {
	semver, err := ToSemverVersion(&v)
	if err != nil {
		return "", err
	}

	semver.Major++
	semver.Patch = 0
	semver.Minor = 0

	if semver.PreRelease != "" {
		semver.PreRelease = ""
	}

	return semver.String(), nil
}

// BumpMinor increments the minor version by 1
// It resets the patch version to 0
// For example if the version is v1.7.1 the new version will be v1.8.0
// It resets the pre-release version to empty
func BumpMinor(v string) (string, error) {
	semver, err := ToSemverVersion(&v)
	if err != nil {
		return "", err
	}

	semver.Minor++

	semver.Patch = 0
	if semver.PreRelease != "" {
		semver.PreRelease = ""
	}

	return semver.String(), nil
}

// BumpPatch increments the patch version by 1
// For example if the version is v1.7.1 the new version will be v1.7.2
// If the version has a pre-release version, it clears the pre-release version
func BumpPatch(v string) (string, error) {
	semver, err := ToSemverVersion(&v)
	if err != nil {
		return "", err
	}

	semver.BumpPatchSemver()

	return semver.String(), nil
}

// BumpPatch increments the patch version by 1
// For example if the version is v1.7.1 the new version will be v1.7.2
// It resets the pre-release version to empty
func (s *SemverVersion) BumpPatchSemver() {
	//  if its prerelease just clear the prerelease
	if s.PreRelease != "" {
		s.PreRelease = ""
	} else {
		// otherwise increment the patch
		s.Patch++
	}
}

// SetPreRelease sets the pre-release version to "draft"
// For example if the version is v1.7.1 the new version will be v1.7.2-draft
func SetPreRelease(v string) (string, error) {
	semver, err := ToSemverVersion(&v)
	if err != nil {
		return "", err
	}

	if semver.PreRelease != "" {
		preRelease := strings.Split(semver.PreRelease, "-")
		if len(preRelease) > 1 {
			// increment the pre-release version
			var preReleaseVersion int
			if _, err := fmt.Sscanf(preRelease[1], "%d", &preReleaseVersion); err != nil {
				return "", err
			}

			semver.PreRelease = fmt.Sprintf("%s-%d", preRelease[0], preReleaseVersion+1)
		} else {
			semver.PreRelease = fmt.Sprintf("%s-%d", semver.PreRelease, 1)
		}
	} else {
		semver.BumpPatchSemver() // update patch version
		semver.PreRelease = "draft"
	}

	return semver.String(), nil
}
