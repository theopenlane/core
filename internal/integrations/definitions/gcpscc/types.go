package gcpscc

import (
	"encoding/json"
	"strings"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the GCP Security Command Center integration definition
	definitionID = types.NewDefinitionRef("def_01K0GCPSCC00000000000000001")
	// installation is the typed installation metadata handle for the GCP Security Command Center definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// sccSchema is the reflected JSON schema for the GCP SCC credential
	// sccCredential is the credential slot used by the SCC client
	sccSchema, sccCredential = providerkit.CredentialSchema[CredentialSchema]()
	// sccClient is the client ref for the GCP Security Command Center client used by this definition
	sccClient                                       = types.NewClientRef[*cloudscc.Client]()
	healthCheckSchema, healthCheckOperation         = providerkit.OperationSchema[HealthCheck]()
	findingsCollectSchema, findingsCollectOperation = providerkit.OperationSchema[FindingsConfig]()
)

const (
	// projectScopeAll indicates collection should target all GCP projects in the organization
	projectScopeAll = "all"
	// projectScopeSpecific indicates collection should target only the explicitly listed project IDs
	projectScopeSpecific = "specific"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// CredentialSchema holds the GCP SCC credentials for one installation
type CredentialSchema struct {
	// OrganizationID is the GCP organization identifier
	OrganizationID string `json:"organizationId,omitempty" jsonschema:"title=Organization ID"`
	// ProjectID is the fallback GCP project identifier used for quota and parent resolution
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID"`
	// ProjectScope controls whether SCC collection targets all or specific projects
	ProjectScope string `json:"projectScope,omitempty" jsonschema:"title=Project Scope"`
	// ProjectIDs lists the specific GCP projects used when project scope is specific
	ProjectIDs []string `json:"projectIds,omitempty" jsonschema:"title=Project IDs"`
	// SourceID is the default SCC source identifier used when a run does not override it
	SourceID string `json:"sourceId,omitempty" jsonschema:"title=SCC Source ID"`
	// SourceIDs lists the SCC source identifiers used for collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"title=SCC Source IDs"`
	// Scopes lists the OAuth scopes requested for service account credentials
	Scopes []string `json:"scopes,omitempty" jsonschema:"title=OAuth Scopes"`
	// ServiceAccountKey is the service account key JSON used for direct credentials
	ServiceAccountKey string `json:"serviceAccountKey,omitempty" jsonschema:"title=Service Account Key JSON"`
}

// applyDefaults fills in fallback values for missing optional fields
func (m CredentialSchema) applyDefaults() CredentialSchema {
	normalized := m
	if normalized.ProjectScope == "" {
		normalized.ProjectScope = projectScopeAll
	}

	normalized.ServiceAccountKey = normalizeServiceAccountKey(normalized.ServiceAccountKey)

	return normalized
}

// normalizeServiceAccountKey trims and unwraps JSON-encoded service account keys
func normalizeServiceAccountKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var decoded string
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return strings.TrimSpace(decoded)
	}

	return trimmed
}

// InstallationMetadata holds the stable GCP organization and service account identity for one installation
type InstallationMetadata struct {
	// OrganizationID is the GCP organization identifier when collection is organization-scoped
	OrganizationID string `json:"organizationId,omitempty" jsonschema:"title=Organization ID"`
	// ProjectID is the primary GCP project identifier used for quota or fallback parent resolution
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID"`
	// ProjectScope indicates whether collection targets all or specific projects
	ProjectScope string `json:"projectScope,omitempty" jsonschema:"title=Project Scope"`
	// ProjectIDs lists the explicitly selected GCP projects when project scope is specific
	ProjectIDs []string `json:"projectIds,omitempty" jsonschema:"title=Project IDs"`
	// SourceID is the default SCC source identifier used for this installation
	SourceID string `json:"sourceId,omitempty" jsonschema:"title=SCC Source ID"`
	// SourceIDs lists the SCC source identifiers configured for collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"title=SCC Source IDs"`
	// ServiceAccountEmail is the service account email extracted from the configured key when available
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty" jsonschema:"title=Service Account Email"`
}
