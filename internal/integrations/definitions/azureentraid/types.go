package azureentraid

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Azure Entra ID integration definition
	definitionID = types.NewDefinitionRef("def_01K0AZENTRA0000000000000001")
	// installation is the typed installation metadata handle for the Azure Entra ID definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// entraTenantCredential is the auth-managed credential slot holding the consented tenant
	entraTenantCredential = types.CredentialRefOf[entraIDCred]()
	// entraCredential is the client ref for the Azure token credential used by the health check
	entraCredential = types.ClientRefOf[azcore.TokenCredential]().Using(entraTenantCredential)
	// entraClient is the client ref for the Microsoft Graph service client used by directory operations
	entraClient = types.ClientRefOf[*msgraphsdk.GraphServiceClient]().Using(entraTenantCredential)
	// entraConnection is the connection mode selected by the admin-consented tenant credential
	entraConnection = types.NewConnectionRef(entraTenantCredential).Enables(entraCredential).Enables(entraClient)
	// userInput is the installation user input layout, replacing the flat v1 layout
	userInput = types.NewUserInputRef[UserInput]("azureentraid").Replacing(types.NewUserInputRef[oldUserInput]("azureentraid-v1"), func(old oldUserInput) UserInput {
		return UserInput{PrimaryDirectory: old.PrimaryDirectory, DirectorySync: DirectorySync{DisableGroupSync: old.DisableGroupSync, IncludeGuestUsers: old.IncludeGuestUsers, FilterExpr: old.FilterExpr}}
	})
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// PrimaryDirectory marks this installation as the authoritative directory source for identity holder enrichment and lifecycle derivation
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
	// DirectorySync configures the directory sync operation
	DirectorySync DirectorySync `json:"directorySync,omitempty" jsonschema:"title=Directory Sync"`
}

// DirectorySync configures collection of Azure Entra ID directory users, groups, and memberships
type DirectorySync struct {
	// Disable switches the directory sync operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of users and groups from Azure Entra ID"`
	// DisableGroupSync when true only syncs users, skipping groups and memberships
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users from Azure Entra ID, disable groups sync operations"`
	// IncludeGuestUsers controls whether guest-type accounts are included in the sync
	IncludeGuestUsers bool `json:"includeGuestUsers,omitempty" jsonschema:"title=Include Guest Users"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
}

// oldUserInput is the flat v1 installation user input layout
type oldUserInput struct {
	// PrimaryDirectory marks this installation as the authoritative directory source for identity holder enrichment and lifecycle derivation
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
	// DisableGroupSync when true only syncs users, skipping groups and memberships
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users from Azure Entra ID, disable groups sync operations"`
	// IncludeGuestUsers controls whether guest-type accounts are included in the sync
	IncludeGuestUsers bool `json:"includeGuestUsers,omitempty" jsonschema:"title=Include Guest Users"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
}

// entraIDCred holds the per-installation credential for one Entra ID tenant
type entraIDCred struct {
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
