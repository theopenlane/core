package azuresecuritycenter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// complianceAssessmentsQuery selects every regulatory compliance assessment with its recommendation metadata, unhealthy resources, and latest status change,
	// preferring the status recorded for the standard's initiative and falling back to the assessment's own status
	complianceAssessmentsQuery = `securityresources
| where type =~ "microsoft.security/regulatorycompliancestandards/regulatorycompliancecontrols/regulatorycomplianceassessments"
| extend assessmentKey = name, initiative = replace("-", " ", tostring(split(id, "/")[6]))
| join kind=leftouter (
    securityresources
    | where type =~ "microsoft.security/assessments"
    | extend resourceId = tostring(split(tolower(id), "/providers/microsoft.security/assessments/")[0])
    | mv-expand s = properties.statusPerInitiative
    | summarize initiativeChanged = max(todatetime(s.assessmentStatus.statusChangeDate)), initiativeResources = make_set_if(resourceId, tostring(s.assessmentStatus.code) =~ "Unhealthy", 500) by subscriptionId, assessmentKey = name, initiative = tostring(s.policyInitiativeName)
) on subscriptionId, assessmentKey, initiative
| join kind=leftouter (
    securityresources
    | where type =~ "microsoft.security/assessments"
    | extend resourceId = tostring(split(tolower(id), "/providers/microsoft.security/assessments/")[0])
    | summarize assessmentCount = count(), statusChanged = max(todatetime(properties.status.statusChangeDate)), statusResources = make_set_if(resourceId, tostring(properties.status.code) =~ "Unhealthy", 500), metadata = take_any(properties.metadata) by subscriptionId, assessmentKey = name
) on subscriptionId, assessmentKey
| join kind=leftouter (
    securityresources
    | where type =~ "microsoft.security/assessments/governanceassignments"
    | where isnotempty(tostring(properties.owner))
    | extend assessmentKey = tostring(split(tolower(tostring(properties.assignedResourceId)), "/")[-1])
    | summarize owners = make_set(tostring(properties.owner), 50) by subscriptionId, assessmentKey
) on subscriptionId, assessmentKey
| extend sourceUpdatedAt = coalesce(initiativeChanged, statusChanged), resources = iff(isnotnull(initiativeChanged), initiativeResources, statusResources), assessmentCount = coalesce(assessmentCount, 0)
| project id, name, properties, metadata, resources, owners, sourceUpdatedAt, assessmentCount`
	// categoryRollupAssessmentType is the assessment type of compliance assessments covering a whole recommendation category
	categoryRollupAssessmentType = "AssessmentCategory"
)

// ComplianceAssessmentsCollect collects Defender for Cloud regulatory compliance assessments as checks, with failures as findings
type ComplianceAssessmentsCollect struct {
	types.OperationSettings
	// SubscriptionIDs limits collection to these subscriptions; when empty every accessible subscription is collected
	SubscriptionIDs []string `json:"subscriptionIds,omitempty" jsonschema:"title=Subscription IDs,description=Optional list of Azure subscription IDs to collect from; leave empty to collect from every subscription Openlane has been granted access to"`
	// Standards limits collection to these compliance standards; when empty every enabled standard is collected
	Standards []string `json:"standards,omitempty" jsonschema:"title=Compliance Standards,description=Optional list of compliance standard names to collect such as Azure-SOC-2023; leave empty to collect every standard enabled in Defender for Cloud"`
	// CollectChecks records every compliance assessment as a check linked to its control; when off only failed policy-only assessments are collected, as findings
	CollectChecks bool `json:"collectChecks,omitempty" jsonschema:"title=Collect Checks,description=Record every compliance assessment as a check linked to its control; when off only failed assessments without a Defender recommendation are collected as findings"`
}

// runComplianceAssessmentsCollect emits a check result per compliance assessment when checks are collected; failures backed by a Defender recommendation
// are already findings from AssessmentsCollect, so only failures with no recommendation become findings here
func runComplianceAssessmentsCollect(ctx context.Context, _ types.OperationRequest, client *Client, cfg ComplianceAssessmentsCollect) ([]types.IngestPayloadSet, error) {
	rows, err := client.queryRows[complianceAssessment](ctx, complianceAssessmentsQuery, cfg.SubscriptionIDs)
	if err != nil {
		return nil, err
	}

	observedAt := time.Now().UTC()

	var (
		checkPayloads  []CompliancePayload
		failedPolicies []CompliancePayload
	)

	for _, row := range rows {
		payload := buildCompliancePayload(row, observedAt)
		if !standardSelected(cfg.Standards, payload.Standard) {
			continue
		}

		if cfg.CollectChecks {
			checkPayloads = append(checkPayloads, payload)
		}

		// category rollups such as SoftwareUpdate have no single recommendation to report, so they stay checks only
		if row.AssessmentCount == 0 && !isCategoryRollup(row) && strings.EqualFold(payload.State, string(armsecurity.StateFailed)) {
			failedPolicies = append(failedPolicies, payload)
		}
	}

	checks, err := providerkit.MarshalEnvelopes(checkPayloads, ErrIngestPayloadEncode)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("azuresecuritycenter: failed marshal check envelopes")
		return nil, err
	}

	findings, err := providerkit.MarshalEnvelopes(complianceFindings(failedPolicies), ErrIngestPayloadEncode)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("azuresecuritycenter: failed marshal finding envelopes")
		return nil, err
	}

	sets := []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaFinding.Name,
			Envelopes: findings,
		},
	}

	if cfg.CollectChecks {
		sets = append(sets, types.IngestPayloadSet{
			Schema:    entityops.SchemaCheckResult.Name,
			Envelopes: checks,
		})
	}

	return sets, nil
}

