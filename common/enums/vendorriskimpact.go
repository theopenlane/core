package enums

import "io"

// VendorRiskImpact represents the impact level of a vendor risk criterion using the standard
// 5-point TPRM scoring scale. Numeric values: VeryLow=1, Low=2, Medium=3, High=4, Critical=5.
type VendorRiskImpact string

var (
	// VendorRiskImpactVeryLow indicates negligible business consequences (numeric value: 1)
	VendorRiskImpactVeryLow VendorRiskImpact = "VERY_LOW"
	// VendorRiskImpactLow indicates minor consequences with limited disruption (numeric value: 2)
	VendorRiskImpactLow VendorRiskImpact = "LOW"
	// VendorRiskImpactMedium indicates moderate consequences, manageable with effort (numeric value: 3)
	VendorRiskImpactMedium VendorRiskImpact = "MEDIUM"
	// VendorRiskImpactHigh indicates significant consequences requiring major response (numeric value: 4)
	VendorRiskImpactHigh VendorRiskImpact = "HIGH"
	// VendorRiskImpactCritical indicates severe, potentially existential consequences (numeric value: 5)
	VendorRiskImpactCritical VendorRiskImpact = "CRITICAL"
	// VendorRiskImpactInvalid is returned when parsing an unrecognized value
	VendorRiskImpactInvalid VendorRiskImpact = "INVALID"
)

var vendorRiskImpactValues = []VendorRiskImpact{
	VendorRiskImpactVeryLow,
	VendorRiskImpactLow,
	VendorRiskImpactMedium,
	VendorRiskImpactHigh,
	VendorRiskImpactCritical,
}

// Values returns a slice of strings that represents all the possible values of the VendorRiskImpact enum
func (VendorRiskImpact) Values() []string { return stringValues(vendorRiskImpactValues) }

// String returns the VendorRiskImpact as a string
func (v VendorRiskImpact) String() string { return string(v) }

// ToVendorRiskImpact returns the VendorRiskImpact based on string input
func ToVendorRiskImpact(v string) *VendorRiskImpact {
	return parse(v, vendorRiskImpactValues, &VendorRiskImpactInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorRiskImpact) MarshalGQL(w io.Writer) { marshalGQL(v, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorRiskImpact) UnmarshalGQL(val any) error { return unmarshalGQL(v, val) }
