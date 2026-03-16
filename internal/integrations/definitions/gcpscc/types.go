package gcpscc

import (
	"encoding/json"
	"strings"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the GCP Security Command Center integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0GCPSCC00000000000000001")
	// SCCClient is the client ref for the GCP Security Command Center client used by this definition
	SCCClient = types.NewClientRef[*cloudscc.Client]()
	// HealthDefaultOperation is the operation ref for the GCP SCC health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	// FindingsCollectOperation is the operation ref for the GCP SCC findings collection operation
	FindingsCollectOperation = types.NewOperationRef[FindingsCollect]("findings.collect")
	// SettingsScanOperation is the operation ref for the GCP SCC settings scan operation
	SettingsScanOperation = types.NewOperationRef[SettingsScan]("settings.scan")
)

// Slug is the unique identifier for the GCP Security Command Center integration
const Slug = "gcp_scc"

const (
	projectScopeAll      = "all"
	projectScopeSpecific = "specific"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// OrganizationID is the GCP organization identifier
	OrganizationID string `json:"organizationId,omitempty" jsonschema:"title=Organization ID"`
	// ProjectIDs limits collection to specific GCP project identifiers
	ProjectIDs []string `json:"projectIds,omitempty" jsonschema:"title=Project IDs"`
	// ProjectScope controls whether collection covers all or specific projects
	ProjectScope string `json:"projectScope,omitempty" jsonschema:"title=Project Scope"`
	// SourceID is the SCC source identifier
	SourceID string `json:"sourceId,omitempty" jsonschema:"title=SCC Source ID"`
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
	// WorkloadIdentityProvider is the workload identity provider resource name
	WorkloadIdentityProvider string `json:"workloadIdentityProvider,omitempty" jsonschema:"title=Workload Identity Provider"`
	// Audience is the STS audience used for workload identity federation
	Audience string `json:"audience,omitempty" jsonschema:"title=Audience"`
	// ServiceAccountEmail is the service account email used for impersonation
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty" jsonschema:"title=Service Account Email"`
	// SourceID is the default SCC source identifier used when a run does not override it
	SourceID string `json:"sourceId,omitempty" jsonschema:"title=SCC Source ID"`
	// SourceIDs lists the SCC source identifiers used for collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"title=SCC Source IDs"`
	// Scopes lists the OAuth scopes requested for service account credentials
	Scopes []string `json:"scopes,omitempty" jsonschema:"title=OAuth Scopes"`
	// TokenLifetime is the requested lifetime for impersonated access tokens
	TokenLifetime string `json:"tokenLifetime,omitempty" jsonschema:"title=Token Lifetime"`
	// FindingFilter is the default SCC findings filter applied during collection
	FindingFilter string `json:"findingFilter,omitempty" jsonschema:"title=Findings Filter"`
	// ServiceAccountKey is the service account key JSON used for direct credentials
	ServiceAccountKey string `json:"serviceAccountKey,omitempty" jsonschema:"title=Service Account Key JSON"`
	// SubjectToken is the external subject token used for workload identity federation
	SubjectToken string `json:"subjectToken,omitempty" jsonschema:"title=Subject Token"`
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
