package enums

import "io"

// Channel represents the notification channel type
type Channel string

var (
	// ChannelInApp represents in-app notifications
	ChannelInApp Channel = "IN_APP"
	// ChannelSlack represents Slack notifications
	ChannelSlack Channel = "SLACK"
	// ChannelTeams represents Microsoft Teams notifications
	ChannelTeams Channel = "TEAMS"
	// ChannelEmail represents email notifications
	ChannelEmail Channel = "EMAIL"
	// ChannelInvalid represents an invalid channel
	ChannelInvalid Channel = "INVALID"
)

var channelValues = []Channel{ChannelInApp, ChannelSlack, ChannelTeams, ChannelEmail}

// Values returns a slice of strings that represents all the possible values of the Channel enum.
// Possible default values are "IN_APP", "SLACK", "TEAMS", and "EMAIL".
func (Channel) Values() []string { return stringValues(channelValues) }

// String returns the Channel as a string
func (r Channel) String() string { return string(r) }

// ToChannel returns the channel enum based on string input
func ToChannel(r string) *Channel { return parse(r, channelValues, &ChannelInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Channel) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Channel) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
