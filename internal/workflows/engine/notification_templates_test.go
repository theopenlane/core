package engine

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

// TestBuildNotificationOperationConfigSlack verifies Slack config generation
func TestBuildNotificationOperationConfigSlack(t *testing.T) {
	preference := &generated.NotificationPreference{
		Destination: "C12345",
		Config: map[string]any{
			"thread_ts": "123.456",
		},
	}
	rendered := &renderedNotificationTemplate{
		Title:  "Title",
		Body:   "Body text",
		Blocks: []any{map[string]any{"type": "section"}},
	}

	config, err := buildNotificationOperationConfig(enums.ChannelSlack, preference, rendered)
	require.NoError(t, err)

	require.Equal(t, "C12345", config["channel"])
	require.Equal(t, "Body text", config["text"])
	require.Equal(t, "123.456", config["thread_ts"])
	require.NotNil(t, config["blocks"])
}

// TestBuildNotificationOperationConfigTeams verifies Teams config generation
func TestBuildNotificationOperationConfigTeams(t *testing.T) {
	preference := &generated.NotificationPreference{
		Destination: "team-1:channel-2",
	}
	rendered := &renderedNotificationTemplate{
		Title:   "Title",
		Body:    "Body text",
		Subject: "Subject line",
		Template: &generated.NotificationTemplate{
			Format: enums.NotificationTemplateFormatHTML,
		},
	}

	config, err := buildNotificationOperationConfig(enums.ChannelTeams, preference, rendered)
	require.NoError(t, err)

	require.Equal(t, "team-1", config["team_id"])
	require.Equal(t, "channel-2", config["channel_id"])
	require.Equal(t, "Body text", config["body"])
	require.Equal(t, "Subject line", config["subject"])
	require.Equal(t, "html", config["body_format"])
}

// TestResolveTeamsDestination verifies destination parsing for Teams
func TestResolveTeamsDestination(t *testing.T) {
	teamID, channelID := resolveTeamsDestination(&generated.NotificationPreference{Destination: "team/chan"}, map[string]any{})
	require.Equal(t, "team", teamID)
	require.Equal(t, "chan", channelID)
}

// TestOperationNameForChannel verifies channel to operation mapping
func TestOperationNameForChannel(t *testing.T) {
	name, err := operationNameForChannel(enums.ChannelSlack)
	require.NoError(t, err)
	require.Equal(t, "message.send", string(name))

	_, err = operationNameForChannel(enums.ChannelInvalid)
	require.Error(t, err)
}
