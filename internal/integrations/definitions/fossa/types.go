package fossa

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the FOSSA integration definition
	definitionID = types.NewDefinitionRef("def_01K0FOSSA00000000000000001")
	// installation is the typed installation metadata handle for the FOSSA definition
	installation = types.InstallationOf[InstallationMetadata]()
	// fossaConnection is the typed connection handle for the FOSSA API token
	fossaConnection = types.ConnectionOf[CredentialSchema]()
)

// VulnerabilitySync configures the FOSSA issue sync
type VulnerabilitySync struct {
	types.OperationSettings
	// EnableLicenseFindings opts in to OSS license compliance issues, which are not collected by default
	EnableLicenseFindings bool `json:"enableLicenseFindings,omitempty" jsonschema:"title=Enable License Compliance Findings,description=Also collect FOSSA OSS license policy issues as findings. Security vulnerabilities are always collected."`
}

// CredentialSchema holds the FOSSA API credentials for one installation
type CredentialSchema struct {
	// APIToken is the FOSSA API token used to authenticate requests
	APIToken string `json:"apiToken" jsonschema:"required,title=API Token,secret=true,description=Full API token from a FOSSA service account. Push-only tokens cannot read issues. Read scope is controlled by the service account role."`
	// BaseURL is the FOSSA API base URL, overridden only for on-premise deployments
	BaseURL string `json:"baseUrl,omitempty" jsonschema:"title=Base URL,description=FOSSA base URL. Leave blank to use https://app.fossa.com."`
}

// InstallationMetadata holds the stable FOSSA organization identity for one installation
type InstallationMetadata struct {
	// OrganizationID is the FOSSA organization identifier
	OrganizationID string `json:"organizationId,omitempty" jsonschema:"title=Organization ID"`
	// BaseURL is the FOSSA base URL used for this installation
	BaseURL string `json:"baseUrl,omitempty" jsonschema:"title=Base URL"`
	// Subscription is the FOSSA subscription tier reported for the organization
	Subscription string `json:"subscription,omitempty" jsonschema:"title=Subscription"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.OrganizationID,
	}
}
