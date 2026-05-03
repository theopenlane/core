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
	// installation is the typed installation metadata handle for the Microsoft Teams definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// teamsCredential is the auth-managed credential slot used by the Teams client
	teamsCredentialSchema, teamsCredential = providerkit.CredentialSchema[teamsCred]()
	// teamsClient is the client ref for the Microsoft Graph service client used by this definition
	teamsClient = types.NewClientRef[*msgraphsdk.GraphServiceClient]()
	// healthDefaultOperation is the operation ref for the Microsoft Teams health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// messageSendSchema is the operation ref for the Microsoft Teams message send operation
	messageSendSchema, MessageSendOp = providerkit.OperationSchema[MessageSendOperation]()
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
	// DefaultMessaging marks this installation as the preferred Teams tenant for workflow messaging operations
	DefaultMessaging bool `json:"defaultMessaging,omitempty" jsonschema:"title=Default Messaging"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
}

// InstallationMetadata holds the stable Microsoft tenant identity for one Teams installation
type InstallationMetadata struct {
	// TenantID is the Microsoft Entra tenant identifier extracted from the access token when available
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.TenantID,
	}
}
