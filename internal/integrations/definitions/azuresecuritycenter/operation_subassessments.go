package azuresecuritycenter

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// variantSubAssessment identifies granular sub-assessment vulnerability envelopes
	variantSubAssessment = "subassessment"
)

// SubAssessmentPayload is the raw provider payload emitted per sub-assessment for ingest.
// Sub-assessments are granular findings under a top-level assessment. Container registry
// and server sub-assessments carry actual CVE identifiers and CVSS scores; SQL sub-assessments
// carry configuration check results. All types map to the Vulnerability schema.
//
// Note: CVE IDs are intentionally stored in raw_payload rather than mapped to the cve_id
// field, because the Vulnerability schema enforces a (cve_id, owner_id) unique constraint
// that assumes one record per CVE per organization. Azure sub-assessments are scoped per
// resource, so the same CVE can appear across multiple container images or servers.
type SubAssessmentPayload struct {
	// ID is the full ARM resource ID of the sub-assessment — used as external_id
	ID string `json:"id"`
	// VulnID is the vulnerability identifier from the sub-assessment (Properties.ID)
	VulnID string `json:"vuln_id,omitempty"`
	// ResourceID is the ARM resource ID of the affected Azure resource — used as external_owner_id
	ResourceID string `json:"resource_id,omitempty"`
	// DisplayName is the human-readable name of the sub-assessment finding
	DisplayName string `json:"display_name"`
	// Description is the human-readable description of the finding
	Description string `json:"description,omitempty"`
	// Category is the finding category
	Category string `json:"category,omitempty"`
	// StatusCode is Healthy, Unhealthy, or NotApplicable
	StatusCode string `json:"status_code"`
	// Severity is High, Medium, or Low from the sub-assessment status
	Severity string `json:"severity,omitempty"`
	// Impact is the human-readable impact description
	Impact string `json:"impact,omitempty"`
	// Remediation is the human-readable remediation guidance
	Remediation string `json:"remediation,omitempty"`
	// TimeGenerated is when the sub-assessment was generated
	TimeGenerated *time.Time `json:"time_generated,omitempty"`
	// ResourceType is the assessed resource type (e.g. ContainerRegistryVulnerability, ServerVulnerability)
	ResourceType string `json:"resource_type,omitempty"`
	// Patchable indicates whether a patch is available (container/server types)
	Patchable *bool `json:"patchable,omitempty"`
	// PublishedAt is when the vulnerability was published (container/server types)
	PublishedAt *time.Time `json:"published_at,omitempty"`
	// CVEs holds the CVE identifiers from the vulnerability data (container/server types)
	CVEs []string `json:"cves,omitempty"`
	// CVSSScore is the highest CVSS base score available across versions
	CVSSScore *float32 `json:"cvss_score,omitempty"`
}

// SubAssessmentsCollect collects Azure Defender for Cloud sub-assessment findings for ingest
type SubAssessmentsCollect struct{}

// IngestHandle adapts sub-assessments collection to the ingest operation registration boundary
func (s SubAssessmentsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(
		SecurityCenterClient,
		func(ctx context.Context, _ types.OperationRequest, client *azureSecurityClient) ([]types.IngestPayloadSet, error) {
			return s.Run(ctx, client)
		},
	)
}

// Run collects all unhealthy sub-assessment findings for the subscription
func (SubAssessmentsCollect) Run(ctx context.Context, client *azureSecurityClient) ([]types.IngestPayloadSet, error) {
	pager := client.subassessments.NewListAllPager(client.scope(), nil)

	var envelopes []types.MappingEnvelope

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, ErrSubAssessmentFetchFailed
		}

		for _, sa := range page.Value {
			if sa.Properties == nil || sa.Properties.Status == nil {
				continue
			}

			if string(lo.FromPtr(sa.Properties.Status.Code)) != "Unhealthy" {
				continue
			}

			payload := buildSubAssessmentPayload(sa)

			envelope, err := providerkit.MarshalEnvelopeVariant(variantSubAssessment, payload.ID, payload, ErrIngestPayloadEncode)
			if err != nil {
				return nil, ErrIngestPayloadEncode
			}

			envelopes = append(envelopes, envelope)
		}
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: envelopes,
		},
	}, nil
}

// buildSubAssessmentPayload extracts all mappable fields from an armsecurity.SubAssessment
func buildSubAssessmentPayload(sa *armsecurity.SubAssessment) SubAssessmentPayload {
	payload := SubAssessmentPayload{
		ID:          lo.FromPtr(sa.ID),
		VulnID:      lo.FromPtr(sa.Properties.ID),
		DisplayName: lo.FromPtr(sa.Properties.DisplayName),
		Description: lo.FromPtr(sa.Properties.Description),
		Category:    lo.FromPtr(sa.Properties.Category),
		Impact:      lo.FromPtr(sa.Properties.Impact),
		Remediation: lo.FromPtr(sa.Properties.Remediation),
		StatusCode:  string(lo.FromPtr(sa.Properties.Status.Code)),
		Severity:    string(lo.FromPtr(sa.Properties.Status.Severity)),
	}

	if sa.Properties.TimeGenerated != nil {
		payload.TimeGenerated = sa.Properties.TimeGenerated
	}

	switch rd := sa.Properties.ResourceDetails.(type) {
	case *armsecurity.AzureResourceDetails:
		payload.ResourceID = lo.FromPtr(rd.ID)
	case *armsecurity.OnPremiseResourceDetails:
		payload.ResourceID = lo.FromPtr(rd.SourceComputerID)
	}

	switch ad := sa.Properties.AdditionalData.(type) {
	case *armsecurity.ContainerRegistryVulnerabilityProperties:
		payload.ResourceType = "ContainerRegistryVulnerability"
		payload.Patchable = ad.Patchable
		payload.PublishedAt = ad.PublishedTime
		payload.CVEs = extractCVEIDs(ad.Cve)
		payload.CVSSScore = highestCVSSScore(ad.Cvss)

	case *armsecurity.ServerVulnerabilityProperties:
		payload.ResourceType = "ServerVulnerability"
		payload.Patchable = ad.Patchable
		payload.PublishedAt = ad.PublishedTime
		payload.CVEs = extractCVEIDs(ad.Cve)
		payload.CVSSScore = highestCVSSScore(ad.Cvss)

	case *armsecurity.SQLServerVulnerabilityProperties:
		payload.ResourceType = "SQLServerVulnerability"
	}

	return payload
}

// extractCVEIDs collects CVE identifiers from a CVE list.
// The armsecurity.CVE struct exposes the identifier via the Title field
// (e.g. "CVE-2021-44228") rather than a dedicated ID field.
func extractCVEIDs(cves []*armsecurity.CVE) []string {
	ids := make([]string, 0, len(cves))

	for _, c := range cves {
		if c != nil && c.Title != nil && *c.Title != "" {
			ids = append(ids, *c.Title)
		}
	}

	return ids
}

// highestCVSSScore returns the highest base score across all CVSS versions
func highestCVSSScore(cvss map[string]*armsecurity.CVSS) *float32 {
	var best float32

	for _, v := range cvss {
		if v != nil && v.Base != nil && *v.Base > best {
			best = *v.Base
		}
	}

	if best == 0 {
		return nil
	}

	return lo.ToPtr(best)
}
