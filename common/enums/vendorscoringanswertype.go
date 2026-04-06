package enums

import "io"

// VendorScoringAnswerType represents the expected input type for a vendor scoring question
type VendorScoringAnswerType string

var (
	// VendorScoringAnswerTypeBoolean expects a "true" or "false" answer
	VendorScoringAnswerTypeBoolean VendorScoringAnswerType = "BOOLEAN"
	// VendorScoringAnswerTypeText expects a free-form string answer
	VendorScoringAnswerTypeText VendorScoringAnswerType = "TEXT"
	// VendorScoringAnswerTypeSingleSelect expects one value from a predefined list in answer_options
	VendorScoringAnswerTypeSingleSelect VendorScoringAnswerType = "SINGLE_SELECT"
	// VendorScoringAnswerTypeNumeric expects a numeric value parseable as a float
	VendorScoringAnswerTypeNumeric VendorScoringAnswerType = "NUMERIC"
	// VendorScoringAnswerTypeInvalid is returned when parsing an unrecognized value
	VendorScoringAnswerTypeInvalid VendorScoringAnswerType = "INVALID"
)

var vendorScoringAnswerTypeValues = []VendorScoringAnswerType{
	VendorScoringAnswerTypeBoolean,
	VendorScoringAnswerTypeText,
	VendorScoringAnswerTypeSingleSelect,
	VendorScoringAnswerTypeNumeric,
}

// Values returns a slice of strings that represents all the possible values of the VendorScoringAnswerType enum
func (VendorScoringAnswerType) Values() []string { return stringValues(vendorScoringAnswerTypeValues) }

// String returns the VendorScoringAnswerType as a string
func (v VendorScoringAnswerType) String() string { return string(v) }

// ToVendorScoringAnswerType returns the VendorScoringAnswerType based on string input
func ToVendorScoringAnswerType(v string) *VendorScoringAnswerType {
	return parse(v, vendorScoringAnswerTypeValues, &VendorScoringAnswerTypeInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorScoringAnswerType) MarshalGQL(w io.Writer) { marshalGQL(v, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorScoringAnswerType) UnmarshalGQL(val any) error { return unmarshalGQL(v, val) }
