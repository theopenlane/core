package enums

import "io"

type Tier string

var (
	TierFree       Tier = "FREE"
	TierPro        Tier = "PRO"
	TierEnterprise Tier = "ENTERPRISE"
	TierInvalid    Tier = "INVALID"
)

var tierValues = []Tier{TierFree, TierPro, TierEnterprise}

// Values returns a slice of strings that represents all the possible values of the Tier enum.
// Possible default values are "FREE", "PRO" and "ENTERPRISE".
func (Tier) Values() []string { return stringValues(tierValues) }

// String returns the Tier as a string
func (r Tier) String() string { return string(r) }

// ToTier returns the Tier based on string input
func ToTier(r string) *Tier { return parse(r, tierValues, &TierInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Tier) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Tier) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
