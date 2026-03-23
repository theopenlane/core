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
	// Installation is the typed installation metadata handle for the Slack definition
	Installation = types.NewInstallationRef(resolveInstallationMetadata)

	// slackCredential is the auth-managed credential slot used by the Slack client
	_, slackCredential = providerkit.CredentialSchema[slackCred]()

	// SlackClient is the client ref for the Slack Web API client used by this definition
	SlackClient = types.NewClientRef[*slackgo.Client]()

	// HealthDefaultOperation is the operation ref for the Slack health check
	_, HealthDefaultOperation = providerkit.OperationSchema[HealthCheck]()
	// messageSendSchema is the reflected JSON schema for the message send operation config
	// MessageSendOperation is the operation ref for the Slack message send operation
	messageSendSchema, MessageSendOperation = providerkit.OperationSchema[MessageOperationInput]()
	// DirectorySyncOperation is the operation ref for the directory account sync operation
	_, DirectorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// slackCredential holds the provider-owned credential material for a Slack installation
type slackCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// InstallationMetadata holds the stable Slack workspace identity for one installation
type InstallationMetadata struct {
	// TeamID is the Slack workspace identifier
	TeamID string `json:"teamId,omitempty" jsonschema:"title=Team ID"`
	// TeamName is the Slack workspace display name
	TeamName string `json:"teamName,omitempty" jsonschema:"title=Team Name"`
}
