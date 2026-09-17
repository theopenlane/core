package oci

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprFinding is the CEL mapping expression for OCI Cloud Guard problem payloads mapped to Finding; unset SDK pointer fields marshal as null, hence the != null guards
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`'id' in payload && payload.id != null ? payload.id : ""`),
	entityops.FindingFields.ExternalOwnerID.Expr(`resource`),

	entityops.FindingFields.Category.Expr(`'detectorRuleId' in payload && payload.detectorRuleId != null ? payload.detectorRuleId : ""`),
	entityops.FindingFields.FindingClass.Expr(`'detectorId' in payload ? payload.detectorId : ""`),
	entityops.FindingFields.DisplayName.Expr(`'detectorRuleId' in payload && payload.detectorRuleId != null && payload.detectorRuleId != "" ? payload.detectorRuleId : ('id' in payload && payload.id != null ? payload.id : "")`),

	entityops.FindingFields.Description.Expr(`paragraphs('description' in payload && payload.description != null ? payload.description : "")`),
	entityops.FindingFields.RecommendedActions.Expr(`'recommendation' in payload && payload.recommendation != null ? payload.recommendation : ""`),

	entityops.FindingFields.Severity.Expr(`'riskLevel' in payload ? payload.riskLevel : ""`),
	entityops.FindingFields.Score.Expr(`'riskScore' in payload && payload.riskScore != null ? payload.riskScore : 0.0`),

	entityops.FindingFields.State.Expr(`'lifecycleDetail' in payload && payload.lifecycleDetail != "" ? payload.lifecycleDetail : ('lifecycleState' in payload ? payload.lifecycleState : "")`),
	entityops.FindingFields.Open.Expr(`'lifecycleDetail' in payload && payload.lifecycleDetail != "" ? payload.lifecycleDetail == "OPEN" : ('lifecycleState' in payload ? payload.lifecycleState == "ACTIVE" : false)`),
	entityops.FindingFields.FindingStatusName.Expr(`'lifecycleDetail' in payload && payload.lifecycleDetail == "DISMISSED" ? "Dismissed" : ('lifecycleDetail' in payload && (payload.lifecycleDetail == "RESOLVED" || payload.lifecycleDetail == "DELETED") ? "Closed" : "Open")`),

	entityops.FindingFields.ResourceName.Expr(`'resourceName' in payload && payload.resourceName != null ? payload.resourceName : ""`),
	entityops.FindingFields.Targets.Expr(`'resourceId' in payload && payload.resourceId != null && payload.resourceId != "" ? [payload.resourceId] : []`),
	entityops.FindingFields.TargetDetails.Expr(`
  'resourceId' in payload && payload.resourceId != null && payload.resourceId != ""
    ? {
        "resourceId": payload.resourceId,
        "resourceType": ('resourceType' in payload && payload.resourceType != null ? payload.resourceType : ""),
        "compartmentId": ('compartmentId' in payload && payload.compartmentId != null ? payload.compartmentId : ""),
        "regions": ('regions' in payload && payload.regions != null ? payload.regions : [])
      }
    : {}
`),

	entityops.FindingFields.ReportedAt.Expr(`'timeFirstDetected' in payload ? payload.timeFirstDetected : null`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`'timeLastDetected' in payload && payload.timeLastDetected != null ? payload.timeLastDetected : ('timeFirstDetected' in payload ? payload.timeFirstDetected : null)`),
	entityops.FindingFields.EventTime.Expr(`'timeLastDetected' in payload ? payload.timeLastDetected : null`),

	entityops.FindingFields.Tags.Expr(`'labels' in payload && payload.labels != null ? payload.labels : []`),
	entityops.FindingFields.RawPayload.Expr(`payload`),
)
