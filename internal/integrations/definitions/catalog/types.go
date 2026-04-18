package catalog

import (
	"github.com/theopenlane/core/internal/integrations/definitions/azureentraid"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/definitions/googleworkspace"
	"github.com/theopenlane/core/internal/integrations/definitions/microsoftteams"
	"github.com/theopenlane/core/internal/integrations/definitions/oidclocal"
	"github.com/theopenlane/core/internal/integrations/definitions/slack"
)

// Config aggregates the definitions configuration structs (for when definitions require operator-held credentials or other config)
// this is here purely to drive the konaf output config constructs we use as a standard in this repo
type Config struct {
	// GitHubApp holds operator credentials for the GitHub App definition
	GitHubApp githubapp.Config `json:"githubapp" koanf:"githubapp"`
	// Slack holds OAuth credentials for the Slack definition
	Slack slack.Config `json:"slack" koanf:"slack"`
	// SlackRuntime holds runtime-provisioned configuration for system Slack notifications
	SlackRuntime slack.RuntimeSlackConfig `json:"slackRuntime" koanf:"slackRuntime"`
	// GoogleWorkspace holds OAuth credentials for the Google Workspace definition
	GoogleWorkspace googleworkspace.Config `json:"googleworkspace" koanf:"googleworkspace"`
	// AzureEntraID holds OAuth credentials for the Azure Entra ID definition
	AzureEntraID azureentraid.Config `json:"azureentraid" koanf:"azureentraid"`
	// MicrosoftTeams holds OAuth credentials for the Microsoft Teams definition
	MicrosoftTeams microsoftteams.Config `json:"microsoftteams" koanf:"microsoftteams"`
	// OIDCLocal holds local Dex-backed OIDC settings for end-to-end integration auth testing
	OIDCLocal oidclocal.Config `json:"oidclocal" koanf:"oidclocal"`
	// Email holds operator-level email integration configuration
	Email email.RuntimeEmailConfig `json:"email" koanf:"email"`
}
