package slack

import (
	"time"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Slack integration definition
	definitionID = types.NewDefinitionRef("def_01K0SLACK000000000000000001")
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
	messageSendSchema, messageSendOperation = providerkit.OperationSchema[MessageSendOperation]()
)

// RuntimeSlackConfig is the runtime-provisioned configuration for the system Slack integration.
// Sourced from koanf/environment at startup; populated for the platform-owned workspace
type RuntimeSlackConfig struct {
	// WebhookURL is the Slack incoming webhook URL used to deliver system notifications
	WebhookURL string `json:"webhookURL,omitempty" koanf:"webhookURL" jsonschema:"description=Slack incoming webhook URL used for system notifications"`
}

// Provisioned reports whether the runtime config has the minimum required fields to deliver system messages
func (c RuntimeSlackConfig) Provisioned() bool {
	return c.WebhookURL != ""
}

// SlackClient is the unified Slack client used by every Slack operation. The runtime (system)
// path populates only WebhookURL; customer installations populate API and optionally DefaultChannel
type SlackClient struct {
	// API is the Slack Web API client used by customer-installed operations
	API *slackgo.Client
	// WebhookURL is the Slack incoming webhook used by the runtime system-notification path
	WebhookURL string
	// DefaultChannel is the channel id customer installations use when a system-message op supplies no explicit recipient
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
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
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
