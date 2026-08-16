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
	// sccClient is the client ref for the GCP Security Command Center client used by this definition
	sccClient = types.NewClientRef[*cloudscc.Client]()
	// sccSchema is the credential schema for GCP Security Command Center service account credentials
	sccSchema, sccCredential = providerkit.CredentialSchema[CredentialSchema]()
	// workloadIdentitySchema is the credential schema for GCP workload identity federation
	workloadIdentitySchema, workloadIdentityCredential = providerkit.CredentialSchema[WorkloadIdentityCredentialSchema]()
	// healthCheckSchema is the operation schema for the GCP Security Command Center health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// findingsCollectSchema is the operation schema for the GCP Security Command Center findings collection operation
	findingsCollectSchema, findingsCollectOperation = providerkit.OperationSchema[FindingsSync]()
)

const (
	// projectScopeSpecific indicates collection should target only the explicitly listed project IDs
	projectScopeSpecific = "specific"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FindingsSync includes the configuration for the findings collection operation
	FindingsSync FindingsSyncConfig `json:"findingsSync" jsonschema:"title=Findings Sync"`
}

type FindingsSyncConfig struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.),example=Example: payload.category != \"GKE_SECURITY_BULLETIN\""`
}

// CollectionScope holds the SCC collection targeting shared by both credential schemas
type CollectionScope struct {
	// OrganizationID is the GCP organization identifier
	OrganizationID string `json:"organizationId,omitempty" jsonschema:"title=Organization ID,description=The ID of the organization to use as the parent - either organization ID or project ID are required"`
	// ProjectID is the fallback GCP project identifier used for quota and parent resolution
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID,description=The ID of the project to use as the parent - either organization ID or project ID are required"`
	// ProjectScope controls whether SCC collection targets all or specific projects
	ProjectScope string `json:"projectScope,omitempty" jsonschema:"title=Project Scope,description=Filter project scope; only used if using an Organization ID as the initial filter,enum=all,enum=specific,default=all"`
	// ProjectIDs lists the specific GCP projects used when project scope is specific
	ProjectIDs []string `json:"projectIds,omitempty" jsonschema:"title=Project IDs,description=List of project IDs to include if the project scope is set to specific"`
	// SourceIDs lists the SCC source identifiers used for collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"title=SCC Source IDs,description=Limit sources to include in pulling in findings from SCC"`
}

// CredentialSchema holds the GCP SCC service account credentials for one installation
type CredentialSchema struct {
	// ServiceAccountKey is the service account key JSON used for direct credentials
	ServiceAccountKey string `json:"serviceAccountKey" jsonschema:"required,title=Service Account Key JSON,secret=true,description=Service Account JSON used to authenticate on your behalf to GCP SCC"`
	// CollectionScope targets which GCP organization, projects, and sources to collect from
	CollectionScope
}

// WorkloadIdentityCredentialSchema holds the GCP workload identity federation inputs for
// one installation. Openlane signs a short-lived assertion with its own keys and exchanges
// it at Google STS, so nothing in this schema is secret
type WorkloadIdentityCredentialSchema struct {
	// ProjectNumber is the numeric project number of the GCP project hosting the openlane workload identity pool and provider
	ProjectNumber string `json:"projectNumber" jsonschema:"required,title=GCP Project Number,pattern=^[0-9]+$,description=Numeric project number of the GCP project hosting the openlane workload identity pool and provider"`
	// ServiceAccountEmail optionally impersonates a service account
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty" jsonschema:"title=Service Account Email,description=Optional service account to impersonate"`
	// CollectionScope targets which GCP organization, projects, and sources to collect from
	CollectionScope
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
	// SourceIDs lists the SCC source identifiers configured for collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"title=SCC Source IDs,description=Filter which sources findings are pulled from, by default all sources are included within the specified organization or project"`
	// ServiceAccountEmail is the service account email extracted from the configured key when available
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty" jsonschema:"title=Service Account Email"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.OrganizationID,
	}
}
