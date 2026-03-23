package microsoftteams

import (
	"time"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Microsoft Teams integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0MSTEAMS00000000000000001")
	// Installation is the typed installation metadata handle for the Microsoft Teams definition
	Installation = types.NewInstallationRef(resolveInstallationMetadata)

	// teamsCredentialSchema is the reflected JSON schema for the teams credential
	// teamsCredential is the auth-managed credential slot used by the Teams client
	teamsCredentialSchema, teamsCredential = providerkit.CredentialSchema[teamsCred]()

	// TeamsClient is the client ref for the Microsoft Graph service client used by this definition
	TeamsClient = types.NewClientRef[*msgraphsdk.GraphServiceClient]()

	// HealthDefaultOperation is the operation ref for the Microsoft Teams health check
	_, HealthDefaultOperation = providerkit.OperationSchema[HealthCheck]()
	// messageSendSchema is the reflected JSON schema for the message send operation config
	// MessageSendOperation is the operation ref for the Microsoft Teams message send operation
	messageSendSchema, MessageSendOperation = providerkit.OperationSchema[MessageOperationInput]()
)

// teamsCred holds the provider-owned credential material for a Microsoft Teams installation
type teamsCred struct {
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

// InstallationMetadata holds the stable Microsoft tenant identity for one Teams installation
type InstallationMetadata struct {
	// TenantID is the Microsoft Entra tenant identifier extracted from the access token when available
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}
