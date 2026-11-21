package aikido

import (
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Aikido integration definition
	definitionID = types.NewDefinitionRef("def_01K0AIKIDO00000000000000001")
	// installation is the typed installation metadata handle for the Aikido definition
	installation = types.InstallationOf[InstallationMetadata]()
	// oauthConnection is the typed connection handle for the Openlane Aikido OAuth app
	oauthConnection = types.ConnectionOf[aikidoOAuthCred]()
	// aikidoConnection is the typed connection handle for the Aikido client credentials
	aikidoConnection = types.ConnectionOf[CredentialSchema]()
)

const (
	// oauthAuthorizePath is the OAuth authorization endpoint on the app host
	oauthAuthorizePath = "/oauth/authorize"
	// oauthAppURL is the app host serving the OAuth endpoints, the only host Aikido documents for them
	oauthAppURL = "https://app.aikido.dev"
)

const (
	// regionEU is the default Aikido region, hosted at app.aikido.dev
	regionEU = "eu"
	// regionUS is the Aikido United States region
	regionUS = "us"
	// regionAU is the Aikido Australia region
	regionAU = "au"
	// regionME is the Aikido Middle East region
	regionME = "me"
)

// regionAppURLs maps each Aikido region to its app host
var regionAppURLs = map[string]string{
	regionEU: "https://app.aikido.dev",
	regionUS: "https://app.us.aikido.dev",
	regionAU: "https://app.au.aikido.dev",
	regionME: "https://app.me.aikido.dev",
}

// AssetSync configures the Aikido asset inventory sync
type AssetSync struct {
	types.OperationSettings
}

// IssueSync configures the Aikido issue sync
type IssueSync struct {
	types.OperationSettings
}

// AutofixSync configures the Aikido autofix sync
type AutofixSync struct {
	types.OperationSettings
}

// ComplianceSync configures the Aikido compliance report sync
type ComplianceSync struct {
	types.OperationSettings
}

// ScanSync configures the Aikido AI pentest and code analysis assessment sync
type ScanSync struct {
	types.OperationSettings
}

// aikidoOAuthCred holds the OAuth token material for an installation connected through the Openlane Aikido app
type aikidoOAuthCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the access token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
}

// CredentialSchema holds the Aikido REST API client credentials for one installation
type CredentialSchema struct {
	// ClientID is the Aikido REST API client ID
	ClientID string `json:"clientId" jsonschema:"required,title=Client ID,description=Client ID from Aikido Settings > Integrations > REST API."`
	// ClientSecret is the Aikido REST API client secret
	ClientSecret string `json:"clientSecret" jsonschema:"required,title=Client Secret,secret=true,description=Client secret paired with the client ID."`
	// Region is the Aikido region hosting the workspace
	Region string `json:"region,omitempty" jsonschema:"title=Region,description=Region hosting your Aikido workspace. Leave blank for the default EU region (app.aikido.dev).,enum=eu,enum=us,enum=au,enum=me,default=eu"`
}

// InstallationMetadata holds the stable Aikido workspace identity for one installation
type InstallationMetadata struct {
	// WorkspaceID is the Aikido workspace identifier
	WorkspaceID string `json:"workspaceId,omitempty" jsonschema:"title=Workspace ID"`
	// WorkspaceName is the Aikido workspace name
	WorkspaceName string `json:"workspaceName,omitempty" jsonschema:"title=Workspace"`
	// Region is the Aikido region hosting the workspace
	Region string `json:"region,omitempty" jsonschema:"title=Region"`
	// LinkedProvider is the source control provider linked to the workspace
	LinkedProvider string `json:"linkedProvider,omitempty" jsonschema:"title=Linked Provider"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID:   m.WorkspaceID,
		ExternalName: m.WorkspaceName,
	}
}
