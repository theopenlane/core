package cloudflare

import (
	cf "github.com/cloudflare/cloudflare-go/v6"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Cloudflare integration definition
	definitionID = types.NewDefinitionRef("def_01K0CFLARE00000000000000001")
	// installation is the typed installation metadata handle for the Cloudflare definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// cloudflareSchema is the credential schema for the Cloudflare integration definition
	cloudflareSchema, cloudflareCredential = providerkit.CredentialSchema[CredentialSchema]()
	// cloudflareClient is the client ref for the Cloudflare API client used by this definition
	cloudflareClient = types.NewClientRef[*cf.Client]()
	// healthDefaultOperation is the operation ref for the Cloudflare health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema is the operation ref for the directory account sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// AccountID is the Cloudflare account identifier used for account-scoped API calls
	AccountID string `json:"accountId,omitempty" jsonschema:"required,title=Account ID,description=Cloudflare account ID required for listing account members."`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
}

// CredentialSchema holds the Cloudflare API credentials for one installation
type CredentialSchema struct {
	// APIToken is the Cloudflare API token with permissions to read account and zone metadata
	APIToken string `json:"apiToken"          jsonschema:"required,title=API Token"`
}

// InstallationMetadata holds the stable Cloudflare account identity for one installation
type InstallationMetadata struct {
	// AccountID is the Cloudflare account identifier used for account-scoped collection
	AccountID string `json:"accountId,omitempty" jsonschema:"title=Account ID"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.AccountID,
	}
}
