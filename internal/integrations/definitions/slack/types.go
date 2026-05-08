package slack

import (
	"time"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Slack integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SLACK000000000000000001")
	// installation is the typed installation metadata handle for the Slack definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// slackCredential is the auth-managed credential slot used by the OAuth connection
	_, slackCredential = providerkit.CredentialSchema[slackCred]()
	// slackBotTokenCredential is the credential slot for user-provisioned bot tokens
	slackBotTokenCredentialSchema, slackBotTokenCredential = providerkit.CredentialSchema[slackBotTokenCred]()
	// slackClient is the unified client ref for every Slack operation; runtime and customer
	// paths both build a SlackClient that wraps the Web API client and any system-notification transport
	slackClient = types.NewClientRef[*SlackClient]()
	// runtimeSlackSchema is the JSON schema and typed ref for the runtime Slack config
	runtimeSlackSchema, runtimeSlackRef = providerkit.RuntimeSchema[RuntimeSlackConfig]()
	// healthDefaultOperation is the operation ref for the Slack health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema is the operation ref for the directory account sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
	// messageSendSchema is the operation ref for the Slack message send operation
	messageSendSchema, MessageSendOp = providerkit.OperationSchema[MessageSendOperation]() //nolint:revive // co-initialized with schema
)

// RuntimeSlackConfig is the runtime-provisioned configuration for the system Slack integration.
// Sourced from koanf/environment at startup; populated for the platform-owned workspace.
// Two modes are supported: webhook-only (fire-and-forget via incoming webhook URL) and
// bot-token mode (full Web API access with channel targeting and Block Kit support)
type RuntimeSlackConfig struct {
	// WebhookURL is the Slack incoming webhook URL used to deliver system notifications
	WebhookURL string `json:"webhookURL,omitempty" koanf:"webhookURL" jsonschema:"description=Slack incoming webhook URL for fire-and-forget system notifications"`
	// BotToken is a Slack Bot User OAuth Token (xoxb-...) for the platform-owned workspace
	BotToken string `json:"botToken,omitempty" koanf:"botToken" jsonschema:"description=Bot User OAuth Token for full Web API access to the platform workspace"`
	// DefaultChannel is the channel id used for system messages when no explicit channel is specified
	DefaultChannel string `json:"defaultChannel,omitempty" koanf:"defaultChannel" jsonschema:"description=Default channel id for system messages when no explicit channel is provided"`
}

// Provisioned reports whether the runtime config has the minimum required fields to deliver system messages
func (c RuntimeSlackConfig) Provisioned() bool {
	return c.WebhookURL != "" || c.BotToken != ""
}

// SlackClient is the unified Slack client used by every Slack operation. Both runtime and
// customer paths produce a SlackClient; the active transport depends on which fields are populated:
// API (bot token or OAuth) enables chat.postMessage with channel targeting; WebhookURL provides
// fire-and-forget delivery when no API client is available
type SlackClient struct { //nolint:revive
	// API is the Slack Web API client (present for bot-token runtime and customer installations)
	API *slackgo.Client
	// WebhookURL is the Slack incoming webhook used as a fallback when no API client is configured
	WebhookURL string
	// DefaultChannel is the channel id used for system messages when no explicit channel is specified
	DefaultChannel string
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
	// DefaultMessaging marks this installation as the preferred Slack workspace for workflow messaging operations
	DefaultMessaging bool `json:"defaultMessaging,omitempty" jsonschema:"title=Default Messaging"`
	// DirectorySync includes the configuration for identity accounts from Slack members
	DirectorySync DirectorySync `json:"directorySync,omitempty" jsonschema:"title=Directory Account Sync"`
}

type DirectorySync struct {
	// Disable is used to disable the directory sync operation from GitHub
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of users from Slack"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting.,example=Example: payload.is_external == false'"`
}

// InstallationMetadata holds the stable Slack workspace identity for one installation
type InstallationMetadata struct {
	// TeamID is the Slack workspace identifier
	TeamID string `json:"teamId,omitempty" jsonschema:"title=Team ID"`
	// TeamName is the Slack workspace display name
	TeamName string `json:"teamName,omitempty" jsonschema:"title=Team Name"`
	// DefaultChannel is the Slack channel id used for system notifications on the customer installation
	DefaultChannel string `json:"defaultChannel,omitempty" jsonschema:"title=Default Channel"`
}

// InstallationInput is the provider-defined input supplied when installing the Slack integration.
// Bot-token connections populate it at install time; OAuth connections leave it empty for now
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
