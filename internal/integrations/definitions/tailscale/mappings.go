package tailscale

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Tailscale user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'loginName' in payload ? payload.loginName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('loginName' in payload ? payload.loginName : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('status' in payload ? (payload.status == "active" ? "ACTIVE" : (payload.status == "suspended" ? "INACTIVE" : "INACTIVE")) : "INACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for Tailscale role group payloads mapped to DirectoryGroup
// Payload is tailscaleGroupPayload — a known struct, so direct field access is safe without 'key' in payload guards
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `payload.id`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `payload.name != "" ? payload.name : payload.id`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Tailscale membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'user_id' in payload ? payload.user_id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `'group_id' in payload ? payload.group_id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// mapExprAsset is the CEL mapping expression for Tailscale device payloads mapped to Asset
var mapExprAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingAssetSourceIdentifier, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingAssetName, Expr: `'name' in payload && payload.name != "" ? payload.name : ('hostname' in payload ? payload.hostname : "")`},
	{Key: integrationgenerated.IntegrationMappingAssetDisplayName, Expr: `'hostname' in payload && payload.hostname != "" ? payload.hostname : ('name' in payload ? payload.name : "")`},
	{Key: integrationgenerated.IntegrationMappingAssetDescription, Expr: `'os' in payload && payload.os != "" ? "OS: " + payload.os : ""`},
	{Key: integrationgenerated.IntegrationMappingAssetTags, Expr: `'tags' in payload && payload.tags != null ? payload.tags : []`},
	{Key: integrationgenerated.IntegrationMappingAssetAssetType, Expr: `"DEVICE"`},
	{Key: integrationgenerated.IntegrationMappingAssetInternalOwner, Expr: `'user' in payload && payload.user != "" ? payload.user : 'creator' in payload && payload.creator != "" ? payload.creator : null`},
})

// tailscaleMappings returns the built-in Tailscale ingest mappings
func tailscaleMappings() []types.MappingRegistration {
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
			Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
			Variant: deviceAssetVariant,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprAsset,
			},
		},
	}
}
