package cloudflare

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Cloudflare account member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'user_id' in payload && payload.user_id != "" ? payload.user_id : ('email' in payload ? payload.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'first_name' in payload && payload.first_name != "" ? (payload.first_name + ('last_name' in payload && payload.last_name != "" ? " " + payload.last_name : "")) : ('email' in payload ? payload.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'first_name' in payload ? payload.first_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'last_name' in payload ? payload.last_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountMfaState, Expr: `dyn('two_factor_enabled' in payload && payload.two_factor_enabled ? "ENABLED" : "DISABLED")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('status' in payload && payload.status == "accepted" ? "ACTIVE" : "INACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDirectoryInstanceID, Expr: `payload.account_id`},
})

// mapExprDirectoryGroup is the CEL mapping expression for Cloudflare groups and roles payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `payload.id`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `payload.name`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: `payload`},
})

// mapExprDirectoryMembership is the CEL mapping expression for Cloudflare policy payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `payload.group_id`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `payload.user_id`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// mapExprFinding is the CEL mapping expression for Cloudflare Security Center insight payloads mapped to Finding
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingFindingExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingExternalOwnerID, Expr: `resource`},
	{Key: integrationgenerated.IntegrationMappingFindingDisplayName, Expr: `'issue_class' in payload ? payload.issue_class : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingResourceName, Expr: `'subject' in payload ? payload.subject : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingTargetDetails, Expr: `'subject' in payload && payload.subject != "" ? {"affected_endpoints": [payload.subject]} : {}`},
	{Key: integrationgenerated.IntegrationMappingFindingTargets, Expr: `'subject' in payload && payload.subject != "" ? [payload.subject] : []`},
	{Key: integrationgenerated.IntegrationMappingFindingCategory, Expr: `'issue_type' in payload ? payload.issue_type : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingSourceUpdatedAt, Expr: `'since' in payload && payload.since != "" ? payload.since : null`},
	{Key: integrationgenerated.IntegrationMappingFindingRecommendedActions, Expr: `'resolve_text' in payload ? payload.resolve_text : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingOpen, Expr: `'status' in payload ? payload.status == "active" : false`},
	{Key: integrationgenerated.IntegrationMappingFindingFindingStatusName, Expr: `'dismissed' in payload && payload.dismissed ? "Dismissed" : ('user_classification' in payload && payload.user_classification == "false_positive" ? "False Positive" : ('status' in payload ? payload.status : ""))`},
	{Key: integrationgenerated.IntegrationMappingFindingState, Expr: `'status' in payload ? payload.status : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingReferences, Expr: `'resolve_link' in payload && payload.resolve_link != "" ? [payload.resolve_link] : []`},
	{Key: integrationgenerated.IntegrationMappingFindingDescription, Expr: `'payload' in payload && 'detection_method' in payload.payload ? payload.payload.detection_method : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingEventTime, Expr: `'timestamp' in payload && payload.timestamp != "" ? payload.timestamp : null`},
	{Key: integrationgenerated.IntegrationMappingFindingSeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingExternalURI, Expr: `'resolve_link' in payload ? payload.resolve_link : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingRawPayload, Expr: "payload"},
})

// mapExprAsset is the CEL mapping expression for Cloudflare Registrar domain payloads mapped to Asset
var mapExprAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingAssetSourceIdentifier, Expr: `'domain_name' in payload ? payload.domain_name : ""`},
	{Key: integrationgenerated.IntegrationMappingAssetDisplayName, Expr: `'domain_name' in payload ? payload.domain_name : ""`},
	{Key: integrationgenerated.IntegrationMappingAssetName, Expr: `'domain_name' in payload ? payload.domain_name : ""`},
	{Key: integrationgenerated.IntegrationMappingAssetAssetType, Expr: `"DOMAIN"`},
	{Key: integrationgenerated.IntegrationMappingAssetObservedAt, Expr: `'created_at' in payload && payload.created_at != "" ? payload.created_at : null`},
})

// cloudflareMappings returns the built-in Cloudflare ingest mappings
func cloudflareMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryGroup,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryMembership,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaFinding,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprFinding,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaAsset,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprAsset,
			},
		},
	}
}
