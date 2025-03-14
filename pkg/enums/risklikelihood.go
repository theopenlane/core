package enums

import (
	"fmt"
	"io"
	"strings"
)

type RiskLikelihood string

var (
	RiskLikelihoodLow     RiskLikelihood = "UNLIKELY"
	RiskLikelihoodMid     RiskLikelihood = "LIKELY"
	RiskLikelihoodHigh    RiskLikelihood = "HIGHLY_LIKELY"
	RiskLikelihoodInvalid RiskLikelihood = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the RiskLikelihood enum.
// Possible default values are "UNLIKELY", "LIKELY", and "HIGHLY_LIKELY"
func (RiskLikelihood) Values() (kinds []string) {
	for _, s := range []RiskLikelihood{RiskLikelihoodLow, RiskLikelihoodMid, RiskLikelihoodHigh} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the RiskLikelihood as a string
func (r RiskLikelihood) String() string {
	return string(r)
}

// ToRiskLikelihood returns the user status enum based on string input
func ToRiskLikelihood(r string) *RiskLikelihood {
	switch r := strings.ToUpper(r); r {
	case RiskLikelihoodLow.String():
		return &RiskLikelihoodLow
	case RiskLikelihoodMid.String():
		return &RiskLikelihoodMid
	case RiskLikelihoodHigh.String():
		return &RiskLikelihoodHigh
	default:
		return &RiskLikelihoodInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r RiskLikelihood) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *RiskLikelihood) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Risk Likelihood, got: %T", v) //nolint:err113
	}

	*r = RiskLikelihood(str)

	return nil
}
