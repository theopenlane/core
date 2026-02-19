package enums

import "io"

// CampaignStatus is a custom type representing the lifecycle state of a campaign
type CampaignStatus string

var (
	// CampaignStatusDraft indicates a campaign is in draft
	CampaignStatusDraft CampaignStatus = "DRAFT"
	// CampaignStatusScheduled indicates a campaign is scheduled to launch
	CampaignStatusScheduled CampaignStatus = "SCHEDULED"
	// CampaignStatusActive indicates a campaign is active
	CampaignStatusActive CampaignStatus = "ACTIVE"
	// CampaignStatusCompleted indicates a campaign is completed
	CampaignStatusCompleted CampaignStatus = "COMPLETED"
	// CampaignStatusCanceled indicates a campaign was canceled
	CampaignStatusCanceled CampaignStatus = "CANCELED"
	// CampaignStatusInvalid is used when an unknown or unsupported value is provided
	CampaignStatusInvalid CampaignStatus = "INVALID"
)

var campaignStatusValues = []CampaignStatus{
	CampaignStatusDraft,
	CampaignStatusScheduled,
	CampaignStatusActive,
	CampaignStatusCompleted,
	CampaignStatusCanceled,
}

// Values returns a slice of strings that represents all the possible values of the CampaignStatus enum
// Possible default values are "DRAFT", "SCHEDULED", "ACTIVE", "COMPLETED", and "CANCELED"
func (CampaignStatus) Values() []string { return stringValues(campaignStatusValues) }

// String returns the CampaignStatus as a string
func (r CampaignStatus) String() string { return string(r) }

// ToCampaignStatus returns the campaign status enum based on string input
func ToCampaignStatus(r string) *CampaignStatus {
	return parse(r, campaignStatusValues, &CampaignStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r CampaignStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *CampaignStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
