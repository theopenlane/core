package models

import (
	"fmt"
	"io"
	"slices"
)

// OrgModule identifies a purchasable module
type OrgModule string

// IsValid reports whether m is a known module constant
func (m OrgModule) IsValid() bool {
	return slices.Contains(AllOrgModules, m)
}

// String returns the string representation of the OrgModule
func (m OrgModule) String() string {
	return string(m)
}

// UnmarshalText implements encoding.TextUnmarshaler
func (m *OrgModule) UnmarshalText(text []byte) error {
	*m = OrgModule(text)

	if !m.IsValid() {
		return fmt.Errorf("invalid OrgModule: %q", text) //nolint:err113
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler
func (m OrgModule) MarshalText() ([]byte, error) {
	return []byte(m), nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (m OrgModule) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + m.String() + `"`))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (m *OrgModule) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for OrgModule, got: %T", v) //nolint:err113
	}

	*m = OrgModule(str)

	return nil
}
