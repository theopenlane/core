package engine

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
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
		Blocks: []map[string]any{{"type": "section"}},
	}

	config, err := buildNotificationOperationConfig(enums.ChannelSlack, preference, rendered)
	require.NoError(t, err)
	configMap, err := jsonx.ToMap(config)
	require.NoError(t, err)

	require.Equal(t, "C12345", configMap["channel"])
	require.Equal(t, "Body text", configMap["text"])
	require.Equal(t, "123.456", configMap["thread_ts"])
	require.NotNil(t, configMap["blocks"])
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
	configMap, err := jsonx.ToMap(config)
	require.NoError(t, err)

	require.Equal(t, "team-1", configMap["team_id"])
	require.Equal(t, "channel-2", configMap["channel_id"])
	require.Equal(t, "Body text", configMap["body"])
	require.Equal(t, "Subject line", configMap["subject"])
	require.Equal(t, "html", configMap["body_format"])
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

// TestDefinitionIDForNotificationChannel verifies channel to definition ID mapping
func TestDefinitionIDForNotificationChannel(t *testing.T) {
	id, err := definitionIDForNotificationChannel(enums.ChannelSlack)
	require.NoError(t, err)
	require.Equal(t, "def_01K0SLACK000000000000000001", string(id))

	_, err = definitionIDForNotificationChannel(enums.ChannelInvalid)
	require.Error(t, err)
}

// TestInstallationIDForRecord verifies installation id extraction from records
func TestInstallationIDForRecord(t *testing.T) {
	require.Equal(t, "", installationIDForRecord(nil))
	require.Equal(t, "int_123", installationIDForRecord(&generated.Integration{ID: "int_123"}))
}
