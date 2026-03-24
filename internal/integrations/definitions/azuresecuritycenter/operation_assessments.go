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
	// variantAssessment identifies top-level Defender for Cloud assessment envelopes
	variantAssessment = "assessment"
)

// AssessmentPayload is the raw provider payload emitted per unhealthy assessment for ingest.
// Assessments are security posture policy checks (e.g. "MFA should be enabled",
// "Storage accounts should disable public access"). They are misconfigurations, not CVE
// vulnerabilities. This data maps to the Vulnerability schema until the Finding schema
// gains integration mapping pipeline support, at which point assessments should be remapped.
type AssessmentPayload struct {
	// ID is the full ARM resource ID of the assessment — used as external_id
	ID string `json:"id"`
	// ResourceID is the ARM resource ID of the assessed Azure resource — used as external_owner_id
	ResourceID string `json:"resource_id,omitempty"`
	// DisplayName is the human-readable assessment policy name
	DisplayName string `json:"display_name"`
	// StatusCode is Healthy, Unhealthy, or NotApplicable
	StatusCode string `json:"status_code"`
	// StatusCause is the programmatic reason for the status when set
	StatusCause string `json:"status_cause,omitempty"`
	// Severity is High, Medium, or Low from the assessment metadata
	Severity string `json:"severity,omitempty"`
	// Category is the primary category (first of Categories)
	Category string `json:"category,omitempty"`
	// Categories is the full set of categories from the assessment metadata
	Categories []string `json:"categories,omitempty"`
	// Description is the human-readable description of the security check
	Description string `json:"description,omitempty"`
	// Remediation is the human-readable remediation guidance
	Remediation string `json:"remediation,omitempty"`
	// ExternalURI is the Azure Portal link for this assessment result
	ExternalURI string `json:"external_uri,omitempty"`
	// AssessmentType is BuiltIn, CustomPolicy, CustomerManaged, or VerifiedPartner
	AssessmentType string `json:"assessment_type,omitempty"`
	// Threats is the set of threat categories associated with this assessment
	Threats []string `json:"threats,omitempty"`
	// FirstEvaluatedAt is when the assessment was first evaluated
	FirstEvaluatedAt *time.Time `json:"first_evaluated_at,omitempty"`
	// StatusChangedAt is when the assessment status last changed
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
}

// AssessmentsCollect collects Azure Defender for Cloud assessment findings for ingest
type AssessmentsCollect struct{}

// IngestHandle adapts assessments collection to the ingest operation registration boundary
func (a AssessmentsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(securityCenterClient, func(ctx context.Context, _ types.OperationRequest, client *azureSecurityClient) ([]types.IngestPayloadSet, error) {
		return a.Run(ctx, client)
	})
}

// Run collects all unhealthy security assessment findings for the subscription
func (AssessmentsCollect) Run(ctx context.Context, client *azureSecurityClient) ([]types.IngestPayloadSet, error) {
	pager := client.assessments.NewListPager(client.scope(), nil)

	var envelopes []types.MappingEnvelope

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, ErrAssessmentFetchFailed
		}

		for _, a := range page.Value {
			if a.Properties == nil || a.Properties.Status == nil {
				continue
			}

			if string(lo.FromPtr(a.Properties.Status.Code)) != "Unhealthy" {
				continue
			}

			payload := buildAssessmentPayload(a)

			envelope, err := providerkit.MarshalEnvelopeVariant(variantAssessment, payload.ID, payload, ErrIngestPayloadEncode)
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

// buildAssessmentPayload extracts all mappable fields from an armsecurity.AssessmentResponse
func buildAssessmentPayload(a *armsecurity.AssessmentResponse) AssessmentPayload {
	payload := AssessmentPayload{
		ID:               lo.FromPtr(a.ID),
		StatusCode:       string(lo.FromPtr(a.Properties.Status.Code)),
		StatusCause:      lo.FromPtr(a.Properties.Status.Cause),
		DisplayName:      lo.FromPtr(a.Properties.DisplayName),
		FirstEvaluatedAt: a.Properties.Status.FirstEvaluationDate,
		StatusChangedAt:  a.Properties.Status.StatusChangeDate,
	}

	if a.Properties.Links != nil {
		payload.ExternalURI = lo.FromPtr(a.Properties.Links.AzurePortalURI)
	}

	if meta := a.Properties.Metadata; meta != nil {
		payload.Severity = string(lo.FromPtr(meta.Severity))
		payload.Description = lo.FromPtr(meta.Description)
		payload.Remediation = lo.FromPtr(meta.RemediationDescription)
		payload.AssessmentType = string(lo.FromPtr(meta.AssessmentType))

		for _, c := range meta.Categories {
			if c != nil {
				payload.Categories = append(payload.Categories, string(*c))
			}
		}

		if len(payload.Categories) > 0 {
			payload.Category = payload.Categories[0]
		}

		for _, t := range meta.Threats {
			if t != nil {
				payload.Threats = append(payload.Threats, string(*t))
			}
		}
	}

	switch rd := a.Properties.ResourceDetails.(type) {
	case *armsecurity.AzureResourceDetails:
		payload.ResourceID = lo.FromPtr(rd.ID)
	case *armsecurity.OnPremiseResourceDetails:
		payload.ResourceID = lo.FromPtr(rd.SourceComputerID)
	}

	return payload
}
