package enums

import "io"

type JoinPolicy string

var (
	// JoinPolicyOpen is when the group is open for anyone to join
	JoinPolicyOpen JoinPolicy = "OPEN"
	// JoinPolicyInviteOnly is when the group is only joinable by invite
	JoinPolicyInviteOnly JoinPolicy = "INVITE_ONLY"
	// JoinPolicyApplicationOnly is when the group is only joinable by application
	JoinPolicyApplicationOnly JoinPolicy = "APPLICATION_ONLY"
	// JoinPolicyInviteOrApplication is when the group is joinable by invite or application
	JoinPolicyInviteOrApplication JoinPolicy = "INVITE_OR_APPLICATION"
	// JoinPolicyInvalid is the default value for the JoinPolicy enum
	JoinPolicyInvalid JoinPolicy = "INVALID"
)

var joinPolicyValues = []JoinPolicy{JoinPolicyOpen, JoinPolicyInviteOnly, JoinPolicyApplicationOnly, JoinPolicyInviteOrApplication}

// Values returns a slice of strings that represents all the possible values of the JoinPolicy enum.
// Possible default values are "OPEN", "INVITE_ONLY", "APPLICATION_ONLY", and "INVITE_OR_APPLICATION".
func (JoinPolicy) Values() []string { return stringValues(joinPolicyValues) }

// String returns the JoinPolicy as a string
func (r JoinPolicy) String() string { return string(r) }

// ToGroupJoinPolicy returns the user status enum based on string input
func ToGroupJoinPolicy(r string) *JoinPolicy {
	return parse(r, joinPolicyValues, &JoinPolicyInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r JoinPolicy) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *JoinPolicy) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
