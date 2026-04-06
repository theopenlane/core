package enums

import "io"

// VendorScoringMode controls how unanswered questions affect the aggregate vendor risk score
type VendorScoringMode string

var (
	// VendorScoringModeAnsweredOnly aggregates only questions that have been answered; unanswered questions contribute zero to the total score
	VendorScoringModeAnsweredOnly VendorScoringMode = "ANSWERED_ONLY"
	// VendorScoringModeFullQuestionnaire treats unanswered questions as maximum risk (impact=CRITICAL x likelihood=VERY_HIGH) in the aggregate score
	VendorScoringModeFullQuestionnaire VendorScoringMode = "FULL_QUESTIONNAIRE"
	// VendorScoringModeManual disables automatic entity-level risk aggregation; risk_score, risk_rating, and risk_score_coverage are set directly by the user
	VendorScoringModeManual VendorScoringMode = "MANUAL"
	// VendorScoringModeInvalid is returned when parsing an unrecognized value
	VendorScoringModeInvalid VendorScoringMode = "INVALID"
)

var vendorScoringModeValues = []VendorScoringMode{
	VendorScoringModeAnsweredOnly,
	VendorScoringModeFullQuestionnaire,
	VendorScoringModeManual,
}

// Values returns a slice of strings that represents all the possible values of the VendorScoringMode enum
func (VendorScoringMode) Values() []string { return stringValues(vendorScoringModeValues) }

// String returns the VendorScoringMode as a string
func (v VendorScoringMode) String() string { return string(v) }

// ToVendorScoringMode returns the VendorScoringMode based on string input
func ToVendorScoringMode(v string) *VendorScoringMode {
	return parse(v, vendorScoringModeValues, &VendorScoringModeInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorScoringMode) MarshalGQL(w io.Writer) { marshalGQL(v, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorScoringMode) UnmarshalGQL(val any) error { return unmarshalGQL(v, val) }
