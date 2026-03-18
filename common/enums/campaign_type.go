package enums

import "io"

// CampaignType is a custom type representing the type of a campaign
type CampaignType string

var (
	// CampaignTypeQuestionnaire indicates a questionnaire campaign
	CampaignTypeQuestionnaire CampaignType = "QUESTIONNAIRE"
	// CampaignTypeTraining indicates a training campaign
	CampaignTypeTraining CampaignType = "TRAINING"
	// CampaignTypePolicyAttestation indicates a policy attestation campaign
	CampaignTypePolicyAttestation CampaignType = "POLICY_ATTESTATION"
	// CampaignTypeVendorAssessment indicates a vendor assessment campaign
	CampaignTypeVendorAssessment CampaignType = "VENDOR_ASSESSMENT"
	// CampaignTypeCustom indicates a custom campaign
	CampaignTypeCustom CampaignType = "CUSTOM"
	// CampaignTypeInvalid is used when an unknown or unsupported value is provided
	CampaignTypeInvalid CampaignType = "INVALID"
)

var campaignTypeValues = []CampaignType{
	CampaignTypeQuestionnaire,
	CampaignTypeTraining,
	CampaignTypePolicyAttestation,
	CampaignTypeVendorAssessment,
	CampaignTypeCustom,
}

// Values returns a slice of strings that represents all the possible values of the CampaignType enum
// Possible default values are "QUESTIONNAIRE", "TRAINING", "POLICY_ATTESTATION", "VENDOR_ASSESSMENT", and "CUSTOM"
func (CampaignType) Values() []string { return stringValues(campaignTypeValues) }

// String returns the CampaignType as a string
func (r CampaignType) String() string { return string(r) }

// ToCampaignType returns the campaign type enum based on string input
func ToCampaignType(r string) *CampaignType {
	return parse(r, campaignTypeValues, &CampaignTypeInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r CampaignType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *CampaignType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
