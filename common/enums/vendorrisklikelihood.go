package enums

import "io"

// VendorRiskLikelihood represents the likelihood level of a vendor risk criterion using the standard
// 5-point TPRM scoring scale. Numeric values: VeryLow=0.5, Low=1, Medium=2, High=3, VeryHigh=4.
type VendorRiskLikelihood string

var (
	// VendorRiskLikelihoodVeryLow indicates rare or theoretical occurrence, <10% probability (numeric value: 0.5)
	VendorRiskLikelihoodVeryLow VendorRiskLikelihood = "VERY_LOW"
	// VendorRiskLikelihoodLow indicates unlikely within 5 years, 10-30% probability (numeric value: 1)
	VendorRiskLikelihoodLow VendorRiskLikelihood = "LOW"
	// VendorRiskLikelihoodMedium indicates possible within 2-5 years, 30-70% probability (numeric value: 2)
	VendorRiskLikelihoodMedium VendorRiskLikelihood = "MEDIUM"
	// VendorRiskLikelihoodHigh indicates likely within 1-2 years, 70-90% probability (numeric value: 3)
	VendorRiskLikelihoodHigh VendorRiskLikelihood = "HIGH"
	// VendorRiskLikelihoodVeryHigh indicates almost certain within 1 year, >90% probability (numeric value: 4)
	VendorRiskLikelihoodVeryHigh VendorRiskLikelihood = "VERY_HIGH"
	// VendorRiskLikelihoodInvalid is returned when parsing an unrecognized value
	VendorRiskLikelihoodInvalid VendorRiskLikelihood = "INVALID"
)

var vendorRiskLikelihoodValues = []VendorRiskLikelihood{
	VendorRiskLikelihoodVeryLow,
	VendorRiskLikelihoodLow,
	VendorRiskLikelihoodMedium,
	VendorRiskLikelihoodHigh,
	VendorRiskLikelihoodVeryHigh,
}

// Values returns a slice of strings that represents all the possible values of the VendorRiskLikelihood enum
func (VendorRiskLikelihood) Values() []string { return stringValues(vendorRiskLikelihoodValues) }

// String returns the VendorRiskLikelihood as a string
func (v VendorRiskLikelihood) String() string { return string(v) }

// ToVendorRiskLikelihood returns the VendorRiskLikelihood based on string input
func ToVendorRiskLikelihood(v string) *VendorRiskLikelihood {
	return parse(v, vendorRiskLikelihoodValues, &VendorRiskLikelihoodInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorRiskLikelihood) MarshalGQL(w io.Writer) { marshalGQL(v, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorRiskLikelihood) UnmarshalGQL(val any) error { return unmarshalGQL(v, val) }
