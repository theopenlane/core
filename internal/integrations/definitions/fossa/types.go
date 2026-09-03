package fossa

import (
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the FOSSA integration definition
	definitionID = types.NewDefinitionRef("def_01K0FOSSA00000000000000001")
	// installation is the typed installation metadata handle for the FOSSA definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// fossaSchema is the credential schema for the FOSSA API token
	fossaSchema, fossaCredential = providerkit.CredentialSchema[CredentialSchema]()
	// fossaClient is the client ref for the FOSSA REST API client used by this definition
	fossaClient = types.NewClientRef[*APIClient]()
	// healthCheckSchema is the operation schema for the FOSSA health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// vulnerabilitySyncSchema is the operation schema for the FOSSA vulnerability sync operation
	vulnerabilitySyncSchema, vulnerabilitySyncOperation = providerkit.OperationSchema[VulnerabilitySync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// VulnerabilitySync includes the configuration for issues collected from FOSSA
	VulnerabilitySync VulnerabilitySync `json:"vulnerabilitySync,omitempty" jsonschema:"title=FOSSA Issue Sync"`
}

// VulnerabilitySync are the configuration settings for the FOSSA issue sync.
// There is deliberately no disable flag; security vulnerability collection is always on.
type VulnerabilitySync struct {
	// EnableLicenseFindings opts in to OSS license compliance issues, which are not collected by default
	EnableLicenseFindings bool `json:"enableLicenseFindings,omitempty" jsonschema:"title=Enable License Compliance Findings,description=Also collect FOSSA OSS license policy issues as findings. Security vulnerabilities are always collected."`
	// IncludeIgnored collects issues that have been dismissed in FOSSA in addition to active ones
	IncludeIgnored bool `json:"includeIgnored,omitempty" jsonschema:"title=Include Ignored Issues,description=Include issues that have been dismissed in FOSSA. By default only active issues are collected."`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.severity == 'critical' || payload.severity == 'high'"`
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
