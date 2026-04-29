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
	// slackClient is the client ref for the Slack Web API client used by this definition
	slackClient = types.NewClientRef[*slackgo.Client]()
	// healthDefaultOperation is the operation ref for the Slack health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema is the operation ref for the directory account sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
	// messageSendSchema is the operation ref for the Slack message send operation
	messageSendSchema, messageSendOperation = providerkit.OperationSchema[MessageSendOperation]()
)

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
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.TeamName,
		ExternalID:   m.TeamID,
	}
}
