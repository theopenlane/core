package enums

import "io"

type InviteStatus string

var (
	InvitationSent     InviteStatus = "INVITATION_SENT"
	ApprovalRequired   InviteStatus = "APPROVAL_REQUIRED"
	InvitationAccepted InviteStatus = "INVITATION_ACCEPTED"
	InvitationExpired  InviteStatus = "INVITATION_EXPIRED"
	InviteInvalid      InviteStatus = "INVITE_INVALID"
)

var inviteStatusValues = []InviteStatus{InvitationSent, ApprovalRequired, InvitationAccepted, InvitationExpired}

// Values returns a slice of strings that represents all the possible values of the InviteStatus enum.
// Possible default values are "INVITATION_SENT", "APPROVAL_REQUIRED", "INVITATION_ACCEPTED", and "INVITATION_EXPIRED"
func (InviteStatus) Values() []string { return stringValues(inviteStatusValues) }

// String returns the invite status as a string
func (r InviteStatus) String() string { return string(r) }

// ToInviteStatus returns the invite status enum based on string input
func ToInviteStatus(r string) *InviteStatus { return parse(r, inviteStatusValues, &InviteInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r InviteStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *InviteStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
