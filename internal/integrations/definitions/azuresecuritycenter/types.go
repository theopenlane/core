package azuresecuritycenter

import (
	"encoding/json"
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

var (
	// definitionID is the stable identifier for the Azure Security Center integration definition
	definitionID = types.NewDefinitionRef("def_01K0AZSECC000000000000000001")
	// installation is the typed installation metadata handle for the definition
	installation = types.InstallationOf[InstallationMetadata]()
	// securityCenterConnection is the typed connection handle for the consented tenant credential
	securityCenterConnection = types.ConnectionOf[securityCenterCred]()
	// lighthouseConnection is the typed connection handle for subscriptions delegated through Azure Lighthouse
	lighthouseConnection = types.ConnectionOf[lighthouseCred]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// SubscriptionIDs limits every operation to these subscriptions; when empty every accessible subscription is collected
	SubscriptionIDs []string `json:"subscriptionIds,omitempty" jsonschema:"title=Subscription IDs,description=Optional list of Azure subscription IDs to collect from; leave empty to collect from every subscription Openlane has been granted access to"`
}

// requestedSubscriptions returns the subscription IDs the installation's user input limits collection to, empty for every accessible subscription
func requestedSubscriptions(req types.OperationRequest) ([]string, error) {
	var input UserInput
	if err := jsonx.UnmarshalIfPresent(req.Integration.UserInput.Data, &input); err != nil {
		return nil, err
	}

	return input.SubscriptionIDs, nil
}

// securityCenterCred holds the per-installation credential for one consented Azure tenant
type securityCenterCred struct {
	// TenantID is the Azure Active Directory tenant identifier for this installation
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID"`
}

// newSecurityCenterCred builds the Security Center credential for a consented tenant
func newSecurityCenterCred(tenantID, _ string) securityCenterCred {
	return securityCenterCred{TenantID: tenantID}
}

// lighthouseCred holds the per-installation credential for subscriptions delegated through Azure Lighthouse
type lighthouseCred struct {
	// TenantID is the customer tenant that owns the delegated subscriptions
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID,description=Your Microsoft Entra tenant ID."`
	// ConnectionCode ties the Lighthouse delegation to this installation through the offer name
	ConnectionCode string `json:"connectionCode" jsonschema:"required,title=Connection Code,description=Enter this code as the connectionCode parameter when deploying the Lighthouse template." jsonschema_extras:"generate=true"`
}

// InstallationMetadata holds the stable Azure tenant identity for one installation
type InstallationMetadata struct {
	// TenantID is the Azure Active Directory tenant identifier
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.TenantID,
	}
}

// ResourceDetails describes the Azure or on-premise resource an assessment was run against
type ResourceDetails struct {
	// ID is the ARM resource ID, or the source computer ID for on-premise machines
	ID string `json:"id,omitempty"`
	// Source is Azure, OnPremise, or OnPremiseSql
	Source string `json:"source,omitempty"`
	// Name is the resource name
	Name string `json:"name,omitempty"`
	// Type is the ARM resource type such as Microsoft.Compute/virtualMachines
	Type string `json:"type,omitempty"`
	// SubscriptionID is the subscription the resource belongs to
	SubscriptionID string `json:"subscription_id,omitempty"`
	// ResourceGroup is the resource group the resource belongs to
	ResourceGroup string `json:"resource_group,omitempty"`
	// Provider is the scanning provider of the resource, e.g. acr for container registry images
	Provider string `json:"provider,omitempty"`
	// ParentResourceID is the ARM ID of the resource holding a scanned entity, e.g. the registry of a container image
	ParentResourceID string `json:"parent_resource_id,omitempty"`
	// ParentResourceName is the name of the parent resource, e.g. the registry name
	ParentResourceName string `json:"parent_resource_name,omitempty"`
	// Repository is the container image repository
	Repository string `json:"repository,omitempty"`
	// Digest is the container image digest
	Digest string `json:"digest,omitempty"`
}

// assessment is one Defender for Cloud assessment row returned by Resource Graph
type assessment struct {
	// ID is the full ARM resource ID of the assessment
	ID string `json:"id"`
	// Name is the assessment key
	Name string `json:"name"`
	// Properties holds the assessment status, resource, and metadata
	Properties assessmentProperties `json:"properties"`
	// GovernanceOwner is the owner email from the assessment's governance assignment, empty when unassigned
	GovernanceOwner string `json:"governanceOwner"`
	// ComplianceControls are the regulatory compliance control IDs that reference this assessment
	ComplianceControls []string `json:"complianceControls"`
	// ComplianceStandards are the regulatory compliance standards that reference this assessment
	ComplianceStandards []string `json:"complianceStandards"`
}