// complianceFindings merges failed compliance-only assessments into one finding per subscription and assessment key, with every referencing control in categories
func complianceFindings(failed []CompliancePayload) []AssessmentPayload {
	var findings []AssessmentPayload

	index := map[string]int{}

	for _, p := range failed {
		id := fmt.Sprintf(recommendationFindingIDFormat, p.SubscriptionID, p.AssessmentKey)

		i, seen := index[id]
		if !seen {
			index[id] = len(findings)
			findings = append(findings, complianceFindingPayload(id, p))

			continue
		}

		if p.Control != "" && !lo.Contains(findings[i].Categories, p.Control) {
			findings[i].Categories = append(findings[i].Categories, p.Control)
		}

		if p.Standard != "" && !lo.Contains(findings[i].Standards, p.Standard) {
			findings[i].Standards = append(findings[i].Standards, p.Standard)
		}
	}

	return findings
}

// buildCompliancePayload extracts the standard and control from the assessment ID and the mappable fields from its properties
func buildCompliancePayload(row complianceAssessment, observedAt time.Time) CompliancePayload {
	props := row.Properties

	payload := CompliancePayload{
		ID:               row.ID,
		AssessmentKey:    row.Name,
		DisplayName:      lo.CoalesceOrEmpty(props.Description, row.Metadata.DisplayName),
		Description:      row.Metadata.Description,
		Remediation:      row.Metadata.RemediationDescription,
		Severity:         row.Metadata.Severity,
		Resources:        row.Resources,
		Owners:           row.Owners,
		State:            props.State,
		Status:           complianceCheckStatus(props.State).String(),
		AssessmentType:   props.AssessmentType,
		ExternalURI:      lo.Ternary(isCategoryRollup(row), "", portalRecommendationURL(row.Name)),
		DetailsLink:      props.AssessmentDetailsLink,
		PassedResources:  props.PassedResources,
		FailedResources:  props.FailedResources,
		SkippedResources: props.SkippedResources,
		SourceUpdatedAt:  row.SourceUpdatedAt,
		ObservedAt:       observedAt,
	}

	if parsed, err := arm.ParseResourceID(row.ID); err == nil && parsed.Parent != nil {
		payload.SubscriptionID = parsed.SubscriptionID
		payload.Control = parsed.Parent.Name

		if parsed.Parent.Parent != nil {
			payload.Standard = parsed.Parent.Parent.Name
		}
	}

	return payload
}

// complianceFindingPayload converts a failed compliance assessment to the finding payload shared with posture assessments;
// the standard and control go in category and categories so the finding links to its control
func complianceFindingPayload(id string, p CompliancePayload) AssessmentPayload {
	payload := AssessmentPayload{
		ID:          id,
		DisplayName: p.DisplayName,
		Description: p.Description,
		Remediation: p.Remediation,
		Severity:    p.Severity,
		StatusCode:  string(armsecurity.AssessmentStatusCodeUnhealthy),
		ExternalURI: p.ExternalURI,
		UpdatedAt:   p.SourceUpdatedAt,
		Resources:   p.Resources,
		Metadata: map[string]any{
			"assessment_key":    p.AssessmentKey,
			"state":             p.State,
			"passed_resources":  p.PassedResources,
			"failed_resources":  p.FailedResources,
			"skipped_resources": p.SkippedResources,
			"owners":            p.Owners,
		},
	}

	if p.Control != "" {
		payload.Categories = []string{p.Control}
	}

	if p.Standard != "" {
		payload.Standards = []string{p.Standard}
	}

	if len(p.Owners) > 0 {
		payload.GovernanceOwner = p.Owners[0]
	}

	return payload
}

// isCategoryRollup reports whether a compliance assessment covers a whole recommendation category, such as SoftwareUpdate, rather than one recommendation
func isCategoryRollup(row complianceAssessment) bool {
	return strings.EqualFold(row.Properties.AssessmentType, categoryRollupAssessmentType)
}

// complianceCheckStatus maps a compliance assessment state to a check status
func complianceCheckStatus(state string) enums.CheckStatus {
	switch {
	case strings.EqualFold(state, string(armsecurity.StatePassed)):
		return enums.CheckStatusPass
	case strings.EqualFold(state, string(armsecurity.StateFailed)):
		return enums.CheckStatusFail
	default:
		return enums.CheckStatusUnknown
	}
}

// standardSelected reports whether a standard is in the configured list, every standard being selected when the list is empty
func standardSelected(standards []string, standard string) bool {
	if len(standards) == 0 {
		return true
	}

	for _, s := range standards {
		if strings.EqualFold(s, standard) {
			return true
		}
	}

	return false
}
