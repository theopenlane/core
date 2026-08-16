package oci

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// mapExprFinding is the CEL mapping expression for OCI Cloud Guard problem payloads mapped to Finding; unset SDK pointer fields marshal as null, hence the != null guards
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyFindingExternalID, Expr: `'id' in payload && payload.id != null ? payload.id : ""`},
	{Key: entityops.InputKeyFindingExternalOwnerID, Expr: `resource`},

	{Key: entityops.InputKeyFindingCategory, Expr: `'detectorRuleId' in payload && payload.detectorRuleId != null ? payload.detectorRuleId : ""`},
	{Key: entityops.InputKeyFindingFindingClass, Expr: `'detectorId' in payload ? payload.detectorId : ""`},
	{Key: entityops.InputKeyFindingDisplayName, Expr: `'detectorRuleId' in payload && payload.detectorRuleId != null && payload.detectorRuleId != "" ? payload.detectorRuleId : ('id' in payload && payload.id != null ? payload.id : "")`},

	{Key: entityops.InputKeyFindingDescription, Expr: `'description' in payload && payload.description != null ? payload.description : ""`},
	{Key: entityops.InputKeyFindingRecommendedActions, Expr: `'recommendation' in payload && payload.recommendation != null ? payload.recommendation : ""`},

	{Key: entityops.InputKeyFindingSeverity, Expr: `'riskLevel' in payload ? payload.riskLevel : ""`},
	{Key: entityops.InputKeyFindingScore, Expr: `'riskScore' in payload && payload.riskScore != null ? payload.riskScore : 0.0`},

	{Key: entityops.InputKeyFindingState, Expr: `'lifecycleDetail' in payload && payload.lifecycleDetail != "" ? payload.lifecycleDetail : ('lifecycleState' in payload ? payload.lifecycleState : "")`},
	{Key: entityops.InputKeyFindingOpen, Expr: `'lifecycleDetail' in payload && payload.lifecycleDetail != "" ? payload.lifecycleDetail == "OPEN" : ('lifecycleState' in payload ? payload.lifecycleState == "ACTIVE" : false)`},
	{Key: entityops.InputKeyFindingFindingStatusName, Expr: `'lifecycleDetail' in payload && payload.lifecycleDetail == "DISMISSED" ? "Dismissed" : ('lifecycleDetail' in payload && (payload.lifecycleDetail == "RESOLVED" || payload.lifecycleDetail == "DELETED") ? "Closed" : "Open")`},

	{Key: entityops.InputKeyFindingResourceName, Expr: `'resourceName' in payload && payload.resourceName != null ? payload.resourceName : ""`},
	{Key: entityops.InputKeyFindingTargets, Expr: `'resourceId' in payload && payload.resourceId != null && payload.resourceId != "" ? [payload.resourceId] : []`},
	{Key: entityops.InputKeyFindingTargetDetails, Expr: `
  'resourceId' in payload && payload.resourceId != null && payload.resourceId != ""
    ? {
        "resourceId": payload.resourceId,
        "resourceType": ('resourceType' in payload && payload.resourceType != null ? payload.resourceType : ""),
        "compartmentId": ('compartmentId' in payload && payload.compartmentId != null ? payload.compartmentId : ""),
        "regions": ('regions' in payload && payload.regions != null ? payload.regions : [])
      }
    : {}
`},

	{Key: entityops.InputKeyFindingReportedAt, Expr: `'timeFirstDetected' in payload ? payload.timeFirstDetected : null`},
	{Key: entityops.InputKeyFindingSourceUpdatedAt, Expr: `'timeLastDetected' in payload && payload.timeLastDetected != null ? payload.timeLastDetected : ('timeFirstDetected' in payload ? payload.timeFirstDetected : null)`},
	{Key: entityops.InputKeyFindingEventTime, Expr: `'timeLastDetected' in payload ? payload.timeLastDetected : null`},

	{Key: entityops.InputKeyFindingTags, Expr: `'labels' in payload && payload.labels != null ? payload.labels : []`},
	{Key: entityops.InputKeyFindingRawPayload, Expr: `payload`},
})
