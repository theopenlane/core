package enums

import (
	"fmt"
	"io"
	"strings"
)

type Tier string

var (
	TierFree       Tier = "FREE"
	TierPro        Tier = "PRO"
	TierEnterprise Tier = "ENTERPRISE"
	TierInvalid    Tier = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the Tier enum.
// Possible default values are "FREE", "PRO" and "ENTERPRISE".
func (Tier) Values() (kinds []string) {
	for _, s := range []Tier{TierFree, TierPro, TierEnterprise} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the Tier as a string
func (r Tier) String() string {
	return string(r)
}

// ToTier returns the Tier based on string input
func ToTier(r string) *Tier {
	switch r := strings.ToUpper(r); r {
	case TierFree.String():
		return &TierFree
	case TierPro.String():
		return &TierPro
	case TierEnterprise.String():
		return &TierEnterprise
	default:
		return &TierInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Tier) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Tier) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Tier, got: %T", v) //nolint:err113
	}

	*r = Tier(str)

	return nil
}
