package enums

import "io"

// RiskStatus is a custom type for risk status
type RiskStatus string

var (
	// RiskOpen indicates that the risk is open and has not been mitigated
	RiskOpen RiskStatus = "OPEN"
	// RiskInProgress indicates that the risk is being actively worked on
	RiskInProgress RiskStatus = "IN_PROGRESS"
	// RiskOngoing indicates that the risk is ongoing and has not been mitigated
	RiskOngoing RiskStatus = "ONGOING"
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

var riskStatusValues = []RiskStatus{
	RiskOpen, RiskInProgress, RiskOngoing, RiskIdentified, RiskMitigated,
	RiskAccepted, RiskClosed, RiskTransferred, RiskArchived,
}

// Values returns a slice of strings that represents all the possible values of the RiskStatus enum.
// Possible default values are "OPEN", "IN_PROGRESS", "ONGOING", "IDENTIFIED", "MITIGATED", "ACCEPTED", "CLOSED", "TRANSFERRED", and "ARCHIVED"
func (RiskStatus) Values() []string { return stringValues(riskStatusValues) }

// String returns the risk status as a string
func (r RiskStatus) String() string { return string(r) }

// ToRiskStatus returns the risk status enum based on string input
func ToRiskStatus(r string) *RiskStatus { return parse(r, riskStatusValues, nil) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r RiskStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *RiskStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
