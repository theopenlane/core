package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the JoinPolicy enum.
// Possible default values are "OPEN", "INVITE_ONLY", "APPLICATION_ONLY", and "INVITE_OR_APPLICATION".
func (JoinPolicy) Values() (kinds []string) {
	for _, s := range []JoinPolicy{JoinPolicyOpen, JoinPolicyInviteOnly, JoinPolicyApplicationOnly, JoinPolicyInviteOrApplication} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the JoinPolicy as a string
func (r JoinPolicy) String() string {
	return string(r)
}

// ToGroupJoinPolicy returns the user status enum based on string input
func ToGroupJoinPolicy(r string) *JoinPolicy {
	switch r := strings.ToUpper(r); r {
	case JoinPolicyOpen.String():
		return &JoinPolicyOpen
	case JoinPolicyInviteOnly.String():
		return &JoinPolicyInviteOnly
	case JoinPolicyApplicationOnly.String():
		return &JoinPolicyApplicationOnly
	case JoinPolicyInviteOrApplication.String():
		return &JoinPolicyInviteOrApplication
	default:
		return &JoinPolicyInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r JoinPolicy) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *JoinPolicy) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for JoinPolicy, got: %T", v) //nolint:err113
	}

	*r = JoinPolicy(str)

	return nil
}
