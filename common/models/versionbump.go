package models

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/utils/contextx"
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

// WithVersionBumpContext adds the VersionBump to the context
func WithVersionBumpContext(ctx context.Context, v *VersionBump) context.Context {
	return contextx.With(ctx, v)
}

// VersionBumpFromContext returns the VersionBump from the context
func VersionBumpFromContext(ctx context.Context) (*VersionBump, bool) {
	return contextx.From[*VersionBump](ctx)
}

// WithVersionBumpContext adds the VersionBump to the context
func WithVersionBumpRequestContext(ctx context.Context, v *VersionBump) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err == nil {
		ctx = WithVersionBumpContext(ctx, v)

		ec.SetRequest(ec.Request().WithContext(ctx))
	}
}

// VersionBumpFromContext returns the VersionBump from the context
func VersionBumpFromRequestContext(ctx context.Context) (*VersionBump, bool) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		// try getting it from the context directly
		if v, ok := VersionBumpFromContext(ctx); ok {
			return v, true
		}

		return nil, false
	}

	return VersionBumpFromContext(ec.Request().Context())
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

// ToVersionBump returns the version bump enum based on string input
func ToVersionBump(r string) *VersionBump {
	switch r := strings.ToUpper(r); r {
	case Major.String():
		return &Major
	case Minor.String():
		return &Minor
	case Patch.String():
		return &Patch
	case PreRelease.String():
		return &PreRelease
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (v VersionBump) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + v.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (v *VersionBump) UnmarshalGQL(a any) error {
	str, ok := a.(string)
	if !ok {
		return fmt.Errorf("wrong type for Visibility, got: %T", a) //nolint:err113
	}

	*v = VersionBump(str)

	return nil
}
