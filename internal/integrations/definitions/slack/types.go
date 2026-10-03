package slack

import (
	"time"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Slack integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SLACK000000000000000001")
	// installation is the typed installation metadata handle for the Slack definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// slackCredential is the auth-managed credential slot used by the OAuth connection
	slackCredential = types.CredentialRefOf[slackCred]()
	// slackBotTokenCredential is the credential slot for user-provisioned bot tokens
	slackBotTokenCredential = types.CredentialRefOf[slackBotTokenCred]()
	// slackClient is the unified client ref for every Slack operation
	slackClient = types.ClientRefOf[*SlackClient]()
	// userInput is the installation user input layout for the Slack definition
	userInput = types.UserInputRefOf[UserInput]()
)

// RuntimeSlackConfig is the runtime-provisioned configuration for the system Slack integration
type RuntimeSlackConfig struct {
	// WebhookURL is the Slack incoming webhook URL used to deliver system notifications
	WebhookURL string `json:"webhookURL,omitempty" koanf:"webhookURL" jsonschema:"description=Slack incoming webhook URL for fire-and-forget system notifications"`
	// BotToken is a Slack Bot User OAuth Token (xoxb-...) for the platform-owned workspace
	BotToken string `json:"botToken,omitempty" koanf:"botToken" jsonschema:"description=Bot User OAuth Token for full Web API access to the platform workspace"`
	// DefaultChannel is the channel id used for system messages when no explicit channel is specified
	DefaultChannel string `json:"defaultChannel,omitempty" koanf:"defaultChannel" jsonschema:"description=Default channel id for system messages when no explicit channel is provided"`
}

// Provisioned reports whether the runtime config has the minimum required fields
func (c RuntimeSlackConfig) Provisioned() bool {
	return c.WebhookURL != "" || c.BotToken != ""
}

// SlackClient is the unified Slack client used by every Slack operation
type SlackClient struct { //nolint:revive
	// API is the Slack Web API client (present for bot-token runtime and customer installations)
	API *slackgo.Client
	// WebhookURL is the Slack incoming webhook used as a fallback when no API client is configured
	WebhookURL string
	// DefaultChannel is the channel id used for system messages when no explicit channel is specified
	DefaultChannel string
	// devMode silently drops messages when no transport is configured
	devMode bool
}

// slackCred holds the provider-owned credential material for an OAuth-based Slack installation
type slackCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
}

// slackBotTokenCred holds a user-provisioned bot token for a Slack installation
type slackBotTokenCred struct {
	// BotToken is a Slack Bot User OAuth Token (xoxb-...) created by the user in their Slack app
	BotToken string `json:"botToken" jsonschema:"required,title=Bot Token,description=Bot User OAuth Token from your Slack app (starts with xoxb-)"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// DefaultMessaging marks this installation as the preferred workspace for messaging
	DefaultMessaging bool `json:"defaultMessaging,omitempty" jsonschema:"title=Default Messaging"`
}

// DirectorySync is the Slack directory account sync operation configuration
type DirectorySync struct {
	types.OperationSettings
}

// InstallationMetadata holds the stable Slack workspace identity for one installation
type InstallationMetadata struct {
	// TeamID is the Slack workspace identifier
	TeamID string `json:"teamId,omitempty" jsonschema:"title=Team ID"`
	// TeamName is the Slack workspace display name
	TeamName string `json:"teamName,omitempty" jsonschema:"title=Team Name"`
	// DefaultChannel is the Slack channel id used for system notifications
	DefaultChannel string `json:"defaultChannel,omitempty" jsonschema:"title=Default Channel"`
}

// InstallationInput is the provider-defined input supplied when installing the Slack integration
type InstallationInput struct {
	// DefaultChannel is the Slack channel id used as the default delivery target for system messages
	DefaultChannel string `json:"defaultChannel,omitempty" jsonschema:"title=Default Channel,description=Slack channel id used as the default delivery target for system notifications"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.TeamName,
		ExternalID:   m.TeamID,
	}
}
