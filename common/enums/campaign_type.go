package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the CampaignType enum
// Possible default values are "QUESTIONNAIRE", "TRAINING", "POLICY_ATTESTATION", "VENDOR_ASSESSMENT", and "CUSTOM"
func (CampaignType) Values() (kinds []string) {
	for _, s := range []CampaignType{
		CampaignTypeQuestionnaire,
		CampaignTypeTraining,
		CampaignTypePolicyAttestation,
		CampaignTypeVendorAssessment,
		CampaignTypeCustom,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the CampaignType as a string
func (r CampaignType) String() string {
	return string(r)
}

// ToCampaignType returns the campaign type enum based on string input
func ToCampaignType(r string) *CampaignType {
	switch r := strings.ToUpper(r); r {
	case CampaignTypeQuestionnaire.String():
		return &CampaignTypeQuestionnaire
	case CampaignTypeTraining.String():
		return &CampaignTypeTraining
	case CampaignTypePolicyAttestation.String():
		return &CampaignTypePolicyAttestation
	case CampaignTypeVendorAssessment.String():
		return &CampaignTypeVendorAssessment
	case CampaignTypeCustom.String():
		return &CampaignTypeCustom
	default:
		return &CampaignTypeInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r CampaignType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *CampaignType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for CampaignType, got: %T", v) //nolint:err113
	}

	*r = CampaignType(str)

	return nil
}