// assessmentProperties holds the fields read from an assessment's properties
type assessmentProperties struct {
	// DisplayName is the assessment display name
	DisplayName string `json:"displayName"`
	// Status is the assessment evaluation status
	Status assessmentStatus `json:"status"`
	// ResourceDetails describes the assessed resource; keys arrive in mixed case and decode case-insensitively
	ResourceDetails assessmentResourceDetails `json:"resourceDetails"`
	// Links holds the portal link for the assessment
	Links assessmentLinks `json:"links"`
	// Metadata is the raw assessment metadata
	Metadata json.RawMessage `json:"metadata"`
	// AdditionalData holds source-specific data such as CVE details on vulnerability recommendations
	AdditionalData map[string]any `json:"additionalData"`
}

// assessmentStatus is the evaluation status of an assessment
type assessmentStatus struct {
	// Code is Healthy, Unhealthy, or NotApplicable
	Code string `json:"code"`
	// Cause is the programmatic reason for the status
	Cause string `json:"cause"`
	// FirstEvaluationDate is when the assessment was first evaluated
	FirstEvaluationDate *time.Time `json:"firstEvaluationDate"`
	// StatusChangeDate is when the status last changed
	StatusChangeDate *time.Time `json:"statusChangeDate"`
}

// assessmentResourceDetails describes the resource an assessment ran against
type assessmentResourceDetails struct {
	// ID is the ARM resource ID for Azure resources
	ID string `json:"id"`
	// Source is Azure, OnPremise, or OnPremiseSql
	Source string `json:"source"`
	// MachineName is the on-premise machine name
	MachineName string `json:"machineName"`
	// SourceComputerID is the on-premise machine identifier
	SourceComputerID string `json:"sourceComputerId"`
	// ResourceName is the resource name, e.g. the image repository for container images
	ResourceName string `json:"resourceName"`
	// ResourceType is the resource type, e.g. .containerimage
	ResourceType string `json:"resourceType"`
	// ResourceProvider is the scanning provider, e.g. acr
	ResourceProvider string `json:"resourceProvider"`
}

// assessmentLinks holds the links attached to an assessment
type assessmentLinks struct {
	// AzurePortalURI is the portal link for the assessment
	AzurePortalURI string `json:"azurePortalUri"`
}

// assessmentMetadata holds the fields read from an assessment's metadata
type assessmentMetadata struct {
	// DisplayName is the recommendation display name
	DisplayName string `json:"displayName"`
	// Severity is High, Medium, or Low
	Severity string `json:"severity"`
	// Description is the recommendation description
	Description string `json:"description"`
	// RemediationDescription is the remediation guidance
	RemediationDescription string `json:"remediationDescription"`
	// AssessmentType is BuiltIn, CustomPolicy, CustomerManaged, or VerifiedPartner
	AssessmentType string `json:"assessmentType"`
	// RecommendationCategory groups individual recommendations, e.g. SoftwareUpdate
	RecommendationCategory string `json:"recommendationCategory"`
	// Categories are the security categories of the recommendation
	Categories []string `json:"categories"`
	// Threats are the threat categories of the recommendation
	Threats []string `json:"threats"`
}

// complianceAssessment is one regulatory compliance assessment row returned by Resource Graph
type complianceAssessment struct {
	// ID is the full ARM resource ID, nested under its standard and control
	ID string `json:"id"`
	// Name is the assessment key, shared with the matching Defender recommendation
	Name string `json:"name"`
	// Properties holds the assessment state and resource counts
	Properties complianceAssessmentProperties `json:"properties"`
	// SourceUpdatedAt is the latest status change of the underlying assessment for this standard, nil when none is recorded
	SourceUpdatedAt *time.Time `json:"sourceUpdatedAt"`
	// Metadata is the recommendation metadata of the underlying assessment
	Metadata assessmentMetadata `json:"metadata"`
	// Resources are the IDs of the resources whose underlying assessment is unhealthy for this standard
	Resources []string `json:"resources"`
	// Owners are the governance assignment owners of the underlying assessment
	Owners []string `json:"owners"`
	// AssessmentCount is the number of Defender recommendation rows backing this compliance assessment, zero for policy-only checks
	AssessmentCount int `json:"assessmentCount"`
}

