package cloudflare

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Cloudflare account member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'user_id' in payload && payload.user_id != "" ? payload.user_id : ('email' in payload ? payload.email : "")`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'first_name' in payload && payload.first_name != "" ? (payload.first_name + ('last_name' in payload && payload.last_name != "" ? " " + payload.last_name : "")) : ('email' in payload ? payload.email : "")`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `'first_name' in payload ? payload.first_name : ""`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `'last_name' in payload ? payload.last_name : ""`},
	{Key: entityops.InputKeyDirectoryAccountMfaState, Expr: `dyn('two_factor_enabled' in payload && payload.two_factor_enabled ? "ENABLED" : "DISABLED")`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn('status' in payload && payload.status == "accepted" ? "ACTIVE" : "INACTIVE")`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
	{Key: entityops.InputKeyDirectoryAccountDirectoryInstanceID, Expr: `payload.account_id`},
})

// mapExprDirectoryGroup is the CEL mapping expression for Cloudflare groups and roles payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `payload.id`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `payload.name`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: `payload`},
})

// mapExprDirectoryMembership is the CEL mapping expression for Cloudflare policy payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `payload.group_id`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `payload.user_id`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
})

// mapExprFinding is the CEL mapping expression for Cloudflare Security Center insight payloads mapped to Finding
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyFindingExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyFindingExternalOwnerID, Expr: `resource`},
	{Key: entityops.InputKeyFindingDisplayName, Expr: `'issue_class' in payload ? payload.issue_class : ""`},
	{Key: entityops.InputKeyFindingResourceName, Expr: `'subject' in payload ? payload.subject : ""`},
	{Key: entityops.InputKeyFindingTargetDetails, Expr: `'subject' in payload && payload.subject != "" ? {"affected_endpoints": [payload.subject]} : {}`},
	{Key: entityops.InputKeyFindingTargets, Expr: `'subject' in payload && payload.subject != "" ? [payload.subject] : []`},
	{Key: entityops.InputKeyFindingCategory, Expr: `'issue_type' in payload ? payload.issue_type : ""`},
	{Key: entityops.InputKeyFindingSourceUpdatedAt, Expr: `'since' in payload && payload.since != "" ? payload.since : null`},
	{Key: entityops.InputKeyFindingRecommendedActions, Expr: `'resolve_text' in payload ? payload.resolve_text : ""`},
	{Key: entityops.InputKeyFindingOpen, Expr: `'status' in payload ? payload.status == "active" : false`},
	{Key: entityops.InputKeyFindingFindingStatusName, Expr: `'dismissed' in payload && payload.dismissed ? "Dismissed" : ('user_classification' in payload && payload.user_classification == "false_positive" ? "False Positive" : ('status' in payload ? payload.status : ""))`},
	{Key: entityops.InputKeyFindingState, Expr: `'status' in payload ? payload.status : ""`},
	{Key: entityops.InputKeyFindingReferences, Expr: `'resolve_link' in payload && payload.resolve_link != "" ? [payload.resolve_link] : []`},
	{Key: entityops.InputKeyFindingDescription, Expr: `'payload' in payload && 'detection_method' in payload.payload ? payload.payload.detection_method : ""`},
	{Key: entityops.InputKeyFindingEventTime, Expr: `'timestamp' in payload && payload.timestamp != "" ? payload.timestamp : null`},
	{Key: entityops.InputKeyFindingSeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: entityops.InputKeyFindingExternalURI, Expr: `'resolve_link' in payload ? payload.resolve_link : ""`},
	{Key: entityops.InputKeyFindingRawPayload, Expr: "payload"},
})

// mapExprAsset is the CEL mapping expression for Cloudflare Registrar domain payloads mapped to Asset
var mapExprAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyAssetSourceIdentifier, Expr: `'domain_name' in payload ? payload.domain_name : ""`},
	{Key: entityops.InputKeyAssetDisplayName, Expr: `'domain_name' in payload ? payload.domain_name : ""`},
	{Key: entityops.InputKeyAssetName, Expr: `'domain_name' in payload ? payload.domain_name : ""`},
	{Key: entityops.InputKeyAssetAssetType, Expr: `"DOMAIN"`},
	{Key: entityops.InputKeyAssetObservedAt, Expr: `'created_at' in payload && payload.created_at != "" ? payload.created_at : null`},
})
