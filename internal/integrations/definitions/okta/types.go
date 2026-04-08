package okta

import (
	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Okta integration definition
	definitionID = types.NewDefinitionRef("def_01K0OKTA0000000000000000001")
	// integration is the typed installation metadata handle for the Okta definition
	integration = types.NewInstallationRef(resolveInstallationMetadata)
	// oktaCredential is the auth-managed credential slot used by the Okta client
	oktaCredentialSchema, oktaCredential = providerkit.CredentialSchema[CredentialSchema]()
	// oktaClient is the client ref for the Okta API client used by this definition
	oktaClient = types.NewClientRef[*oktagosdk.APIClient]()
	// healthCheckSchema is the operation ref for the Okta health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema is the operation ref for the Okta directory sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
	// Search is an optional Okta search expression applied server-side when listing users
	Search string `json:"search,omitempty" jsonschema:"title=User Search Expression,description=Optional Okta search expression for filtering users (e.g. profile.department eq \"Engineering\")."`
	// EnableGroupSync controls whether group and membership records are collected
	EnableGroupSync bool `json:"enableGroupSync,omitempty" jsonschema:"title=Sync Groups"`
	// PrimaryDirectory marks this installation as the authoritative directory source for identity holder enrichment and lifecycle derivation
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}

// CredentialSchema holds the Okta tenant credentials for one installation
type CredentialSchema struct {
	// OrgURL is the Okta organization URL
	OrgURL string `json:"orgUrl"   jsonschema:"required,title=Org URL"`
	// APIToken is the Okta API token with permissions to read tenant and policy metadata
	APIToken string `json:"apiToken" jsonschema:"required,title=API Token"`
}

// InstallationMetadata holds the stable Okta tenant identity for one installation
type InstallationMetadata struct {
	// OrgURL is the Okta organization URL configured for this installation
	OrgURL string `json:"orgUrl,omitempty" jsonschema:"title=Org URL"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.OrgURL,
	}
}