// complianceAssessmentProperties holds the fields read from a regulatory compliance assessment
type complianceAssessmentProperties struct {
	// Description is the assessment display name
	Description string `json:"description"`
	// State is Passed, Failed, Skipped, or Unsupported
	State string `json:"state"`
	// AssessmentType is the assessment type
	AssessmentType string `json:"assessmentType"`
	// AssessmentDetailsLink links to the assessment details
	AssessmentDetailsLink string `json:"assessmentDetailsLink"`
	// PassedResources is the number of resources that passed
	PassedResources int `json:"passedResources"`
	// FailedResources is the number of resources that failed
	FailedResources int `json:"failedResources"`
	// SkippedResources is the number of resources that were skipped
	SkippedResources int `json:"skippedResources"`
}

// CompliancePayload is the raw provider payload emitted per regulatory compliance assessment for ingest
type CompliancePayload struct {
	// ID is the full ARM resource ID of the compliance assessment — used as the check parent_external_id
	ID string `json:"id"`
	// SubscriptionID is the subscription the compliance assessment belongs to
	SubscriptionID string `json:"subscription_id,omitempty"`
	// Standard is the compliance standard name, e.g. Azure-SOC-2023
	Standard string `json:"standard"`
	// Control is the control ID within the standard, e.g. CC7.4
	Control string `json:"control"`
	// AssessmentKey is the assessment key, shared with the matching Defender recommendation
	AssessmentKey string `json:"assessment_key"`
	// DisplayName is the assessment display name
	DisplayName string `json:"display_name"`
	// Description is the recommendation description
	Description string `json:"description,omitempty"`
	// Remediation is the recommendation remediation guidance
	Remediation string `json:"remediation,omitempty"`
	// Severity is High, Medium, or Low
	Severity string `json:"severity,omitempty"`
	// Resources are the IDs of the unhealthy resources for this standard
	Resources []string `json:"resources,omitempty"`
	// Owners are the governance assignment owners of the underlying assessment — the first is used as internal_owner
	Owners []string `json:"owners,omitempty"`
	// State is Passed, Failed, Skipped, or Unsupported
	State string `json:"state"`
	// Status is the check status, PASS, FAIL, or UNKNOWN
	Status string `json:"status"`
	// AssessmentType is the assessment type
	AssessmentType string `json:"assessment_type,omitempty"`
	// ExternalURI is the Defender for Cloud portal link for the assessment
	ExternalURI string `json:"external_uri,omitempty"`
	// DetailsLink is the assessment details link from the source
	DetailsLink string `json:"details_link,omitempty"`
	// PassedResources is the number of resources that passed
	PassedResources int `json:"passed_resources"`
	// FailedResources is the number of resources that failed
	FailedResources int `json:"failed_resources"`
	// SkippedResources is the number of resources that were skipped
	SkippedResources int `json:"skipped_resources"`
	// SourceUpdatedAt is the latest status change of the underlying assessment for this standard
	SourceUpdatedAt *time.Time `json:"source_updated_at,omitempty"`
	// ObservedAt is when the result was collected
	ObservedAt time.Time `json:"observed_at"`
}

// complianceFindingMetadata is the finding metadata recorded for a failed compliance assessment
type complianceFindingMetadata struct {
	// AssessmentKey is the assessment key shared with the matching Defender recommendation
	AssessmentKey string `json:"assessment_key"`
	// State is Passed, Failed, Skipped, or Unsupported
	State string `json:"state"`
	// PassedResources is the number of resources that passed
	PassedResources int `json:"passed_resources"`
	// FailedResources is the number of resources that failed
	FailedResources int `json:"failed_resources"`
	// SkippedResources is the number of resources that were skipped
	SkippedResources int `json:"skipped_resources"`
	// Owners are the governance assignment owners of the underlying assessment
	Owners []string `json:"owners"`
}

