package azureentraid

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Azure Entra ID integration definition
	definitionID = types.NewDefinitionRef("def_01K0AZENTRA0000000000000001")
	// installation is the typed installation metadata handle for the Azure Entra ID definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// entraTenantSchema is the cred schema for entraID
	entraTenantSchema, entraTenantCredential = providerkit.CredentialSchema[entraIDCred]()
	// EntraCredential is the client ref for the Azure token credential used by the health check
	entraCredential = types.NewClientRef[azcore.TokenCredential]()
	// EntraClient is the client ref for the Microsoft Graph service client used by directory operations
	entraClient = types.NewClientRef[*msgraphsdk.GraphServiceClient]()
	// healthDefaultOperation is the operation ref for the Azure Entra ID health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// DirectorySyncOperation is the operation ref for the Azure Entra ID directory sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
	// EnableGroupSync controls whether group and membership records are collected
	EnableGroupSync bool `json:"enableGroupSync,omitempty" jsonschema:"title=Sync Groups"`
	// IncludeGuestUsers controls whether guest-type accounts are included in the sync
	IncludeGuestUsers bool `json:"includeGuestUsers,omitempty" jsonschema:"title=Include Guest Users"`
	// PrimaryDirectory marks this installation as the authoritative directory source for identity holder enrichment and lifecycle derivation
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}

// entraIDCred holds the provider-owned credential material for an Azure Entra ID installation
type entraIDCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
	// TenantID is the Azure AD tenant identifier extracted from OIDC claims
	TenantID string `json:"tenantId"`
}

// CredentialSchema holds the per-installation credential for one Entra ID tenant
type CredentialSchema struct {
	// TenantID is the Azure Active Directory tenant identifier for this installation
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID"`
}

// VerifiedDomain holds one verified domain entry for an Azure Entra ID tenant
type VerifiedDomain struct {
	// Name is the domain name
	Name string `json:"name,omitempty"`
	// IsDefault indicates whether this is the default domain for the tenant
	IsDefault bool `json:"isDefault,omitempty"`
}

// InstallationMetadata holds the stable Azure Entra tenant identity for one installation
type InstallationMetadata struct {
	// TenantID is the Azure Active Directory tenant identifier selected during setup
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
	// DisplayName is the organization display name from Microsoft Graph
	DisplayName string `json:"displayName,omitempty" jsonschema:"title=Display Name"`
	// VerifiedDomains is the list of verified domains for the tenant
	VerifiedDomains []VerifiedDomain `json:"verifiedDomains,omitempty" jsonschema:"title=Verified Domains"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.DisplayName,
		ExternalID:   m.TenantID,
	}
}
