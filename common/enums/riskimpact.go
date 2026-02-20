package enums

import "io"

type RiskImpact string

var (
	RiskImpactLow      RiskImpact = "LOW"
	RiskImpactModerate RiskImpact = "MODERATE"
	RiskImpactHigh     RiskImpact = "HIGH"
	RiskImpactCritical RiskImpact = "CRITICAL"
	RiskImpactInvalid  RiskImpact = "INVALID"
)

var riskImpactValues = []RiskImpact{RiskImpactLow, RiskImpactModerate, RiskImpactHigh, RiskImpactCritical}

var riskImpactParseValues = []RiskImpact{RiskImpactLow, RiskImpactModerate, RiskImpactHigh}

// Values returns a slice of strings that represents all the possible values of the RiskImpact enum.
// Possible default values are "LOW", "MODERATE", "HIGH", and "CRITICAL"
func (RiskImpact) Values() []string { return stringValues(riskImpactValues) }

// String returns the RiskImpact as a string
func (r RiskImpact) String() string { return string(r) }

// ToRiskImpact returns the user status enum based on string input
func ToRiskImpact(r string) *RiskImpact { return parse(r, riskImpactParseValues, &RiskImpactInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r RiskImpact) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *RiskImpact) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