// cveDetail is one entry of a vulnerability recommendation's CvesDetails list
type cveDetail struct {
	// CveID is the CVE identifier
	CveID string `json:"CveId"`
	// Severity is the CVE severity, e.g. Medium
	Severity string `json:"Severity,omitempty"`
	// FixStatus reports whether a fix exists, e.g. FixAvailable
	FixStatus string `json:"FixStatus,omitempty"`
	// FixedVersion is the package version that fixes this CVE
	FixedVersion string `json:"FixedVersion,omitempty"`
}

// AssessmentPayload is the raw provider payload emitted per unhealthy assessment for ingest
type AssessmentPayload struct {
	// ID is the full ARM resource ID of the assessment — used as external_id
	ID string `json:"id"`
	// ResourceID is the ARM resource ID of the assessed Azure resource, empty once a finding spans several resources
	ResourceID string `json:"resource_id,omitempty"`
	// SubscriptionID is the subscription the assessment belongs to
	SubscriptionID string `json:"subscription_id,omitempty"`
	// AssessmentKey is the recommendation key, shared by every resource the recommendation assesses
	AssessmentKey string `json:"assessment_key,omitempty"`
	// GovernanceOwner is the owner assigned through a Defender governance rule — used as external_owner_id
	GovernanceOwner string `json:"governance_owner,omitempty"`
	// ResourceDetails describes the assessed resource, nil for compliance assessments that span several resources
	ResourceDetails *ResourceDetails `json:"resource_details,omitempty"`
	// Resources are the IDs of the affected resources — used as targets
	Resources []string `json:"resources,omitempty"`
	// AffectedResources describes every resource a recommendation finding covers — used as target details
	AffectedResources []ResourceDetails `json:"affected_resources,omitempty"`
	// Standards are the compliance standards that reference the finding, e.g. Azure-SOC-2023 — used as tags
	Standards []string `json:"standards,omitempty"`
	// Metadata is the assessment metadata from the source
	Metadata any `json:"metadata,omitempty"`
	// AdditionalData is the assessment additional data from the source
	AdditionalData map[string]any `json:"additional_data,omitempty"`
	// DisplayName is the human-readable assessment policy or finding name
	DisplayName string `json:"display_name"`
	// Description is the human-readable description of the security check or finding
	Description string `json:"description,omitempty"`
	// Category is the primary category (first of Categories for assessments)
	Category string `json:"category,omitempty"`
	// StatusCode is Healthy, Unhealthy, or NotApplicable
	StatusCode string `json:"status_code"`
	// Severity is High, Medium, or Low
	Severity string `json:"severity,omitempty"`
	// Remediation is the human-readable remediation guidance
	Remediation string `json:"remediation,omitempty"`
	// ExternalURI is the Azure Portal link for this assessment result
	ExternalURI string `json:"external_uri,omitempty"`
	// DiscoveredAt is when the assessment was first evaluated
	DiscoveredAt *time.Time `json:"discovered_at,omitempty"`
	// UpdatedAt is when the assessment status last changed
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// StatusCause is the programmatic reason for the status when set
	StatusCause string `json:"status_cause,omitempty"`
	// Categories is the full set of categories from the assessment metadata
	Categories []string `json:"categories,omitempty"`
	// AssessmentType is BuiltIn, CustomPolicy, CustomerManaged, or VerifiedPartner
	AssessmentType string `json:"assessment_type,omitempty"`
	// Threats is the set of threat categories associated with this assessment
	Threats []string `json:"threats,omitempty"`
	// RecommendationCategory groups individual recommendations, e.g. SoftwareUpdate
	RecommendationCategory string `json:"recommendation_category,omitempty"`
	// CVEDetails holds the CVEs of a vulnerability recommendation
	CVEDetails []cveDetail `json:"cve_details,omitempty"`
	// PackageName is the vulnerable package name
	PackageName string `json:"package_name,omitempty"`
	// PackageEcosystem is the package vendor or distribution, e.g. debian
	PackageEcosystem string `json:"package_ecosystem,omitempty"`
	// DetectedVersions are the installed versions of the vulnerable package
	DetectedVersions []string `json:"detected_versions,omitempty"`
	// CVSSScore is the highest CVSS score across the recommendation's CVEs
	CVSSScore *float64 `json:"cvss_score,omitempty"`
	// CveID is the single CVE a per-CVE vulnerability record describes
	CveID string `json:"cve_id,omitempty"`
	// FixedVersion is the software version that resolves a vulnerability recommendation
	FixedVersion string `json:"fixed_version,omitempty"`
}
