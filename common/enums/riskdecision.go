package enums

import "io"

// RiskDecision is a custom type representing the various states of RiskDecision.
type RiskDecision string

var (
	// RiskDecisionAvoid indicates the avoid.
	RiskDecisionAvoid RiskDecision = "AVOID"
	// RiskDecisionMitigate indicates the  mitigate.
	RiskDecisionMitigate RiskDecision = " MITIGATE"
	// RiskDecisionAccept indicates the  accept.
	RiskDecisionAccept RiskDecision = " ACCEPT"
	// RiskDecisionTransfer indicates the  transfer.
	RiskDecisionTransfer RiskDecision = " TRANSFER"
	// RiskDecisionNone indicates the no decision. has been made
	RiskDecisionNone RiskDecision = " NONE"
	// RiskDecisionInvalid is used when an unknown or unsupported value is provided.
	RiskDecisionInvalid RiskDecision = "RISKDECISION_INVALID"
)

var riskDecisionValues = []RiskDecision{
	RiskDecisionAvoid,
	RiskDecisionMitigate,
	RiskDecisionAccept,
	RiskDecisionTransfer,
	RiskDecisionNone,
}

// Values returns a slice of strings representing all valid RiskDecision values.
func (RiskDecision) Values() []string { return stringValues(riskDecisionValues) }

// String returns the string representation of the RiskDecision value.
func (r RiskDecision) String() string { return string(r) }

// ToRiskDecision converts a string to its corresponding RiskDecision enum value.
func ToRiskDecision(r string) *RiskDecision {
	return parse(r, riskDecisionValues, &RiskDecisionInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r RiskDecision) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *RiskDecision) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
