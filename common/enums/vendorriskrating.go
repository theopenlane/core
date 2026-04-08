package enums

import "io"

// VendorRiskRating represents the risk rating tier for a vendor entity
type VendorRiskRating string

var (
	// VendorRiskRatingNone indicates no assessed risk; the default for a zero score
	VendorRiskRatingNone VendorRiskRating = "NONE"
	// VendorRiskRatingVeryLow indicates negligible overall risk (default max score: 3)
	VendorRiskRatingVeryLow VendorRiskRating = "VERY_LOW"
	// VendorRiskRatingLow indicates minor overall risk (default max score: 5)
	VendorRiskRatingLow VendorRiskRating = "LOW"
	// VendorRiskRatingMedium indicates moderate overall risk (default max score: 11)
	VendorRiskRatingMedium VendorRiskRating = "MEDIUM"
	// VendorRiskRatingHigh indicates significant overall risk (default max score: 15)
	VendorRiskRatingHigh VendorRiskRating = "HIGH"
	// VendorRiskRatingCritical indicates severe overall risk (catch-all for scores above HIGH threshold)
	VendorRiskRatingCritical VendorRiskRating = "CRITICAL"
	// VendorRiskRatingInvalid is returned when parsing an unrecognized value
	VendorRiskRatingInvalid VendorRiskRating = "INVALID"
)

var vendorRiskRatingValues = []VendorRiskRating{
	VendorRiskRatingNone,
	VendorRiskRatingVeryLow,
	VendorRiskRatingLow,
	VendorRiskRatingMedium,
	VendorRiskRatingHigh,
	VendorRiskRatingCritical,
}

// Values returns a slice of strings that represents all the possible values of the VendorRiskRating enum
func (VendorRiskRating) Values() []string { return stringValues(vendorRiskRatingValues) }

// String returns the VendorRiskRating as a string
func (v VendorRiskRating) String() string { return string(v) }

// ToVendorRiskRating returns the VendorRiskRating based on string input
func ToVendorRiskRating(v string) *VendorRiskRating {
	return parse(v, vendorRiskRatingValues, &VendorRiskRatingInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorRiskRating) MarshalGQL(w io.Writer) { marshalGQL(v, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorRiskRating) UnmarshalGQL(val any) error { return unmarshalGQL(v, val) }
