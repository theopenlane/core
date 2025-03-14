package enums

import (
	"fmt"
	"io"
	"strings"
)

type RiskImpact string

var (
	RiskImpactLow      RiskImpact = "LOW"
	RiskImpactModerate RiskImpact = "MODERATE"
	RiskImpactHigh     RiskImpact = "HIGH"
	RiskImpactCritical RiskImpact = "CRITICAL"
	RiskImpactInvalid  RiskImpact = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the RiskImpact enum.
// Possible default values are "LOW", "MODERATE", "HIGH", and "CRITICAL"
func (RiskImpact) Values() (kinds []string) {
	for _, s := range []RiskImpact{RiskImpactLow, RiskImpactModerate, RiskImpactHigh, RiskImpactCritical} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the RiskImpact as a string
func (r RiskImpact) String() string {
	return string(r)
}

// ToRiskImpact returns the user status enum based on string input
func ToRiskImpact(r string) *RiskImpact {
	switch r := strings.ToUpper(r); r {
	case RiskImpactLow.String():
		return &RiskImpactLow
	case RiskImpactModerate.String():
		return &RiskImpactModerate
	case RiskImpactHigh.String():
		return &RiskImpactHigh
	default:
		return &RiskImpactInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r RiskImpact) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *RiskImpact) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Risk Impact, got: %T", v) //nolint:err113
	}

	*r = RiskImpact(str)

	return nil
}
