package enums

import "io"

type RiskLikelihood string

var (
	RiskLikelihoodLow     RiskLikelihood = "UNLIKELY"
	RiskLikelihoodMid     RiskLikelihood = "LIKELY"
	RiskLikelihoodHigh    RiskLikelihood = "HIGHLY_LIKELY"
	RiskLikelihoodInvalid RiskLikelihood = "INVALID"
)

var riskLikelihoodValues = []RiskLikelihood{RiskLikelihoodLow, RiskLikelihoodMid, RiskLikelihoodHigh}

// Values returns a slice of strings that represents all the possible values of the RiskLikelihood enum.
// Possible default values are "UNLIKELY", "LIKELY", and "HIGHLY_LIKELY"
func (RiskLikelihood) Values() []string { return stringValues(riskLikelihoodValues) }

// String returns the RiskLikelihood as a string
func (r RiskLikelihood) String() string { return string(r) }

// ToRiskLikelihood returns the user status enum based on string input
func ToRiskLikelihood(r string) *RiskLikelihood {
	return parse(r, riskLikelihoodValues, &RiskLikelihoodInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r RiskLikelihood) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *RiskLikelihood) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
