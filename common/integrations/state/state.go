package state

// IntegrationProviderState stores provider-specific integration state captured during auth/config.
type IntegrationProviderState struct {
	// GitHub contains the GitHub integration state
	GitHub *GitHubState `json:"github,omitempty"`
	// Slack contains the Slack integration state
	Slack *SlackState `json:"slack,omitempty"`
}

// GitHubState captures GitHub App installation details for an integration.
type GitHubState struct {
	// AppID is the GitHub App ID
	AppID string `json:"appId,omitempty"`
	// InstallationID is the GitHub App installation ID
	InstallationID string `json:"installationId,omitempty"`
}

// SlackState captures Slack workspace details for an integration.
type SlackState struct {
	// AppID is the Slack App ID
	AppID string `json:"appId,omitempty"`
	// TeamID is the Slack workspace team ID
	TeamID string `json:"teamId,omitempty"`
	// TeamName is the Slack workspace team name
	TeamName string `json:"teamName,omitempty"`
	// BotUserID is the Slack bot user ID
	BotUserID string `json:"botUserId,omitempty"`
}
