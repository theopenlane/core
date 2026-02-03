package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the Channel enum.
// Possible default values are "IN_APP", "SLACK", "TEAMS", and "EMAIL".
func (Channel) Values() (kinds []string) {
	for _, s := range []Channel{ChannelInApp, ChannelSlack, ChannelTeams, ChannelEmail} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the Channel as a string
func (r Channel) String() string {
	return string(r)
}

// ToChannel returns the channel enum based on string input
func ToChannel(r string) *Channel {
	switch r := strings.ToUpper(r); r {
	case ChannelInApp.String():
		return &ChannelInApp
	case ChannelSlack.String():
		return &ChannelSlack
	case ChannelTeams.String():
		return &ChannelTeams
	case ChannelEmail.String():
		return &ChannelEmail
	default:
		return &ChannelInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Channel) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Channel) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Channel, got: %T", v) //nolint:err113
	}

	*r = Channel(str)

	return nil
}
