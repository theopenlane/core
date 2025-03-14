package enums

import (
	"fmt"
	"io"
	"strings"
)

type RiskStatus string

var (
	RiskOpen       RiskStatus = "OPEN"
	RiskInProgress RiskStatus = "IN_PROGRESS"
	RiskOngoing    RiskStatus = "ONGOING"
	RiskMitigated  RiskStatus = "MITIGATED"
	RiskArchived   RiskStatus = "ARCHIVED"
	RiskInvalid    RiskStatus = "RISK_STATUS_INVALID"
)

// Values returns a slice of strings that represents all the possible values of the RiskStatus enum.
// Possible default values are "OPEN", "IN_PROGRESS",  "ONGOING", "MITIGATED", and "ARCHIVED"
func (RiskStatus) Values() (kinds []string) {
	for _, s := range []RiskStatus{RiskOpen, RiskInProgress, RiskOngoing, RiskMitigated, RiskArchived} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the risk status as a string
func (r RiskStatus) String() string {
	return string(r)
}

// ToRiskStatus returns the risk status enum based on string input
func ToRiskStatus(r string) *RiskStatus {
	switch r := strings.ToUpper(r); r {
	case RiskOpen.String():
		return &RiskOpen
	case RiskInProgress.String():
		return &RiskInProgress
	case RiskOngoing.String():
		return &RiskOngoing
	case RiskMitigated.String():
		return &RiskMitigated
	case RiskArchived.String():
		return &RiskArchived
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r RiskStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *RiskStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for RiskStatus, got: %T", v) //nolint:err113
	}

	*r = RiskStatus(str)

	return nil
}
