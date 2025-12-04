package enums

import (
	"fmt"
	"io"
	"strings"
)

// RiskStatus is a custom type for risk status
type RiskStatus string

var (
	// RiskIdentified indicates that the risk has been identified
	RiskIdentified RiskStatus = "IDENTIFIED"
	// RiskMitigated indicates that the risk has been mitigated
	RiskMitigated RiskStatus = "MITIGATED"
	// RiskAccepted indicates that the risk has been accepted
	RiskAccepted RiskStatus = "ACCEPTED"
	// RiskClosed indicates that the risk has been closed
	RiskClosed RiskStatus = "CLOSED"
	// RiskTransferred indicates that the risk has been transferred
	RiskTransferred RiskStatus = "TRANSFERRED"
	// RiskArchived indicates that the risk has been archived and is no longer active
	RiskArchived RiskStatus = "ARCHIVED"
	// RiskInvalid indicates that the risk status is invalid
	RiskInvalid RiskStatus = "RISK_STATUS_INVALID"
)

// Values returns a slice of strings that represents all the possible values of the RiskStatus enum.
// Possible default values are "IDENTIFIED", "MITIGATED", "ACCEPTED", "CLOSED", "TRANSFERRED", and "ARCHIVED"
func (RiskStatus) Values() (kinds []string) {
	for _, s := range []RiskStatus{RiskIdentified, RiskMitigated, RiskAccepted, RiskClosed, RiskTransferred, RiskArchived} {
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
	case RiskIdentified.String():
		return &RiskIdentified
	case RiskMitigated.String():
		return &RiskMitigated
	case RiskAccepted.String():
		return &RiskAccepted
	case RiskClosed.String():
		return &RiskClosed
	case RiskTransferred.String():
		return &RiskTransferred
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
