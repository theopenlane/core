package slack

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSlackCredentialSetFromTokenExtras(t *testing.T) {
	token := (&oauth2.Token{AccessToken: "x"}).WithExtra(map[string]any{
		"app_id":      "A123",
		"bot_user_id": "U456",
		"team": map[string]any{
			"id":   "T789",
			"name": "Acme",
		},
	})

	set := slackCredentialSet(token)
	require.Equal(t, "A123", set.ProviderData["appId"])
	require.Equal(t, "U456", set.ProviderData["botUserId"])
	require.Equal(t, "T789", set.ProviderData["teamId"])
	require.Equal(t, "Acme", set.ProviderData["teamName"])
}

func TestSlackCredentialSetFromTopLevelTeamFields(t *testing.T) {
	token := (&oauth2.Token{AccessToken: "x"}).WithExtra(map[string]any{
		"team_id":   "T111",
		"team_name": "Workspace",
	})

	set := slackCredentialSet(token)
	require.Equal(t, "T111", set.ProviderData["teamId"])
	require.Equal(t, "Workspace", set.ProviderData["teamName"])
}

func TestSlackCredentialSetEmptyWhenNoMetadata(t *testing.T) {
	set := slackCredentialSet((&oauth2.Token{AccessToken: "x"}).WithExtra(map[string]any{}))
	require.Nil(t, set.ProviderData)
}
