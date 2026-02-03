package helpers

import (
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
)

// ProviderStateFromProviderData builds provider state from persisted provider data.
func ProviderStateFromProviderData(provider types.ProviderType, data map[string]any) *openapi.IntegrationProviderState {
	if len(data) == 0 {
		return nil
	}

	switch provider {
	case types.ProviderType("github"), types.ProviderType("github_app"):
		state := gitHubStateFromProviderData(data)
		if state == nil {
			return nil
		}
		return &openapi.IntegrationProviderState{GitHub: state}
	case types.ProviderType("slack"):
		state := slackStateFromProviderData(data)
		if state == nil {
			return nil
		}
		return &openapi.IntegrationProviderState{Slack: state}
	default:
		return nil
	}
}

// MergeProviderState overlays non-empty values from update onto current.
func MergeProviderState(current openapi.IntegrationProviderState, update *openapi.IntegrationProviderState) openapi.IntegrationProviderState {
	if update == nil {
		return current
	}

	if update.GitHub != nil {
		if current.GitHub == nil {
			current.GitHub = &openapi.IntegrationGitHubState{}
		}
		mergeGitHubState(current.GitHub, update.GitHub)
	}

	if update.Slack != nil {
		if current.Slack == nil {
			current.Slack = &openapi.IntegrationSlackState{}
		}
		mergeSlackState(current.Slack, update.Slack)
	}

	return current
}

func gitHubStateFromProviderData(data map[string]any) *openapi.IntegrationGitHubState {
	if len(data) == 0 {
		return nil
	}

	state := openapi.IntegrationGitHubState{
		AppID:          FirstStringValue(data, "appId", "app_id"),
		InstallationID: FirstStringValue(data, "installationId", "installation_id"),
	}
	if isGitHubStateEmpty(state) {
		return nil
	}

	return &state
}

func slackStateFromProviderData(data map[string]any) *openapi.IntegrationSlackState {
	if len(data) == 0 {
		return nil
	}

	state := openapi.IntegrationSlackState{
		AppID:     FirstStringValue(data, "appId", "app_id"),
		TeamID:    FirstStringValue(data, "teamId", "team_id"),
		TeamName:  FirstStringValue(data, "teamName", "team_name"),
		BotUserID: FirstStringValue(data, "botUserId", "bot_user_id"),
	}
	if isSlackStateEmpty(state) {
		return nil
	}

	return &state
}

func mergeGitHubState(dst *openapi.IntegrationGitHubState, src *openapi.IntegrationGitHubState) {
	if dst == nil || src == nil {
		return
	}
	if src.AppID != "" {
		dst.AppID = src.AppID
	}
	if src.InstallationID != "" {
		dst.InstallationID = src.InstallationID
	}
}

func mergeSlackState(dst *openapi.IntegrationSlackState, src *openapi.IntegrationSlackState) {
	if dst == nil || src == nil {
		return
	}
	if src.AppID != "" {
		dst.AppID = src.AppID
	}
	if src.TeamID != "" {
		dst.TeamID = src.TeamID
	}
	if src.TeamName != "" {
		dst.TeamName = src.TeamName
	}
	if src.BotUserID != "" {
		dst.BotUserID = src.BotUserID
	}
}

func isGitHubStateEmpty(state openapi.IntegrationGitHubState) bool {
	return state.AppID == "" &&
		state.InstallationID == ""
}

func isSlackStateEmpty(state openapi.IntegrationSlackState) bool {
	return state.AppID == "" &&
		state.TeamID == "" &&
		state.TeamName == "" &&
		state.BotUserID == ""
}
