package enums

import "io"

// VendorTier represents the risk tier classification for a vendor entity,
// used to determine the depth of TPRM assessment required
type VendorTier string

var (
	// VendorTierCritical indicates a strategic vendor with high data access, PII processing,
	// or operational essentiality — requires comprehensive assessment across all domains
	VendorTierCritical VendorTier = "CRITICAL"
	// VendorTierHigh indicates a vendor with significant data access or spend
	// but not operationally critical — requires assessment across core domains
	VendorTierHigh VendorTier = "HIGH"
	// VendorTierStandard indicates a vendor with limited data access and moderate spend
	// — requires assessment of foundational controls only
	VendorTierStandard VendorTier = "STANDARD"
	// VendorTierLow indicates a commodity vendor with minimal data access and easily replaceable
	// — requires minimal assessment
	VendorTierLow VendorTier = "LOW"
	// VendorTierInvalid is returned when parsing an unrecognized value
	VendorTierInvalid VendorTier = "INVALID"
)

var vendorTierValues = []VendorTier{
	VendorTierCritical,
	VendorTierHigh,
	VendorTierStandard,
	VendorTierLow,
}

// Values returns a slice of strings that represents all the possible values of the VendorTier enum
func (VendorTier) Values() []string { return stringValues(vendorTierValues) }

// String returns the VendorTier as a string
func (v VendorTier) String() string { return string(v) }

// ToVendorTier returns the VendorTier based on string input
func ToVendorTier(v string) *VendorTier {
	return parse(v, vendorTierValues, &VendorTierInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorTier) MarshalGQL(w io.Writer) { marshalGQL(v, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorTier) UnmarshalGQL(val any) error { return unmarshalGQL(v, val) }
