package okta

import (
	oktagosdk "github.com/okta/okta-sdk-golang/v6/okta"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Okta integration definition
	definitionID = types.NewDefinitionRef("def_01K0OKTA0000000000000000001")
	// installation is the typed installation metadata handle for the Okta definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// oktaCredential is the typed credential slot used by the Okta client
	oktaCredential = types.CredentialRefOf[CredentialSchema]()
	// oktaClient is the client ref for the Okta API client used by this definition
	oktaClient = types.ClientRefOf[*oktagosdk.APIClient]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// PrimaryDirectory marks this installation as the authoritative directory source
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}

// DirectorySync configures collection of Okta directory users, groups, and memberships
type DirectorySync struct {
	types.OperationSettings
	// Search is an optional Okta search expression applied server-side when listing users
	Search string `json:"search,omitempty" jsonschema:"title=User Search Expression,description=Optional Okta search expression for filtering users (e.g. profile.department eq \"Engineering\")."`
	// EnableGroupSync controls whether group and membership records are collected
	EnableGroupSync bool `json:"enableGroupSync,omitempty" jsonschema:"title=Sync Groups"`
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
	// OrgID is the immutable Okta organization identifier resolved from the org settings API
	OrgID string `json:"orgId,omitempty" jsonschema:"title=Org ID"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.OrgURL,
		ExternalID:   m.OrgID,
	}
}
