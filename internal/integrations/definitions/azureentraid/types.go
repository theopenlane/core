package azureentraid

import (
	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Azure Entra ID integration definition
	definitionID = types.NewDefinitionRef("def_01K0AZENTRA0000000000000001")
	// adminConsent is the connection handle for the consented tenant credential under the current app registration
	adminConsent = types.ConnectionOf[entraIDCred]().Revision(2)
	// legacyConsent is the retired connection for tenants consented to the app registration used before the tenant move
	legacyConsent = types.ConnectionOf[entraIDCred]()
	// installation is the typed installation metadata handle for the Azure Entra ID definition
	installation = types.InstallationOf[InstallationMetadata]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// PrimaryDirectory marks this installation as the authoritative directory source
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}

// DirectorySync configures collection of Azure Entra ID directory users, groups, and memberships
type DirectorySync struct {
	providerkit.DirectorySync
	// IncludeGuestUsers controls whether guest-type accounts are included in the sync
	IncludeGuestUsers bool `json:"includeGuestUsers,omitempty" jsonschema:"title=Include Guest Users"`
}

// entraIDCred holds the per-installation credential for one Entra ID tenant
type entraIDCred struct {
	// TenantID is the Azure Active Directory tenant identifier for this installation
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID"`
}

// newEntraIDCred builds the Entra ID credential for a consented tenant
func newEntraIDCred(consent microsoft.AdminConsent) entraIDCred {
	return entraIDCred{TenantID: consent.TenantID}
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
