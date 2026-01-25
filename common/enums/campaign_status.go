package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the CampaignStatus enum
// Possible default values are "DRAFT", "SCHEDULED", "ACTIVE", "COMPLETED", and "CANCELED"
func (CampaignStatus) Values() (kinds []string) {
	for _, s := range []CampaignStatus{
		CampaignStatusDraft,
		CampaignStatusScheduled,
		CampaignStatusActive,
		CampaignStatusCompleted,
		CampaignStatusCanceled,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the CampaignStatus as a string
func (r CampaignStatus) String() string {
	return string(r)
}

// ToCampaignStatus returns the campaign status enum based on string input
func ToCampaignStatus(r string) *CampaignStatus {
	switch r := strings.ToUpper(r); r {
	case CampaignStatusDraft.String():
		return &CampaignStatusDraft
	case CampaignStatusScheduled.String():
		return &CampaignStatusScheduled
	case CampaignStatusActive.String():
		return &CampaignStatusActive
	case CampaignStatusCompleted.String():
		return &CampaignStatusCompleted
	case CampaignStatusCanceled.String():
		return &CampaignStatusCanceled
	default:
		return &CampaignStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r CampaignStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *CampaignStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for CampaignStatus, got: %T", v) //nolint:err113
	}

	*r = CampaignStatus(str)

	return nil
}
