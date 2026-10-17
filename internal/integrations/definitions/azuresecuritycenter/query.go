package azuresecuritycenter

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// vulnerabilityCategories are the recommendation categories that carry CVE findings
	vulnerabilityCategories = `("SoftwareUpdate", "ServiceUpgrade")`
	// assessmentsQuery selects every Defender for Cloud assessment
	assessmentsQuery = `securityresources | where type =~ "microsoft.security/assessments"`
	// assessmentsProbeQuery reads a single assessment to confirm Resource Graph access
	assessmentsProbeQuery = assessmentsQuery + ` | take 1`
	// unhealthyAssessmentsQuery selects every unhealthy Defender for Cloud assessment
	unhealthyAssessmentsQuery = assessmentsQuery + ` | where properties.status.code =~ "Unhealthy"`
	// governanceOwnerJoin attaches the owner of each assessment's governance assignment, when one exists
	governanceOwnerJoin = ` | extend assessmentId = tolower(id) | join kind=leftouter (securityresources | where type =~ "microsoft.security/assessments/governanceassignments" | where isnotempty(tostring(properties.owner)) | extend assessmentId = tolower(tostring(properties.assignedResourceId)) | summarize governanceOwner = take_any(tostring(properties.owner)) by assessmentId) on assessmentId`
	// complianceControlsJoin attaches the regulatory compliance control IDs that reference each assessment
	complianceControlsJoin = ` | extend assessmentKey = name | join kind=leftouter (securityresources | where type =~ "microsoft.security/regulatorycompliancestandards/regulatorycompliancecontrols/regulatorycomplianceassessments" | summarize complianceControls = make_set(tostring(split(id, "/")[8])), complianceStandards = make_set(tostring(split(id, "/")[6])) by subscriptionId, assessmentKey = name) on subscriptionId, assessmentKey`
	// findingsQuery selects unhealthy posture assessments outside the vulnerability categories, with their owner and compliance controls
	findingsQuery = unhealthyAssessmentsQuery + ` | where tostring(properties.metadata.recommendationCategory) !in~ ` + vulnerabilityCategories + governanceOwnerJoin + complianceControlsJoin + ` | project id, name, properties, governanceOwner, complianceControls, complianceStandards`
	// vulnerabilitiesQuery selects unhealthy individual recommendations in the vulnerability categories, with their owner
	vulnerabilitiesQuery = unhealthyAssessmentsQuery + ` | where tostring(properties.metadata.recommendationCategory) in~ ` + vulnerabilityCategories + governanceOwnerJoin + ` | project id, name, properties, governanceOwner`
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
)

// subscriptions returns the requested subscription IDs, or every accessible subscription when none are requested
func (c *Client) subscriptions(ctx context.Context, requested []string) ([]string, error) {
	if len(requested) > 0 {
		return requested, nil
	}

	ids, err := c.arm.ListSubscriptionIDs(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Msg("azuresecuritycenter: error getting subscriptions")
		return nil, ErrNoSubscriptions
	}

	if len(ids) == 0 {
		logx.FromContext(ctx).Error().Msg("azuresecuritycenter: no subscriptions returned")
		return nil, ErrNoSubscriptions
	}

	return ids, nil
}

// queryRows runs a Resource Graph query over the requested subscriptions, or every accessible one, and decodes each row into T
func (c *Client) queryRows[T any](ctx context.Context, query string, requested []string) ([]T, error) {
	subscriptionIDs, err := c.subscriptions(ctx, requested)
	if err != nil {
		return nil, err
	}

	rows, err := c.arm.QueryResourceGraph(ctx, query, subscriptionIDs)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("azuresecuritycenter: resource graph query failed")
		return nil, ErrResourceGraphQueryFailed
	}

	decoded := make([]T, 0, len(rows))

	for _, row := range rows {
		var item T
		if err := json.Unmarshal(row, &item); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("azuresecuritycenter: failed to decode resource graph row")
			return nil, ErrAssessmentDecode
		}

		decoded = append(decoded, item)
	}

	return decoded, nil
}
